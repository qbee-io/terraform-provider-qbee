package provider

import (
	"context"
	"fmt"
	"github.com/hashicorp/terraform-plugin-framework-validators/resourcevalidator"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
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
	_ resource.Resource                     = &templateResource{}
	_ resource.ResourceWithConfigure        = &templateResource{}
	_ resource.ResourceWithConfigValidators = &templateResource{}
	_ resource.ResourceWithImportState      = &templateResource{}
)

const (
	errorImportingTemplate = "error importing template resource"
	errorWritingTemplate   = "error writing template resource"
	errorReadingTemplate   = "error reading template resource"
	errorDeletingTemplate  = "error deleting template resource"
)

// NewTemplateResource is a helper function to simplify the provider implementation.
func NewTemplate() resource.Resource {
	return &templateResource{}
}

type templateResource struct {
	client *client.Client
}

type templateResourceModel struct {
	Node   types.String `tfsdk:"node"`
	Tag    types.String `tfsdk:"tag"`
	ID     types.String `tfsdk:"id"`
	Extend types.Bool   `tfsdk:"extend"`
}

func (m templateResourceModel) typeAndIdentifier() (config.EntityType, string) {
	return typeAndIdentifier(m.Tag, m.Node)
}

// Metadata returns the resource type name.
func (r *templateResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_connectivity_watchdog"
}

// Configure adds the provider configured client to the resource.
func (r *templateResource) Configure(_ context.Context, req resource.ConfigureRequest, _ *resource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}

	r.client = req.ProviderData.(*client.Client)
}

// Schema defines the schema for the resource.
func (r *templateResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed:    true,
				Description: "Placeholder ID value",
			},
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
		},
	}
}

func (r *templateResource) ConfigValidators(ctx context.Context) []resource.ConfigValidator {
	return []resource.ConfigValidator{
		resourcevalidator.ExactlyOneOf(
			path.MatchRoot("tag"),
			path.MatchRoot("node"),
		),
	}
}

// Create creates the resource and sets the initial Terraform state.
func (r *templateResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	// Retrieve values from the plan
	var plan templateResourceModel
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	diags = r.writeTemplate(ctx, plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Map response body to schema and populate Computed attribute values
	plan.ID = types.StringValue("placeholder")

	// Set state to fully populated data
	diags = resp.State.Set(ctx, plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}

// Update updates the resource and sets the updated Terraform state on success.
func (r *templateResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	// Retrieve values from the plan
	var plan templateResourceModel
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	diags = r.writeTemplate(ctx, plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Map response body to schema and populate Computed attribute values
	plan.ID = types.StringValue("placeholder")

	// Set state to fully populated data
	diags = resp.State.Set(ctx, plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}

// Read refreshes the Terraform state with the latest data.
func (r *templateResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	// Get the current state
	var state *templateResourceModel
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	configType, identifier := state.typeAndIdentifier()

	// Read the real status
	activeConfig, err := r.client.GetActiveConfig(ctx, configType, identifier, config.EntityConfigScopeOwn)
	if err != nil {
		resp.Diagnostics.AddError(errorReadingTemplate,
			"error reading the active configuration: "+err.Error())

		return
	}

	// Update the current state
	// TODO: Actually perform mapping to state
	fmt.Printf("Active config to be mapped: %v\n", activeConfig)
	//currentTemplate := activeConfig.BundleData.Template
	//if currentTemplate == nil {
	//	resp.State.RemoveResource(ctx)
	//	return
	//}

	//state.ID = types.StringValue("placeholder")
	//state.Extend = types.BoolValue(currentTemplate.Extend)
	// state.Property = mappedProperty

	diags = resp.State.Set(ctx, state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}

// Delete deletes the resource and removes the Terraform state on success.
func (r *templateResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	// Retrieve values from the state
	var state templateResourceModel
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Delete the resource
	configType, identifier := state.typeAndIdentifier()
	tflog.Info(ctx, fmt.Sprintf("Deleting template for %v %v", configType, identifier))

	//// TODO: Create correct content
	//content := config.Template{
	//	Metadata: config.Metadata{
	//		Reset:   true,
	//		Version: "v1",
	//	},
	//}
	//
	//changeRequest, err := createChangeRequest(config.TemplateBundle, content, configType, identifier)
	//if err != nil {
	//	resp.Diagnostics.AddError(
	//		errorDeletingTemplate,
	//		err.Error(),
	//	)
	//	return
	//}
	//
	//change, err := r.client.CreateConfigurationChange(ctx, changeRequest)
	//if err != nil {
	//	resp.Diagnostics.AddError(
	//		errorDeletingTemplate,
	//		err.Error(),
	//	)
	//	return
	//}
	//
	//_, err = r.client.CommitConfiguration(ctx, "terraform: create template_resource")
	//if err != nil {
	//	resp.Diagnostics.AddError(errorDeletingTemplate,
	//		"error creating a commit to delete the template resource: "+err.Error(),
	//	)
	//
	//	err = r.client.DeleteConfigurationChange(ctx, change.SHA)
	//	if err != nil {
	//		resp.Diagnostics.AddError(
	//			errorDeletingTemplate,
	//			"error deleting uncommitted template changes: "+err.Error(),
	//		)
	//	}
	//
	//	return
	//}
}

// ImportState imports the resource state from the Terraform state.
func (r *templateResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	configType, identifier, found := strings.Cut(req.ID, ":")
	if !found || configType == "" || identifier == "" {
		resp.Diagnostics.AddError(
			errorImportingTemplate,
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
			errorImportingTemplate,
			fmt.Sprintf("Import type must be either 'node' or 'tag'. Got: %q", configType),
		)
		return
	}
}

func (r *templateResource) writeTemplate(ctx context.Context, plan templateResourceModel) diag.Diagnostics {
	// TODO: Implement resource creation logic
	return nil
}
