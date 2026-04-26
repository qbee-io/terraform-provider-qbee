package provider

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"go.qbee.io/client/config"
)

// Ensure the implementation satisfies the expected interfaces.
var (
	_ resourceModelManager                  = &raucResourceModel{}
	_ resource.Resource                     = &raucResource{}
	_ resource.ResourceWithConfigure        = &raucResource{}
	_ resource.ResourceWithConfigValidators = &raucResource{}
	_ resource.ResourceWithImportState      = &raucResource{}
)

// NewRaucResource is a helper function to simplify the provider implementation.
func NewRaucResource() resource.Resource {
	return &raucResource{
		configurationResource: configurationResource{
			resourceBase: newResourceBase(config.RaucBundle),
			modelFactory: func() any {
				return new(raucResourceModel)
			},
		},
	}
}

type raucResource struct {
	configurationResource
}

// Schema defines the schema for the resource.
func (r *raucResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Rauc configures an A/B system update using RAUC.",
		Attributes: map[string]schema.Attribute{
			"tag": schema.StringAttribute{
				Optional:      true,
				Description:   "The tag for which to set the configuration. Either tag or node is required.",
				PlanModifiers: []planmodifier.String{stringplanmodifier.RequiresReplace()},
			},
			"node": schema.StringAttribute{
				Optional:      true,
				Description:   "The node for which to set the configuration. Either tag or node is required.",
				PlanModifiers: []planmodifier.String{stringplanmodifier.RequiresReplace()},
			},
			"extend": schema.BoolAttribute{
				Required: true,
				Description: "If the configuration should extend configuration from the parent nodes of the node " +
					"the configuration is applied to. If set to false, configuration from parent nodes is ignored.",
			},
			"pre_condition": schema.StringAttribute{
				Optional:    true,
				Description: "An optional command which needs to return 0 in order for RAUC bundle to be installed.",
			},
			"rauc_bundle": schema.StringAttribute{
				Required:    true,
				Description: "The RAUC bundle to be installed.",
			},
		},
	}
}

type raucResourceModel struct {
	configurationResourceModel
	PreCondition types.String `tfsdk:"pre_condition"`
	RaucBundle   types.String `tfsdk:"rauc_bundle"`
}

func (m raucResourceModel) getConfigBundle() config.Bundle {
	return config.RaucBundle
}

func (m *raucResourceModel) fromBundleData(bundleData config.BundleData) error {
	data := bundleData.Rauc
	if data == nil {
		return fmt.Errorf("rauc bundle data is nil")
	}

	m.Extend = types.BoolValue(data.Metadata.Extend)

	m.RaucBundle = types.StringValue(data.RaucBundle)
	if data.PreCondition != "" {
		m.PreCondition = types.StringValue(data.PreCondition)
	}

	return nil
}

func (m raucResourceModel) toBundleData(metadata config.Metadata) any {
	bundleData := config.Rauc{
		Metadata: metadata,
	}

	if metadata.Reset {
		return bundleData
	}

	bundleData.RaucBundle = m.RaucBundle.ValueString()
	if !m.PreCondition.IsNull() {
		bundleData.PreCondition = m.PreCondition.ValueString()
	}

	return bundleData
}
