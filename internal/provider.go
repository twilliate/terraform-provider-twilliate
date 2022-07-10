package internal

import (
	"context"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/cloudfront"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

func New() tfsdk.Provider {
	return &provider{}
}

type provider struct {
	configured bool
	client     *cloudfront.Client
}

// GetSchema
func (p *provider) GetSchema(_ context.Context) (tfsdk.Schema, diag.Diagnostics) {
	return tfsdk.Schema{
		Attributes: map[string]tfsdk.Attribute{
			"region": {
				Type:     types.StringType,
				Optional: true,
			},
		},
	}, nil
}

// Provider schema struct
type providerData struct {
	Region types.String `tfsdk:"region"`
}

func (p *provider) Configure(ctx context.Context, req tfsdk.ConfigureProviderRequest, resp *tfsdk.ConfigureProviderResponse) {
	// Retrieve provider data from configuration
	var providerConfig providerData
	diags := req.Config.Get(ctx, &providerConfig)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	cfg, err := config.LoadDefaultConfig(context.Background())
	if !providerConfig.Region.IsNull() {
		cfg.Region = providerConfig.Region.Value
	}
	if err != nil {
		resp.Diagnostics.AddError("Unable to create aws client", "Cannot get default aws config")
		return
	}
	client := cloudfront.NewFromConfig(cfg)

	p.configured = true
	p.client = client
}

// GetResources - Defines provider resources
func (p *provider) GetResources(_ context.Context) (map[string]tfsdk.ResourceType, diag.Diagnostics) {
	return map[string]tfsdk.ResourceType{
		"twilliate_cloudfront_origin": OriginResourceType{},
	}, nil
}

// GetDataSources - Defines provider data sources
func (p *provider) GetDataSources(_ context.Context) (map[string]tfsdk.DataSourceType, diag.Diagnostics) {
	return map[string]tfsdk.DataSourceType{}, nil
}
