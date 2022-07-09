package twilliate

import (
	"context"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/cloudfront"
	"github.com/aws/aws-sdk-go-v2/service/cloudfront/types"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-go/tftypes"
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
	distributionConfig.Origins.Items = append(distributionConfig.Origins.Items, types.Origin{
		DomainName: aws.String(plan.Domain.Value),
		Id:         aws.String(plan.Id.Value),
		CustomHeaders: &types.CustomHeaders{
			Quantity: aws.Int32(0),
		},
		OriginPath: aws.String(""),
		OriginShield: &types.OriginShield{
			Enabled: aws.Bool(false),
		},
		S3OriginConfig: &types.S3OriginConfig{
			OriginAccessIdentity: aws.String("origin-access-identity/cloudfront/" + plan.AccessIdentity.Value),
		},
	})
	*distributionConfig.Origins.Quantity++

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
}

// Delete is called when the provider must delete the resource. Config
// values may be read from the DeleteResourceRequest.
//
// If execution completes without error, the framework will automatically
// call DeleteResourceResponse.State.RemoveResource(), so it can be omitted
// from provider logic.
func (o OriginResource) Delete(ctx context.Context, req tfsdk.DeleteResourceRequest, resp *tfsdk.DeleteResourceResponse) {
}

// Import resource
func (o OriginResource) ImportState(ctx context.Context, req tfsdk.ImportResourceStateRequest, resp *tfsdk.ImportResourceStateResponse) {
	// Save the import identifier in the id attribute
	tfsdk.ResourceImportStatePassthroughID(ctx, tftypes.NewAttributePath().WithAttributeName("id"), req, resp)
}
