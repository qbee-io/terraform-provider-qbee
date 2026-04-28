package provider

import (
	"context"
	"fmt"
	"strconv"

	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"go.qbee.io/client/config"

	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
)

// Ensure the implementation satisfies the expected interfaces.
var (
	_ resourceModelManager                  = &connectivityWatchdogResourceModel{}
	_ resource.Resource                     = &connectivityWatchdogResource{}
	_ resource.ResourceWithConfigure        = &connectivityWatchdogResource{}
	_ resource.ResourceWithConfigValidators = &connectivityWatchdogResource{}
	_ resource.ResourceWithImportState      = &connectivityWatchdogResource{}
)

// NewConnectivityWatchdogResource is a helper function to simplify the provider implementation.
func NewConnectivityWatchdogResource() resource.Resource {
	return &connectivityWatchdogResource{
		configurationResource: configurationResource{
			resourceBase: newResourceBase(config.ConnectivityWatchdogBundle),
			modelFactory: func() any {
				return new(connectivityWatchdogResourceModel)
			},
		},
	}
}

type connectivityWatchdogResource struct {
	configurationResource
}

// Schema defines the schema for the resource.
func (r *connectivityWatchdogResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "When enabled, will count failed connection attempts to the device hub and reboot the device if the threshold is reached.",
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
			"threshold": schema.Int64Attribute{
				Required:    true,
				Description: "defines how many consecutive failed pings are allowed before the watchdog triggers a reboot.",
			},
		},
	}
}

type connectivityWatchdogResourceModel struct {
	configurationResourceModel
	Threshold types.Int64 `tfsdk:"threshold"`
}

func (m connectivityWatchdogResourceModel) getConfigBundle() config.Bundle {
	return config.ConnectivityWatchdogBundle
}

func (m *connectivityWatchdogResourceModel) fromBundleData(bundleData config.BundleData) error {
	data := bundleData.ConnectivityWatchdog
	if data == nil {
		return fmt.Errorf("connectivity_watchdog bundle data is nil")
	}

	m.Extend = types.BoolValue(data.Metadata.Extend)

	threshold, err := strconv.ParseInt(data.Threshold, 10, 64)
	if err != nil {
		return fmt.Errorf("error parsing the threshold value to an integer value: %w", err)
	}

	m.Threshold = types.Int64Value(threshold)

	return nil
}

func (m connectivityWatchdogResourceModel) toBundleData(metadata config.Metadata) any {
	bundleData := config.ConnectivityWatchdog{
		Metadata: metadata,
	}

	if metadata.Reset {
		return bundleData
	}

	bundleData.Threshold = strconv.FormatInt(m.Threshold.ValueInt64(), 10)

	return bundleData
}
