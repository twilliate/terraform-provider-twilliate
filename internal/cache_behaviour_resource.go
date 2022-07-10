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

type CacheBehaviourResource struct {
	client *cloudfront.Client
}

// Create is called when the provider must create a new resource. Config
// and planned state values should be read from the
// CreateResourceRequest and new state values set on the
// CreateResourceResponse.
func (c CacheBehaviourResource) Create(ctx context.Context, req tfsdk.CreateResourceRequest, resp *tfsdk.CreateResourceResponse) {
	var plan CacheBehaviour
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	out, err := c.client.GetDistributionConfig(ctx, &cloudfront.GetDistributionConfigInput{
		Id: aws.String(plan.DistributionId.Value),
	})

	if err != nil {
		resp.Diagnostics.AddError("failed to get distribution config", err.Error())
		return
	}

	distributionConfig := out.DistributionConfig
	distributionConfig.CacheBehaviors.Items = append(distributionConfig.CacheBehaviors.Items, plan.ToCloudfrontCacheBehaviour())
	*distributionConfig.CacheBehaviors.Quantity++
	// Add new Cache Behaviour to existing configuration

	_, err = c.client.UpdateDistribution(ctx, &cloudfront.UpdateDistributionInput{
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
func (c CacheBehaviourResource) Read(ctx context.Context, req tfsdk.ReadResourceRequest, resp *tfsdk.ReadResourceResponse) {
	var state CacheBehaviour
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
func (c CacheBehaviourResource) Update(ctx context.Context, req tfsdk.UpdateResourceRequest, resp *tfsdk.UpdateResourceResponse) {
	// current state
	var state CacheBehaviour
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if diags.HasError() {
		return
	}

	// planned state
	var plan CacheBehaviour
	diags = req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if diags.HasError() {
		return
	}

	if state.DistributionId.Value != plan.DistributionId.Value {
		err := c.deleteFromDistribution(ctx, state)
		if err != nil {
			resp.Diagnostics.AddError("failed to remove cache behaviour from previous distribution", err.Error())
		}
	}

	out, err := c.client.GetDistributionConfig(ctx, &cloudfront.GetDistributionConfigInput{
		Id: aws.String(plan.DistributionId.Value),
	})

	if err != nil {
		resp.Diagnostics.AddError("failed to get distribution config", err.Error())
		return
	}

	distributionConfig := out.DistributionConfig

	idx := slices.IndexFunc(distributionConfig.CacheBehaviors.Items, func(behaviour types.CacheBehavior) bool {
		return *behaviour.TargetOriginId == state.OriginId.Value && *behaviour.PathPattern == state.PathPattern.Value
	})

	if idx == -1 {
		distributionConfig.CacheBehaviors.Items = append(distributionConfig.CacheBehaviors.Items, plan.ToCloudfrontCacheBehaviour())
		*distributionConfig.CacheBehaviors.Quantity++
	} else {
		distributionConfig.CacheBehaviors.Items[idx] = plan.ToCloudfrontCacheBehaviour()
	}

	_, err = c.client.UpdateDistribution(ctx, &cloudfront.UpdateDistributionInput{
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
func (c CacheBehaviourResource) Delete(ctx context.Context, req tfsdk.DeleteResourceRequest, resp *tfsdk.DeleteResourceResponse) {
	var state CacheBehaviour
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if diags.HasError() {
		return
	}

	err := c.deleteFromDistribution(ctx, state)

	// Its okay if the behaviour has already been deleted
	if err != nil {
		resp.Diagnostics.AddWarning("failed to delete cache behaviour from distribution", err.Error())
	}

	resp.State.RemoveResource(ctx)
}

func (c CacheBehaviourResource) deleteFromDistribution(ctx context.Context, state CacheBehaviour) error {
	out, err := c.client.GetDistributionConfig(ctx, &cloudfront.GetDistributionConfigInput{
		Id: aws.String(state.DistributionId.Value),
	})

	if err != nil {
		return err
	}

	idx := slices.IndexFunc(out.DistributionConfig.CacheBehaviors.Items, func(behaviour types.CacheBehavior) bool {
		return *behaviour.TargetOriginId == state.OriginId.Value && *behaviour.PathPattern == state.PathPattern.Value
	})

	if idx == -1 {
		return fmt.Errorf("the cache behaviour with origin id %s and path %s can not be found, it has been modified or removed", state.DistributionId, state.PathPattern)
	}

	out.DistributionConfig.CacheBehaviors.Items = append(out.DistributionConfig.CacheBehaviors.Items[:idx], out.DistributionConfig.CacheBehaviors.Items[idx+1:]...)
	*out.DistributionConfig.CacheBehaviors.Quantity--

	_, err = c.client.UpdateDistribution(ctx, &cloudfront.UpdateDistributionInput{
		DistributionConfig: out.DistributionConfig,
		Id:                 aws.String(state.DistributionId.Value),
		IfMatch:            out.ETag,
	})

	return err
}
