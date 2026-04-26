package provider

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"go.qbee.io/client/config"
)

// Ensure the implementation satisfies the expected interfaces.
var (
	_ resourceModelManager                  = &processWatchResourceModel{}
	_ resource.Resource                     = &processWatchResource{}
	_ resource.ResourceWithConfigure        = &processWatchResource{}
	_ resource.ResourceWithConfigValidators = &processWatchResource{}
	_ resource.ResourceWithImportState      = &processWatchResource{}
)

// NewProcessWatchResource is a helper function to simplify the provider implementation.
func NewProcessWatchResource() resource.Resource {
	return &processWatchResource{
		configurationResource: configurationResource{
			// for backwards compatibility, use process_watch instead of config.ProcessWatchBundle
			resourceBase: newResourceBase("process_watch"),
			modelFactory: func() any {
				return new(processWatchResourceModel)
			},
		},
	}
}

type processWatchResource struct {
	configurationResource
}

// Schema defines the schema for the resource.
func (r *processWatchResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "ProcessWatch ensures running process are running (or not).",
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
			"processes": schema.ListNestedAttribute{
				Required:    true,
				Description: "Processes to watch.",
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"name": schema.StringAttribute{
							Required:    true,
							Description: "Name of the process to watch.",
						},
						"policy": schema.StringAttribute{
							Required:    true,
							Description: "Policy for the process.",
							Validators: []validator.String{
								stringvalidator.OneOf("Present", "Absent"),
							},
						},
						"command": schema.StringAttribute{
							Required: true,
							Description: "Command to use to get the process in the expected state. " +
								"For ProcessPresent it should be a start command, " +
								"for ProcessAbsent it should be a stop command.",
						},
					},
				},
			},
		},
	}
}

type processWatchResourceModel struct {
	configurationResourceModel
	Processes []processWatcher `tfsdk:"processes"`
}

type processWatcher struct {
	Name    types.String `tfsdk:"name"`
	Policy  types.String `tfsdk:"policy"`
	Command types.String `tfsdk:"command"`
}

func (m processWatchResourceModel) getConfigBundle() config.Bundle {
	return config.ProcessWatchBundle
}

func (m *processWatchResourceModel) fromBundleData(bundleData config.BundleData) error {
	data := bundleData.ProcWatch
	if data == nil {
		return fmt.Errorf("process_watch bundle data is nil")
	}

	m.Extend = types.BoolValue(data.Metadata.Extend)

	for _, process := range data.Processes {
		m.Processes = append(m.Processes, processWatcher{
			Name:    types.StringValue(process.Name),
			Policy:  types.StringValue(string(process.Policy)),
			Command: types.StringValue(process.Command),
		})
	}

	return nil
}

func (m processWatchResourceModel) toBundleData(metadata config.Metadata) any {
	bundleData := config.ProcessWatch{
		Metadata: metadata,
	}

	if metadata.Reset {
		return bundleData
	}

	for _, process := range m.Processes {
		bundleData.Processes = append(bundleData.Processes, config.ProcessWatcher{
			Name:    process.Name.ValueString(),
			Policy:  config.ProcessPolicy(process.Policy.ValueString()),
			Command: process.Command.ValueString(),
		})
	}

	return bundleData
}
