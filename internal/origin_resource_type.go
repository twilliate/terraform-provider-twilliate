package internal

import (
	"context"
	"github.com/aws/aws-sdk-go-v2/aws"
	cloudfrontTypes "github.com/aws/aws-sdk-go-v2/service/cloudfront/types"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

type Origin struct {
	DistributionId types.String `tfsdk:"distribution_id"`
	Id             types.String `tfsdk:"origin_id"`
	Domain         types.String `tfsdk:"origin_domain"`
	AccessIdentity types.String `tfsdk:"origin_access_identity"`
}

//type CacheBehaviour struct {
//	ViewerProtocolPolicy types.String `tfsdk:"viewer_protocol_policy"`
//	CachePolicyId        types.String `tfsdk:"cache_policy_id"`
//	PathPattern          types.String `tfsdk:"path_pattern"`
//}

type OriginResourceType struct{}

func (o OriginResourceType) GetSchema(_ context.Context) (tfsdk.Schema, diag.Diagnostics) {
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
			"origin_domain": {
				Type:     types.StringType,
				Required: true,
			},
			"origin_access_identity": {
				Type:     types.StringType,
				Required: true,
			},
		},
	}, nil
}

func (o OriginResourceType) NewResource(_ context.Context, p tfsdk.Provider) (tfsdk.Resource, diag.Diagnostics) {
	return OriginResource{
		client: p.(*provider).client,
	}, nil
}

func OriginFromResource(origin Origin) cloudfrontTypes.Origin {
	return cloudfrontTypes.Origin{
		DomainName: aws.String(origin.Domain.Value),
		Id:         aws.String(origin.Id.Value),
		CustomHeaders: &cloudfrontTypes.CustomHeaders{
			Quantity: aws.Int32(0),
		},
		OriginPath: aws.String(""),
		OriginShield: &cloudfrontTypes.OriginShield{
			Enabled: aws.Bool(false),
		},
		S3OriginConfig: &cloudfrontTypes.S3OriginConfig{
			OriginAccessIdentity: aws.String("origin-access-identity/cloudfront/" + origin.AccessIdentity.Value),
		},
	}
}
