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
	DistributionId             types.String                `tfsdk:"distribution_id"`
	OriginId                   types.String                `tfsdk:"origin_id"`
	ViewerProtocolPolicy       types.String                `tfsdk:"viewer_protocol_policy"`
	PathPattern                types.String                `tfsdk:"path_pattern"`
	CachePolicyId              types.String                `tfsdk:"cache_policy_id"`
	AllowedMethods             *AllowedMethods             `tfsdk:"allowed_methods"`
	Compress                   types.Bool                  `tfsdk:"compress"`
	FieldLevelEncryptionId     types.String                `tfsdk:"field_level_encryption_id"`
	FunctionAssociations       []FunctionAssociation       `tfsdk:"function_associations"`
	LambdaFunctionAssociations []LambdaFunctionAssociation `tfsdk:"lambda_function_associations"`
	OriginRequestPolicyId      types.String                `tfsdk:"origin_request_policy_id"`
	RealtimeLogConfigArn       types.String                `tfsdk:"realtime_log_config_arn"`
	ResponseHeadersPolicyId    types.String                `tfsdk:"response_headers_policy_id"`
	SmoothStreaming            types.Bool                  `tfsdk:"smooth_streaming"`
	TrustedKeyGroups           *TrustedKeyGroups           `tfsdk:"trusted_key_groups"`
	TrustedSigners             *TrustedSigners             `tfsdk:"trusted_signers"`
}

type AllowedMethods struct {
	Items         []types.String `tfsdk:"allowed_methods"`
	CachedMethods []types.String `tfsdk:"cached_methods"`
}

type FunctionAssociation struct {
	EventType types.String `tfsdk:"event_type"`
	Arn       types.String `tfsdk:"function_arn"`
}

type LambdaFunctionAssociation struct {
	EventType   types.String `tfsdk:"event_type"`
	Arn         types.String `tfsdk:"function_arn"`
	IncludeBody types.Bool   `tfsdk:"include_body"`
}

type TrustedKeyGroups struct {
	Enabled types.Bool     `tfsdk:"enabled"`
	Groups  []types.String `tfsdk:"groups"`
}

type TrustedSigners struct {
	Enabled types.Bool     `tfsdk:"enabled"`
	Signers []types.String `tfsdk:"signers"`
}

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
			"allowed_methods": {
				Optional: true,
				Attributes: tfsdk.SingleNestedAttributes(map[string]tfsdk.Attribute{
					"allowed_methods": {
						Type:     types.ListType{ElemType: types.StringType},
						Optional: true,
					},
					"cached_methods": {
						Type:     types.ListType{ElemType: types.StringType},
						Optional: true,
					},
				}),
			},
			"compress": {
				Optional: true,
				Type:     types.BoolType,
			},
			"field_level_encryption_id": {
				Type:     types.StringType,
				Optional: true,
			},
			"function_associations": {
				Optional: true,
				Attributes: tfsdk.ListNestedAttributes(map[string]tfsdk.Attribute{
					"event_type": {
						Type:     types.StringType,
						Required: true,
					},
					"function_arn": {
						Type:     types.StringType,
						Required: true,
					},
				}),
			},
			"lambda_function_associations": {
				Optional: true,
				Attributes: tfsdk.ListNestedAttributes(map[string]tfsdk.Attribute{
					"event_type": {
						Type:     types.StringType,
						Required: true,
					},
					"function_arn": {
						Type:     types.StringType,
						Required: true,
					},
					"include_body": {
						Type:     types.BoolType,
						Optional: true,
					},
				}),
			},
			"origin_request_policy_id": {
				Type:     types.StringType,
				Optional: true,
			},
			"realtime_log_config_arn": {
				Type:     types.StringType,
				Optional: true,
			},
			"response_headers_policy_id": {
				Type:     types.StringType,
				Optional: true,
			},
			"smooth_streaming": {
				Type:     types.BoolType,
				Optional: true,
			},
			"trusted_key_groups": {
				Optional: true,
				Attributes: tfsdk.SingleNestedAttributes(map[string]tfsdk.Attribute{
					"enabled": {
						Type:     types.BoolType,
						Optional: true,
					},
					"groups": {
						Type:     types.ListType{ElemType: types.StringType},
						Required: true,
					},
				}),
			},
			"trusted_signers": {
				Optional: true,
				Attributes: tfsdk.SingleNestedAttributes(map[string]tfsdk.Attribute{
					"enabled": {
						Type:     types.BoolType,
						Optional: true,
					},
					"signers": {
						Type:     types.ListType{ElemType: types.StringType},
						Required: true,
					},
				}),
			},
		},
	}, nil
}

func (o CacheBehaviourResourceType) NewResource(_ context.Context, p tfsdk.Provider) (tfsdk.Resource, diag.Diagnostics) {
	return CacheBehaviourResource{
		client: p.(*provider).client,
	}, nil
}

