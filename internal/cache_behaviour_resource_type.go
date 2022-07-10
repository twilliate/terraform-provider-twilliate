package internal

import (
	"context"
	"github.com/aws/aws-sdk-go-v2/aws"
	cloudfrontTypes "github.com/aws/aws-sdk-go-v2/service/cloudfront/types"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

type CacheBehaviour struct {
	DistributionId       types.String `tfsdk:"distribution_id"`
	OriginId             types.String `tfsdk:"origin_id"`
	ViewerProtocolPolicy types.String `tfsdk:"viewer_protocol_policy"`
	PathPattern          types.String `tfsdk:"path_pattern"`
	CachePolicyId        types.String `tfsdk:"cache_policy_id"`
}

//type CacheBehaviour struct {
//	ViewerProtocolPolicy types.String `tfsdk:"viewer_protocol_policy"`
//	CachePolicyId        types.String `tfsdk:"cache_policy_id"`
//	PathPattern          types.String `tfsdk:"path_pattern"`
//}

type CacheBehaviourResourceType struct{}

func (o CacheBehaviourResourceType) GetSchema(_ context.Context) (tfsdk.Schema, diag.Diagnostics) {
	return tfsdk.Schema{
		Attributes: map[string]tfsdk.Attribute{
			"distribution_id": {
				Type:     types.StringType,
				Required: true,
			},
			"origin_id": {
				Type:     types.StringType,
				Required: true,
			},
			"viewer_protocol_policy": {
				Type:     types.StringType,
				Required: true,
			},
			"path_pattern": {
				Type:     types.StringType,
				Required: true,
			},
			"cache_policy_id": {
				Type:     types.StringType,
				Required: true,
			},
		},
	}, nil
}

func (o CacheBehaviourResourceType) NewResource(_ context.Context, p tfsdk.Provider) (tfsdk.Resource, diag.Diagnostics) {
	return CacheBehaviourResource{
		client: p.(*provider).client,
	}, nil
}

func (c CacheBehaviour) ToCloudfrontCacheBehaviour() cloudfrontTypes.CacheBehavior {
	return cloudfrontTypes.CacheBehavior{
		TargetOriginId:       aws.String(c.OriginId.Value),
		ViewerProtocolPolicy: cloudfrontTypes.ViewerProtocolPolicy(c.ViewerProtocolPolicy.Value),
		PathPattern:          aws.String(c.PathPattern.Value),
		CachePolicyId:        aws.String(c.CachePolicyId.Value),
		// TODO: Convert these hardcoded values to default attributes
		FieldLevelEncryptionId: aws.String(""),
		TrustedSigners: &cloudfrontTypes.TrustedSigners{
			Enabled:  aws.Bool(false),
			Quantity: aws.Int32(0),
		},
		TrustedKeyGroups: &cloudfrontTypes.TrustedKeyGroups{
			Enabled:  aws.Bool(false),
			Quantity: aws.Int32(0),
		},
		AllowedMethods: &cloudfrontTypes.AllowedMethods{
			Items:    []cloudfrontTypes.Method{"HEAD", "GET", "OPTIONS"},
			Quantity: aws.Int32(3),
			CachedMethods: &cloudfrontTypes.CachedMethods{
				Items:    []cloudfrontTypes.Method{"HEAD", "GET"},
				Quantity: aws.Int32(2),
			},
		},
		SmoothStreaming: aws.Bool(false),
		Compress:        aws.Bool(true),
		LambdaFunctionAssociations: &cloudfrontTypes.LambdaFunctionAssociations{
			Quantity: aws.Int32(0),
		},
		FunctionAssociations: &cloudfrontTypes.FunctionAssociations{
			Quantity: aws.Int32(0),
		},
	}
}
