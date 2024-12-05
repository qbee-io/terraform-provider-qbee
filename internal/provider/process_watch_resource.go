package provider

import (
	"context"
	"fmt"
	"github.com/hashicorp/terraform-plugin-framework-validators/resourcevalidator"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	"go.qbee.io/client"
	"go.qbee.io/client/config"
	"strings"

	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
)

// Ensure the implementation satisfies the expected interfaces.
var (
	_ resource.Resource                     = &ProcessWatchResource{}
	_ resource.ResourceWithConfigure        = &ProcessWatchResource{}
	_ resource.ResourceWithConfigValidators = &ProcessWatchResource{}
	_ resource.ResourceWithImportState      = &ProcessWatchResource{}
)

const (
	errorImportingProcessWatch = "error importing process_watch resource"
	errorWritingProcessWatch   = "error writing process_watch resource"
	errorReadingProcessWatch   = "error reading process_watch resource"
	errorDeletingProcessWatch  = "error deleting process_watch resource"
)

// NewProcessWatchResource is a helper function to simplify the provider implementation.
func NewProcessWatchResource() resource.Resource {
	return &ProcessWatchResource{}
}

type ProcessWatchResource struct {
	client *client.Client
}

type ProcessWatchResourceModel struct {
	Node      types.String     `tfsdk:"node"`
	Tag       types.String     `tfsdk:"tag"`
	Extend    types.Bool       `tfsdk:"extend"`
	Processes []processWatcher `tfsdk:"processes"`
}

func (m ProcessWatchResourceModel) typeAndIdentifier() (config.EntityType, string) {
	return typeAndIdentifier(m.Tag, m.Node)
}

type processWatcher struct {
	Name    types.String `tfsdk:"name"`
	Policy  types.String `tfsdk:"policy"`
	Command types.String `tfsdk:"command"`
}

// Metadata returns the resource type name.
func (r *ProcessWatchResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_process_watch"
}

// Configure adds the provider configured client to the resource.
func (r *ProcessWatchResource) Configure(_ context.Context, req resource.ConfigureRequest, _ *resource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}

	r.client = req.ProviderData.(*client.Client)
}