func (c CacheBehaviour) ToCloudfrontAllowedMethods() *cloudfrontTypes.AllowedMethods {
	if c.AllowedMethods == nil || len(c.AllowedMethods.Items) == 0 {
		return &cloudfrontTypes.AllowedMethods{
			Items:    []cloudfrontTypes.Method{"HEAD", "GET", "OPTIONS"},
			Quantity: aws.Int32(3),
			CachedMethods: &cloudfrontTypes.CachedMethods{
				Items:    []cloudfrontTypes.Method{"HEAD", "GET"},
				Quantity: aws.Int32(2),
			},
		}
	}

	var items []cloudfrontTypes.Method
	for _, item := range c.AllowedMethods.Items {
		items = append(items, cloudfrontTypes.Method(item.Value))
	}

	if len(c.AllowedMethods.CachedMethods) == 0 {
		return &cloudfrontTypes.AllowedMethods{
			Items:    items,
			Quantity: aws.Int32(int32(len(c.AllowedMethods.Items))),
			CachedMethods: &cloudfrontTypes.CachedMethods{
				Quantity: aws.Int32(0),
			},
		}
	}

	var cachedMethods []cloudfrontTypes.Method
	for _, method := range c.AllowedMethods.CachedMethods {
		cachedMethods = append(cachedMethods, cloudfrontTypes.Method(method.Value))
	}

	return &cloudfrontTypes.AllowedMethods{
		Items:    items,
		Quantity: aws.Int32(int32(len(c.AllowedMethods.Items))),
		CachedMethods: &cloudfrontTypes.CachedMethods{
			Items:    cachedMethods,
			Quantity: aws.Int32(int32(len(c.AllowedMethods.CachedMethods))),
		},
	}
}

func (c CacheBehaviour) ToFunctionAssociation() *cloudfrontTypes.FunctionAssociations {
	if c.FunctionAssociations == nil || len(c.FunctionAssociations) == 0 {
		return &cloudfrontTypes.FunctionAssociations{
			Quantity: aws.Int32(0),
		}
	}

	var items []cloudfrontTypes.FunctionAssociation
	for _, function := range c.FunctionAssociations {
		items = append(items, cloudfrontTypes.FunctionAssociation{
			EventType:   cloudfrontTypes.EventType(*toString(function.EventType)),
			FunctionARN: toString(function.Arn),
		})
	}

	return &cloudfrontTypes.FunctionAssociations{
		Quantity: aws.Int32(int32(len(c.FunctionAssociations))),
		Items:    items,
	}
}

func (c CacheBehaviour) ToLambdaFunctionAssociation() *cloudfrontTypes.LambdaFunctionAssociations {
	if c.LambdaFunctionAssociations == nil || len(c.LambdaFunctionAssociations) == 0 {
		return &cloudfrontTypes.LambdaFunctionAssociations{
			Quantity: aws.Int32(0),
		}
	}

	var items []cloudfrontTypes.LambdaFunctionAssociation
	for _, function := range c.LambdaFunctionAssociations {
		items = append(items, cloudfrontTypes.LambdaFunctionAssociation{
			EventType:         cloudfrontTypes.EventType(*toString(function.EventType)),
			LambdaFunctionARN: toString(function.Arn),
			IncludeBody:       toBool(function.IncludeBody, false),
		})
	}

	return &cloudfrontTypes.LambdaFunctionAssociations{
		Quantity: aws.Int32(int32(len(c.LambdaFunctionAssociations))),
		Items:    items,
	}
}

func (c CacheBehaviour) ToTrustedKeyGroups() *cloudfrontTypes.TrustedKeyGroups {
	if c.TrustedKeyGroups == nil || len(c.TrustedKeyGroups.Groups) == 0 {
		return &cloudfrontTypes.TrustedKeyGroups{
			Enabled:  aws.Bool(false),
			Quantity: aws.Int32(0),
		}
	}

	var items []string
	for _, group := range c.TrustedKeyGroups.Groups {
		items = append(items, group.Value)
	}

	return &cloudfrontTypes.TrustedKeyGroups{
		Enabled:  toBool(c.TrustedKeyGroups.Enabled, false),
		Quantity: aws.Int32(int32(len(c.TrustedKeyGroups.Groups))),
		Items:    items,
	}
}

func (c CacheBehaviour) ToTrustedSigners() *cloudfrontTypes.TrustedSigners {
	if c.TrustedSigners == nil || len(c.TrustedSigners.Signers) == 0 {
		return &cloudfrontTypes.TrustedSigners{
			Enabled:  aws.Bool(false),
			Quantity: aws.Int32(0),
		}
	}

	var items []string
	for _, signer := range c.TrustedSigners.Signers {
		items = append(items, signer.Value)
	}

	return &cloudfrontTypes.TrustedSigners{
		Enabled:  toBool(c.TrustedSigners.Enabled, false),
		Quantity: aws.Int32(int32(len(c.TrustedSigners.Signers))),
		Items:    items,
	}
}

func (c CacheBehaviour) ToCloudfrontCacheBehaviour() cloudfrontTypes.CacheBehavior {
	return cloudfrontTypes.CacheBehavior{
		PathPattern:                aws.String(c.PathPattern.Value),
		TargetOriginId:             aws.String(c.OriginId.Value),
		ViewerProtocolPolicy:       cloudfrontTypes.ViewerProtocolPolicy(c.ViewerProtocolPolicy.Value),
		AllowedMethods:             c.ToCloudfrontAllowedMethods(),
		CachePolicyId:              aws.String(c.CachePolicyId.Value),
		Compress:                   toBool(c.Compress, true),
		FieldLevelEncryptionId:     toString(c.FieldLevelEncryptionId),
		FunctionAssociations:       c.ToFunctionAssociation(),
		LambdaFunctionAssociations: c.ToLambdaFunctionAssociation(),
		OriginRequestPolicyId:      toStringOrNil(c.OriginRequestPolicyId),
		RealtimeLogConfigArn:       toStringOrNil(c.RealtimeLogConfigArn),
		ResponseHeadersPolicyId:    toStringOrNil(c.ResponseHeadersPolicyId),
		SmoothStreaming:            toBool(c.SmoothStreaming, false),
		TrustedKeyGroups:           c.ToTrustedKeyGroups(),
		TrustedSigners:             c.ToTrustedSigners(),
	}
}
