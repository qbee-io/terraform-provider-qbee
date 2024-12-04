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
	_ resource.Resource                     = &resourceTemplateResource{}
	_ resource.ResourceWithConfigure        = &resourceTemplateResource{}
	_ resource.ResourceWithConfigValidators = &resourceTemplateResource{}
	_ resource.ResourceWithImportState      = &resourceTemplateResource{}
)

// TODO REMOVE
const resourceTemplateBundle config.Bundle = "resource_template"

const (
	errorImportingResourceTemplate = "error importing resource_template resource"
	errorWritingResourceTemplate   = "error writing resource_template resource"
	errorReadingResourceTemplate   = "error reading resource_template resource"
	errorDeletingResourceTemplate  = "error deleting resource_template resource"
)

// NewResourceTemplateResource is a helper function to simplify the provider implementation.
func NewResourceTemplateResource() resource.Resource {
	return &resourceTemplateResource{}
}

type resourceTemplateResource struct {
	client *client.Client
}

type resourceTemplateResourceModel struct {
	Node   types.String `tfsdk:"node"`
	Tag    types.String `tfsdk:"tag"`
	Extend types.Bool   `tfsdk:"extend"`
}

func (m resourceTemplateResourceModel) typeAndIdentifier() (config.EntityType, string) {
	return typeAndIdentifier(m.Tag, m.Node)
}

// Metadata returns the resource type name.
func (r *resourceTemplateResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_resource_template"
}

// Configure adds the provider configured client to the resource.
func (r *resourceTemplateResource) Configure(_ context.Context, req resource.ConfigureRequest, _ *resource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}

	r.client = req.ProviderData.(*client.Client)
}

// Schema defines the schema for the resource.
func (r *resourceTemplateResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
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
		},
	}
}

func (r *resourceTemplateResource) ConfigValidators(ctx context.Context) []resource.ConfigValidator {
	return []resource.ConfigValidator{
		resourcevalidator.ExactlyOneOf(
			path.MatchRoot("tag"),
			path.MatchRoot("node"),
		),
	}
}

// Create creates the resource and sets the initial Terraform state.
func (r *resourceTemplateResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	// Retrieve values from the plan
	var plan resourceTemplateResourceModel
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	diags = r.writeResourceTemplate(ctx, plan)
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
func (r *resourceTemplateResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	// Retrieve values from the plan
	var plan resourceTemplateResourceModel
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	diags = r.writeResourceTemplate(ctx, plan)
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
func (r *resourceTemplateResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	// Get the current state
	var state *resourceTemplateResourceModel
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	configType, identifier := state.typeAndIdentifier()

	// Read the real status
	activeConfig, err := r.client.GetActiveConfig(ctx, configType, identifier, config.EntityConfigScopeOwn)
	if err != nil {
		resp.Diagnostics.AddError(errorReadingResourceTemplate,
			"error reading the active configuration: "+err.Error())

		return
	}

	// Update the current state
	// TODO: Actually perform mapping to state
	fmt.Printf("Active config to be mapped: %v\n", activeConfig)
	//currentResourceTemplate := activeConfig.BundleData.ResourceTemplate
	//if currentResourceTemplate == nil {
	//	resp.State.RemoveResource(ctx)
	//	return
	//}

	//state.Extend = types.BoolValue(currentResourceTemplate.Extend)
	// state.Property = mappedProperty

	diags = resp.State.Set(ctx, state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}

// Delete deletes the resource and removes the Terraform state on success.
func (r *resourceTemplateResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	// Retrieve values from the state
	var state resourceTemplateResourceModel
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Delete the resource
	configType, identifier := state.typeAndIdentifier()
	tflog.Info(ctx, fmt.Sprintf("Deleting resource_template for %v %v", configType, identifier))

	//// TODO: Create correct content
	//content := config.ResourceTemplate{
	//	Metadata: config.Metadata{
	//		Reset:   true,
	//		Version: "v1",
	//	},
	//}
	//
	//changeRequest, err := createChangeRequest(config.ResourceTemplateBundle, content, configType, identifier)
	//if err != nil {
	//	resp.Diagnostics.AddError(
	//		errorDeletingResourceTemplate,
	//		err.Error(),
	//	)
	//	return
	//}
	//
	//change, err := r.client.CreateConfigurationChange(ctx, changeRequest)
	//if err != nil {
	//	resp.Diagnostics.AddError(
	//		errorDeletingResourceTemplate,
	//		err.Error(),
	//	)
	//	return
	//}
	//
	//_, err = r.client.CommitConfiguration(ctx, "terraform: create resource_template")
	//if err != nil {
	//	resp.Diagnostics.AddError(errorDeletingResourceTemplate,
	//		"error creating a commit to delete the resource_template resource: "+err.Error(),
	//	)
	//
	//	err = r.client.DeleteConfigurationChange(ctx, change.SHA)
	//	if err != nil {
	//		resp.Diagnostics.AddError(
	//			errorDeletingResourceTemplate,
	//			"error deleting uncommitted resource_template changes: "+err.Error(),
	//		)
	//	}
	//
	//	return
	//}
}

// ImportState imports the resource state from the Terraform state.
func (r *resourceTemplateResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	configType, identifier, found := strings.Cut(req.ID, ":")
	if !found || configType == "" || identifier == "" {
		resp.Diagnostics.AddError(
			errorImportingResourceTemplate,
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
			errorImportingResourceTemplate,
			fmt.Sprintf("Import type must be either 'node' or 'tag'. Got: %q", configType),
		)
		return
	}
}

func (r *resourceTemplateResource) writeResourceTemplate(ctx context.Context, plan resourceTemplateResourceModel) diag.Diagnostics {
	configType, identifier := plan.typeAndIdentifier()
	extend := plan.Extend.ValueBool()

	// Create the resource
	tflog.Info(ctx, fmt.Sprintf("Creating resource_template for %v %v", configType, identifier))

	//content := config.ResourceTemplate{
	content := struct{ Metadata config.Metadata }{
		Metadata: config.Metadata{
			Enabled: true,
			Extend:  extend,
			Version: "v1",
		},
		// TODO: Add the actual content
	}

	//changeRequest, err := createChangeRequest(config.ResourceTemplateBundle, content, configType, identifier)
	changeRequest, err := createChangeRequest(resourceTemplateBundle, content, configType, identifier)
	if err != nil {
		return diag.Diagnostics{
			diag.NewErrorDiagnostic(
				errorWritingResourceTemplate,
				err.Error(),
			),
		}
	}

	change, err := r.client.CreateConfigurationChange(ctx, changeRequest)
	if err != nil {
		return diag.Diagnostics{
			diag.NewErrorDiagnostic(
				errorWritingResourceTemplate,
				fmt.Sprintf("Error creating a resource_template resource with qbee: %v", err),
			),
		}
	}

	_, err = r.client.CommitConfiguration(ctx, "terraform: create resource_template")
	if err != nil {
		diags := diag.Diagnostics{}

		err = fmt.Errorf("error creating a commit for the resource_template: %w", err)
		diags.AddError(errorWritingResourceTemplate, err.Error())

		err = r.client.DeleteConfigurationChange(ctx, change.SHA)
		if err != nil {
			diags.AddError(
				errorWritingResourceTemplate,
				fmt.Errorf("error deleting uncommitted resource_template changes: %w", err).Error(),
			)
		}

		return diags
	}

	return nil
}
