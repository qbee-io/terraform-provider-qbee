package provider

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"go.qbee.io/client/config"

	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
)

// Ensure the implementation satisfies the expected interfaces.
var (
	_ resourceModelManager                  = &settingsResourceModel{}
	_ resource.Resource                     = &settingsResource{}
	_ resource.ResourceWithConfigure        = &settingsResource{}
	_ resource.ResourceWithConfigValidators = &settingsResource{}
	_ resource.ResourceWithImportState      = &settingsResource{}
)

// NewSettingsResource is a helper function to simplify the provider implementation.
func NewSettingsResource() resource.Resource {
	return &settingsResource{
		configurationResource: configurationResource{
			resourceBase: newResourceBase(config.SettingsBundle),
			modelFactory: func() any {
				return new(settingsResourceModel)
			},
		},
	}
}

type settingsResource struct {
	configurationResource
}

// Schema defines the schema for the resource.
func (r *settingsResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Settings defines agent settings.",
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
			"metrics": schema.BoolAttribute{
				Required:    true,
				Description: "Metrics collection enabled.",
			},
			"reports": schema.BoolAttribute{
				Required:    true,
				Description: "Reports collection enabled.",
			},
			"remote_console": schema.BoolAttribute{
				Required:    true,
				Description: "RemoteConsole access enabled.",
			},
			"software_inventory": schema.BoolAttribute{
				Required:    true,
				Description: "SoftwareInventory collection enabled.",
			},
			"process_inventory": schema.BoolAttribute{
				Required:    true,
				Description: "ProcessInventory collection enabled.",
			},
			"agent_interval": schema.Int64Attribute{
				Required:    true,
				Description: "AgentInterval defines how often agent reports back to the device hub (in minutes).",
			},
		},
	}
}

type settingsResourceModel struct {
	configurationResourceModel
	Metrics           types.Bool  `tfsdk:"metrics"`
	Reports           types.Bool  `tfsdk:"reports"`
	RemoteConsole     types.Bool  `tfsdk:"remote_console"`
	SoftwareInventory types.Bool  `tfsdk:"software_inventory"`
	ProcessInventory  types.Bool  `tfsdk:"process_inventory"`
	AgentInterval     types.Int64 `tfsdk:"agent_interval"`
}

func (m settingsResourceModel) getConfigBundle() config.Bundle {
	return config.SettingsBundle
}

func (m *settingsResourceModel) fromBundleData(bundleData config.BundleData) error {
	data := bundleData.Settings
	if data == nil {
		return fmt.Errorf("settings bundle data is nil")
	}

	m.Extend = types.BoolValue(data.Metadata.Extend)

	m.Metrics = types.BoolValue(data.Metrics)
	m.Reports = types.BoolValue(data.Reports)
	m.RemoteConsole = types.BoolValue(data.RemoteConsole)
	m.SoftwareInventory = types.BoolValue(data.SoftwareInventory)
	m.ProcessInventory = types.BoolValue(data.ProcessInventory)
	m.AgentInterval = types.Int64Value(int64(data.AgentInterval))

	return nil
}

func (m settingsResourceModel) toBundleData(metadata config.Metadata) any {
	return config.Settings{
		Metadata:          metadata,
		Metrics:           m.Metrics.ValueBool(),
		Reports:           m.Reports.ValueBool(),
		RemoteConsole:     m.RemoteConsole.ValueBool(),
		SoftwareInventory: m.SoftwareInventory.ValueBool(),
		ProcessInventory:  m.ProcessInventory.ValueBool(),
		AgentInterval:     int(m.AgentInterval.ValueInt64()),
	}
}
