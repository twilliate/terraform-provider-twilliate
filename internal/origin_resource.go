package internal

import (
	"context"
	"fmt"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/cloudfront"
	"github.com/aws/aws-sdk-go-v2/service/cloudfront/types"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"golang.org/x/exp/slices"
)

type OriginResource struct {
	client *cloudfront.Client
}

// Create is called when the provider must create a new resource. Config
// and planned state values should be read from the
// CreateResourceRequest and new state values set on the
// CreateResourceResponse.
func (o OriginResource) Create(ctx context.Context, req tfsdk.CreateResourceRequest, resp *tfsdk.CreateResourceResponse) {
	var plan Origin
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	out, err := o.client.GetDistributionConfig(ctx, &cloudfront.GetDistributionConfigInput{
		Id: aws.String(plan.DistributionId.Value),
	})

	if err != nil {
		resp.Diagnostics.AddError("failed to get distribution config", err.Error())
		return
	}

	distributionConfig := out.DistributionConfig
	// Add new Origin to existing configuration
	distributionConfig.Origins.Items = append(distributionConfig.Origins.Items, OriginFromResource(plan))
	*distributionConfig.Origins.Quantity++

	_, err = o.client.UpdateDistribution(ctx, &cloudfront.UpdateDistributionInput{
		DistributionConfig: distributionConfig,
		Id:                 aws.String(plan.DistributionId.Value),
		IfMatch:            out.ETag,
	})

	if err != nil {
		resp.Diagnostics.AddError("failed to create origin in distribution", err.Error())
		return
	}

	resp.State.Set(ctx, &plan)
	resp.Diagnostics.Append(diags...)

	if resp.Diagnostics.HasError() {
		return
	}
}

// Read is called when the provider must read resource values in order
// to update state. Planned state values should be read from the
// ReadResourceRequest and new state values set on the
// ReadResourceResponse.
func (o OriginResource) Read(ctx context.Context, req tfsdk.ReadResourceRequest, resp *tfsdk.ReadResourceResponse) {
	var state Origin
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if diags.HasError() {
		return
	}

	diags = resp.State.Set(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}

// Update is called to update the state of the resource. Config, planned
// state, and prior state values should be read from the
// UpdateResourceRequest and new state values set on the
// UpdateResourceResponse.
func (o OriginResource) Update(ctx context.Context, req tfsdk.UpdateResourceRequest, resp *tfsdk.UpdateResourceResponse) {
	// current state
	var state Origin
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if diags.HasError() {
		return
	}

	// planned state
	var plan Origin
	diags = req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if diags.HasError() {
		return
	}

	// distribution changed, remove origin from old distribution
	if state.DistributionId.Value != plan.DistributionId.Value {
		err := o.deleteFromDistribution(ctx, state)
		if err != nil {
			resp.Diagnostics.AddError("failed to remove origin from previous distribution", err.Error())
		}
	}

	out, err := o.client.GetDistributionConfig(ctx, &cloudfront.GetDistributionConfigInput{
		Id: aws.String(plan.DistributionId.Value),
	})

	if err != nil {
		resp.Diagnostics.AddError("failed to get distribution config", err.Error())
		return
	}

	distributionConfig := out.DistributionConfig

	idx := slices.IndexFunc(distributionConfig.Origins.Items, func(origin types.Origin) bool {
		return *origin.Id == state.Id.Value
	})

	if idx == -1 {
		distributionConfig.Origins.Items = append(distributionConfig.Origins.Items, OriginFromResource(plan))
		*distributionConfig.Origins.Quantity++
	} else {
		distributionConfig.Origins.Items[idx] = OriginFromResource(plan)
	}

	_, err = o.client.UpdateDistribution(ctx, &cloudfront.UpdateDistributionInput{
		DistributionConfig: distributionConfig,
		Id:                 aws.String(plan.DistributionId.Value),
		IfMatch:            out.ETag,
	})

	if err != nil {
		resp.Diagnostics.AddError("failed to update distribution", err.Error())
		return
	}

	resp.State.Set(ctx, &plan)
	resp.Diagnostics.Append(diags...)

	if resp.Diagnostics.HasError() {
		return
	}
}

// Delete is called when the provider must delete the resource. Config
// values may be read from the DeleteResourceRequest.
//
// If execution completes without error, the framework will automatically
// call DeleteResourceResponse.State.RemoveResource(), so it can be omitted
// from provider logic.
func (o OriginResource) Delete(ctx context.Context, req tfsdk.DeleteResourceRequest, resp *tfsdk.DeleteResourceResponse) {
	var state Origin
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if diags.HasError() {
		return
	}

	err := o.deleteFromDistribution(ctx, state)

	if err != nil {
		resp.Diagnostics.AddError("failed to delete origin from distribution", err.Error())
		return
	}

	resp.State.RemoveResource(ctx)
}

func (o OriginResource) deleteFromDistribution(ctx context.Context, origin Origin) error {
	out, err := o.client.GetDistributionConfig(ctx, &cloudfront.GetDistributionConfigInput{
		Id: aws.String(origin.DistributionId.Value),
	})

	if err != nil {
		return err
	}

	idx := slices.IndexFunc(out.DistributionConfig.Origins.Items, func(o types.Origin) bool {
		return *o.Id == origin.Id.Value
	})

	if idx == -1 {
		return fmt.Errorf("the origin with id %s can not be found, it has been modified or removed", origin.Id)
	}

	out.DistributionConfig.Origins.Items = append(out.DistributionConfig.Origins.Items[:idx], out.DistributionConfig.Origins.Items[idx+1:]...)
	*out.DistributionConfig.Origins.Quantity--

	// remove cache behaviour associated with this origin, otherwise we can not delete the origin
	for i, item := range out.DistributionConfig.CacheBehaviors.Items {
		if *item.TargetOriginId == origin.Id.Value {
			out.DistributionConfig.CacheBehaviors.Items = append(out.DistributionConfig.CacheBehaviors.Items[:i], out.DistributionConfig.CacheBehaviors.Items[i+1:]...)
			*out.DistributionConfig.CacheBehaviors.Quantity--
		}
	}

	_, err = o.client.UpdateDistribution(ctx, &cloudfront.UpdateDistributionInput{
		DistributionConfig: out.DistributionConfig,
		Id:                 aws.String(origin.DistributionId.Value),
		IfMatch:            out.ETag,
	})

	return err
}

// Import resource
//func (o OriginResource) ImportState(ctx context.Context, req tfsdk.ImportResourceStateRequest, resp *tfsdk.ImportResourceStateResponse) {
//	// Save the import identifier in the id attribute
//	tfsdk.ResourceImportStatePassthroughID(ctx, tftypes.NewAttributePath().WithAttributeName("distribution_id"), req, resp)
//}
