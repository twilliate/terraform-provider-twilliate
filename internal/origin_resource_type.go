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
	DistributionId        types.String        `tfsdk:"distribution_id"`
	Id                    types.String        `tfsdk:"origin_id"`
	Domain                types.String        `tfsdk:"origin_domain"`
	OriginPath            types.String        `tfsdk:"origin_path"`
	CustomHeaders         []CustomHeader      `tfsdk:"custom_headers"`
	S3OriginConfig        *S3OriginConfig     `tfsdk:"s3_origin_config"`
	CustomOriginConfig    *CustomOriginConfig `tfsdk:"custom_origin_config"`
	ConnectionAttempts    types.Int64         `tfsdk:"connection_attempts"`
	ConnectionTimeout     types.Int64         `tfsdk:"connection_timeout"`
	OriginShield          *OriginShield       `tfsdk:"origin_shield"`
	OriginAccessControlId types.String        `tfsdk:"origin_access_control_id"`
}

type OriginShield struct {
	Enabled            types.Bool   `tfsdk:"enabled"`
	OriginShieldRegion types.String `tfsdk:"origin_shield_region"`
}

type CustomHeader struct {
	HeaderName  types.String `tfsdk:"name"`
	HeaderValue types.String `tfsdk:"value"`
}

type S3OriginConfig struct {
	OriginAccessIdentity types.String `tfsdk:"origin_access_identity"`
}

type CustomOriginConfig struct {
	HTTPPort               types.Int64    `tfsdk:"http_port"`
	HTTPSPort              types.Int64    `tfsdk:"https_port"`
	OriginProtocolPolicy   types.String   `tfsdk:"origin_protocol_policy"`
	OriginSslProtocols     []types.String `tfsdk:"origin_ssl_protocols"`
	OriginReadTimeout      types.Int64    `tfsdk:"origin_read_timeout"`
	OriginKeepaliveTimeout types.Int64    `tfsdk:"origin_keep_alive_timeout"`
}

type OriginResourceType struct{}

func (o Origin) getCloudfrontCustomHeaders() *cloudfrontTypes.CustomHeaders {
	if o.CustomHeaders == nil || len(o.CustomHeaders) == 0 {
		return &cloudfrontTypes.CustomHeaders{
			Quantity: aws.Int32(0),
		}
	}

	var items []cloudfrontTypes.OriginCustomHeader
	for _, header := range o.CustomHeaders {
		items = append(items, cloudfrontTypes.OriginCustomHeader{
			HeaderName:  aws.String(header.HeaderName.Value),
			HeaderValue: aws.String(header.HeaderValue.Value),
		})
	}

	return &cloudfrontTypes.CustomHeaders{
		Quantity: aws.Int32(int32(len(o.CustomHeaders))),
		Items:    items,
	}
}

func (o Origin) getCustomOriginConfig() *cloudfrontTypes.CustomOriginConfig {
	if o.CustomOriginConfig == nil {
		return nil
	}

	var items []cloudfrontTypes.SslProtocol
	for _, sslProtocol := range o.CustomOriginConfig.OriginSslProtocols {
		items = append(items, cloudfrontTypes.SslProtocol(sslProtocol.Value))
	}

	var originSslProtocols *cloudfrontTypes.OriginSslProtocols

	if len(o.CustomOriginConfig.OriginSslProtocols) > 0 {
		originSslProtocols = &cloudfrontTypes.OriginSslProtocols{
			Items:    items,
			Quantity: aws.Int32(int32(len(o.CustomOriginConfig.OriginSslProtocols))),
		}
	}

	return &cloudfrontTypes.CustomOriginConfig{
		HTTPPort:               toInt32(o.CustomOriginConfig.HTTPPort),
		HTTPSPort:              toInt32(o.CustomOriginConfig.HTTPSPort),
		OriginProtocolPolicy:   cloudfrontTypes.OriginProtocolPolicy(*toString(o.CustomOriginConfig.OriginProtocolPolicy)),
		OriginKeepaliveTimeout: toInt32(o.CustomOriginConfig.OriginKeepaliveTimeout),
		OriginReadTimeout:      toInt32(o.CustomOriginConfig.OriginReadTimeout),
		OriginSslProtocols:     originSslProtocols,
	}
}

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
			"origin_path": {
				Type:     types.StringType,
				Optional: true,
			},
			"custom_headers": {
				Optional: true,
				Attributes: tfsdk.ListNestedAttributes(map[string]tfsdk.Attribute{
					"name": {
						Type:     types.StringType,
						Required: true,
					},
					"value": {
						Type:     types.StringType,
						Required: true,
					},
				}),
			},
			"s3_origin_config": {
				Optional: true,
				Attributes: tfsdk.SingleNestedAttributes(map[string]tfsdk.Attribute{
					"origin_access_identity": {
						Type:     types.StringType,
						Required: true,
					},
				}),
			},
			"custom_origin_config": {
				Optional: true,
				Attributes: tfsdk.SingleNestedAttributes(map[string]tfsdk.Attribute{
					"http_port": {
						Type:     types.Int64Type,
						Required: true,
					},
					"https_port": {
						Type:     types.Int64Type,
						Required: true,
					},
					"origin_protocol_policy": {
						Type:     types.StringType,
						Required: true,
					},
					"origin_ssl_protocols": {
						Type:     types.ListType{ElemType: types.StringType},
						Optional: true,
					},
					"origin_read_timeout": {
						Type:     types.Int64Type,
						Optional: true,
					},
					"origin_keep_alive_timeout": {
						Type:     types.Int64Type,
						Optional: true,
					},
				}),
			},
			"connection_attempts": {
				Type:     types.Int64Type,
				Optional: true,
			},
			"connection_timeout": {
				Type:     types.Int64Type,
				Optional: true,
			},
			"origin_shield": {
				Optional: true,
				Attributes: tfsdk.SingleNestedAttributes(map[string]tfsdk.Attribute{
					"enabled": {
						Type:     types.BoolType,
						Required: true,
					},
					"origin_shield_region": {
						Type:     types.StringType,
						Optional: true,
					},
				}),
			},
			"origin_access_control_id": {
				Type:     types.StringType,
				Optional: true,
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
	originShield := &cloudfrontTypes.OriginShield{
		Enabled: aws.Bool(false),
	}

	if origin.OriginShield != nil {
		originShield = &cloudfrontTypes.OriginShield{
			Enabled:            toBool(origin.OriginShield.Enabled, false),
			OriginShieldRegion: toString(origin.OriginShield.OriginShieldRegion),
		}
	}

	var s3OriginConfig *cloudfrontTypes.S3OriginConfig
	if origin.S3OriginConfig != nil {
		s3OriginConfig = &cloudfrontTypes.S3OriginConfig{
			OriginAccessIdentity: aws.String("origin-access-identity/cloudfront/" + origin.S3OriginConfig.OriginAccessIdentity.Value),
		}
	}

	return cloudfrontTypes.Origin{
		DomainName:         aws.String(origin.Domain.Value),
		Id:                 aws.String(origin.Id.Value),
		ConnectionAttempts: toInt32(origin.ConnectionAttempts),
		ConnectionTimeout:  toInt32(origin.ConnectionTimeout),
		CustomHeaders:      origin.getCloudfrontCustomHeaders(),
		CustomOriginConfig: origin.getCustomOriginConfig(),
		OriginPath:         toString(origin.OriginPath),
		OriginShield:       originShield,
		S3OriginConfig:     s3OriginConfig,
	}
}