// Schema defines the schema for the resource.
func (r *ProcessWatchResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
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

func (r *ProcessWatchResource) ConfigValidators(ctx context.Context) []resource.ConfigValidator {
	return []resource.ConfigValidator{
		resourcevalidator.ExactlyOneOf(
			path.MatchRoot("tag"),
			path.MatchRoot("node"),
		),
	}
}

// Create creates the resource and sets the initial Terraform state.
func (r *ProcessWatchResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	// Retrieve values from the plan
	var plan ProcessWatchResourceModel
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	diags = r.writeProcessWatch(ctx, plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Map response body to schema and populate Computed attribute values
	// Set state to fully populated data
	diags = resp.State.Set(ctx, plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}

// Update updates the resource and sets the updated Terraform state on success.
func (r *ProcessWatchResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	// Retrieve values from the plan
	var plan ProcessWatchResourceModel
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	diags = r.writeProcessWatch(ctx, plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Map response body to schema and populate Computed attribute values
	// Set state to fully populated data
	diags = resp.State.Set(ctx, plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}

// Read refreshes the Terraform state with the latest data.
func (r *ProcessWatchResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	// Get the current state
	var state *ProcessWatchResourceModel
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	configType, identifier := state.typeAndIdentifier()

	// Read the real status
	activeConfig, err := r.client.GetActiveConfig(ctx, configType, identifier, config.EntityConfigScopeOwn)
	if err != nil {
		resp.Diagnostics.AddError(errorReadingProcessWatch,
			"error reading the active configuration: "+err.Error())

		return
	}

	// Update the current state
	currentProcessWatch := activeConfig.BundleData.ProcWatch
	if currentProcessWatch == nil {
		resp.State.RemoveResource(ctx)
		return
	}

	state.Extend = types.BoolValue(currentProcessWatch.Extend)

	var mappedProcesses []processWatcher
	for _, process := range currentProcessWatch.Processes {
		mappedProcesses = append(mappedProcesses, processWatcher{
			Name:    types.StringValue(process.Name),
			Policy:  types.StringValue(string(process.Policy)),
			Command: types.StringValue(process.Command),
		})
	}
	state.Processes = mappedProcesses

	diags = resp.State.Set(ctx, state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}

// Delete deletes the resource and removes the Terraform state on success.
func (r *ProcessWatchResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	// Retrieve values from the state
	var state ProcessWatchResourceModel
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Delete the resource
	configType, identifier := state.typeAndIdentifier()
	tflog.Info(ctx, fmt.Sprintf("Deleting process_watch for %v %v", configType, identifier))

	content := config.ProcessWatch{
		Metadata: config.Metadata{
			Reset:   true,
			Version: "v1",
		},
	}

	changeRequest, err := createChangeRequest(config.ProcessWatchBundle, content, configType, identifier)
	if err != nil {
		resp.Diagnostics.AddError(
			errorDeletingProcessWatch,
			err.Error(),
		)
		return
	}

	change, err := r.client.CreateConfigurationChange(ctx, changeRequest)
	if err != nil {
		resp.Diagnostics.AddError(
			errorDeletingProcessWatch,
			err.Error(),
		)
		return
	}

	_, err = r.client.CommitConfiguration(ctx, "terraform: create process_watch")
	if err != nil {
		resp.Diagnostics.AddError(errorDeletingProcessWatch,
			"error creating a commit to delete the process_watch resource: "+err.Error(),
		)

		err = r.client.DeleteConfigurationChange(ctx, change.SHA)
		if err != nil {
			resp.Diagnostics.AddError(
				errorDeletingProcessWatch,
				"error deleting uncommitted process_watch changes: "+err.Error(),
			)
		}

		return
	}
}

// ImportState imports the resource state from the Terraform state.
func (r *ProcessWatchResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	configType, identifier, found := strings.Cut(req.ID, ":")
	if !found || configType == "" || identifier == "" {
		resp.Diagnostics.AddError(
			errorImportingProcessWatch,
			fmt.Sprintf("Expected import identifier with format: type:identifier. Got: %q", req.ID),
		)
		return

	}

	if configType == "tag" {
		resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("tag"), identifier)...)
	} else if configType == "node" {
		resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("node"), identifier)...)
	} else {
		resp.Diagnostics.AddError(
			errorImportingProcessWatch,
			fmt.Sprintf("Import type must be either 'node' or 'tag'. Got: %q", configType),
		)
		return
	}
}

func (r *ProcessWatchResource) writeProcessWatch(ctx context.Context, plan ProcessWatchResourceModel) diag.Diagnostics {
	configType, identifier := plan.typeAndIdentifier()
	extend := plan.Extend.ValueBool()

	// Create the resource
	tflog.Info(ctx, fmt.Sprintf("Creating process_watch for %v %v", configType, identifier))

	var mappedProcesses []config.ProcessWatcher
	for _, process := range plan.Processes {
		mappedProcesses = append(mappedProcesses, config.ProcessWatcher{
			Name:    process.Name.ValueString(),
			Policy:  config.ProcessPolicy(process.Policy.ValueString()),
			Command: process.Command.ValueString(),
		})
	}

	content := config.ProcessWatch{
		Metadata: config.Metadata{
			Enabled: true,
			Extend:  extend,
			Version: "v1",
		},
		Processes: mappedProcesses,
	}

	changeRequest, err := createChangeRequest(config.ProcessWatchBundle, content, configType, identifier)
	if err != nil {
		return diag.Diagnostics{
			diag.NewErrorDiagnostic(
				errorWritingProcessWatch,
				err.Error(),
			),
		}
	}

	change, err := r.client.CreateConfigurationChange(ctx, changeRequest)
	if err != nil {
		return diag.Diagnostics{
			diag.NewErrorDiagnostic(
				errorWritingProcessWatch,
				fmt.Sprintf("Error creating a process_watch resource with qbee: %v", err),
			),
		}
	}

	_, err = r.client.CommitConfiguration(ctx, "terraform: create process_watch")
	if err != nil {
		diags := diag.Diagnostics{}

		err = fmt.Errorf("error creating a commit for the process_watch: %w", err)
		diags.AddError(errorWritingProcessWatch, err.Error())

		err = r.client.DeleteConfigurationChange(ctx, change.SHA)
		if err != nil {
			diags.AddError(
				errorWritingProcessWatch,
				fmt.Errorf("error deleting uncommitted process_watch changes: %w", err).Error(),
			)
		}

		return diags
	}

	return nil
}
