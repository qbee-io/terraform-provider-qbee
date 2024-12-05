package provider

import (
	"context"
	"fmt"
	"github.com/hashicorp/terraform-plugin-framework-validators/resourcevalidator"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringdefault"
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
	_ resource.Resource                     = &packageManagementResource{}
	_ resource.ResourceWithConfigure        = &packageManagementResource{}
	_ resource.ResourceWithConfigValidators = &packageManagementResource{}
	_ resource.ResourceWithImportState      = &packageManagementResource{}
)

const (
	errorImportingPackageManagement = "error importing package_management resource"
	errorWritingPackageManagement   = "error writing package_management resource"
	errorReadingPackageManagement   = "error reading package_management resource"
	errorDeletingPackageManagement  = "error deleting package_management resource"
)

// NewPackageManagementResource is a helper function to simplify the provider implementation.
func NewPackageManagementResource() resource.Resource {
	return &packageManagementResource{}
}

type packageManagementResource struct {
	client *client.Client
}

type packageManagementResourceModel struct {
	Node         types.String     `tfsdk:"node"`
	Tag          types.String     `tfsdk:"tag"`
	Extend       types.Bool       `tfsdk:"extend"`
	PreCondition types.String     `tfsdk:"pre_condition"`
	RebootMode   types.String     `tfsdk:"reboot_mode"`
	FullUpgrade  types.Bool       `tfsdk:"full_upgrade"`
	Packages     []managedPackage `tfsdk:"packages"`
}

func (m packageManagementResourceModel) typeAndIdentifier() (config.EntityType, string) {
	return typeAndIdentifier(m.Tag, m.Node)
}

type managedPackage struct {
	Name    types.String `tfsdk:"name"`
	Version types.String `tfsdk:"version"`
}

// Metadata returns the resource type name.
func (r *packageManagementResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_package_management"
}

// Configure adds the provider configured client to the resource.
func (r *packageManagementResource) Configure(_ context.Context, req resource.ConfigureRequest, _ *resource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}

	r.client = req.ProviderData.(*client.Client)
}

// Schema defines the schema for the resource.
func (r *packageManagementResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "PackageManagement controls system packages.",
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
				Description: "If set, will be executed before package maintenance. If the command returns a non-zero exit code, the package maintenance will be skipped.",
			},
			"reboot_mode": schema.StringAttribute{
				Optional:    true,
				Computed:    true,
				Description: "Defines whether the system should be rebooted after package maintenance or not.",
				Default:     stringdefault.StaticString("never"),
				Validators: []validator.String{
					stringvalidator.OneOf("never", "always"),
				},
			},
			"full_upgrade": schema.BoolAttribute{
				Optional:    true,
				Description: "If set to true, will perform a full system upgrade.",
			},
			"packages": schema.ListNestedAttribute{
				Optional:    true,
				Description: "List of packages to be maintained.",
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"name": schema.StringAttribute{
							Required:    true,
							Description: "Name of the package to be maintained.",
						},
						"version": schema.StringAttribute{
							Required:    true,
							Description: "Version of the package to be maintained.",
						},
					},
				},
			},
		},
	}
}

func (r *packageManagementResource) ConfigValidators(ctx context.Context) []resource.ConfigValidator {
	return []resource.ConfigValidator{
		resourcevalidator.ExactlyOneOf(
			path.MatchRoot("tag"),
			path.MatchRoot("node"),
		),
		resourcevalidator.ExactlyOneOf(
			path.MatchRoot("full_upgrade"),
			path.MatchRoot("packages"),
		),
	}
}

// Create creates the resource and sets the initial Terraform state.
func (r *packageManagementResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	// Retrieve values from the plan
	var plan packageManagementResourceModel
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	diags = r.writePackageManagement(ctx, plan)
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
func (r *packageManagementResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	// Retrieve values from the plan
	var plan packageManagementResourceModel
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	diags = r.writePackageManagement(ctx, plan)
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
func (r *packageManagementResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	// Get the current state
	var state *packageManagementResourceModel
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	configType, identifier := state.typeAndIdentifier()

	// Read the real status
	activeConfig, err := r.client.GetActiveConfig(ctx, configType, identifier, config.EntityConfigScopeOwn)
	if err != nil {
		resp.Diagnostics.AddError(errorReadingPackageManagement,
			"error reading the active configuration: "+err.Error())

		return
	}

	// Update the current state
	currentPackageManagement := activeConfig.BundleData.PackageManagement
	if currentPackageManagement == nil {
		resp.State.RemoveResource(ctx)
		return
	}

	state.RebootMode = types.StringValue(string(currentPackageManagement.RebootMode))

	if currentPackageManagement.PreCondition != "" {
		state.PreCondition = types.StringValue(currentPackageManagement.PreCondition)
	}

	if currentPackageManagement.Packages != nil {
		var mappedPackages []managedPackage
		for _, p := range currentPackageManagement.Packages {
			mappedPackages = append(mappedPackages, managedPackage{
				Name:    types.StringValue(p.Name),
				Version: types.StringValue(p.Version),
			})
		}
		state.Packages = mappedPackages
	} else {
		state.FullUpgrade = types.BoolValue(currentPackageManagement.FullUpgrade)
	}

	state.Extend = types.BoolValue(currentPackageManagement.Extend)

	diags = resp.State.Set(ctx, state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}

// Delete deletes the resource and removes the Terraform state on success.
func (r *packageManagementResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	// Retrieve values from the state
	var state packageManagementResourceModel
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Delete the resource
	configType, identifier := state.typeAndIdentifier()
	tflog.Info(ctx, fmt.Sprintf("Deleting package_management for %v %v", configType, identifier))

	content := config.PackageManagement{
		Metadata: config.Metadata{
			Reset:   true,
			Version: "v1",
		},
	}

	changeRequest, err := createChangeRequest(config.PackageManagementBundle, content, configType, identifier)
	if err != nil {
		resp.Diagnostics.AddError(
			errorDeletingPackageManagement,
			err.Error(),
		)
		return
	}

	change, err := r.client.CreateConfigurationChange(ctx, changeRequest)
	if err != nil {
		resp.Diagnostics.AddError(
			errorDeletingPackageManagement,
			err.Error(),
		)
		return
	}

	_, err = r.client.CommitConfiguration(ctx, "terraform: create package_management")
	if err != nil {
		resp.Diagnostics.AddError(errorDeletingPackageManagement,
			"error creating a commit to delete the package_management resource: "+err.Error(),
		)

		err = r.client.DeleteConfigurationChange(ctx, change.SHA)
		if err != nil {
			resp.Diagnostics.AddError(
				errorDeletingPackageManagement,
				"error deleting uncommitted package_management changes: "+err.Error(),
			)
		}

		return
	}
}

// ImportState imports the resource state from the Terraform state.
func (r *packageManagementResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	configType, identifier, found := strings.Cut(req.ID, ":")
	if !found || configType == "" || identifier == "" {
		resp.Diagnostics.AddError(
			errorImportingPackageManagement,
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
			errorImportingPackageManagement,
			fmt.Sprintf("Import type must be either 'node' or 'tag'. Got: %q", configType),
		)
		return
	}
}

func (r *packageManagementResource) writePackageManagement(ctx context.Context, plan packageManagementResourceModel) diag.Diagnostics {
	configType, identifier := plan.typeAndIdentifier()
	extend := plan.Extend.ValueBool()

	// Create the resource
	tflog.Info(ctx, fmt.Sprintf("Creating package_management for %v %v", configType, identifier))

	content := config.PackageManagement{
		Metadata: config.Metadata{
			Enabled: true,
			Extend:  extend,
			Version: "v1",
		},
		RebootMode: config.RebootMode(plan.RebootMode.ValueString()),
	}

	if !plan.PreCondition.IsNull() {
		content.PreCondition = plan.PreCondition.ValueString()
	}

	if plan.Packages != nil {
		for _, p := range plan.Packages {
			content.Packages = append(content.Packages, config.Package{
				Name:    p.Name.ValueString(),
				Version: p.Version.ValueString(),
			})
		}
	} else {
		content.FullUpgrade = plan.FullUpgrade.ValueBool()
	}

	changeRequest, err := createChangeRequest(config.PackageManagementBundle, content, configType, identifier)
	if err != nil {
		return diag.Diagnostics{
			diag.NewErrorDiagnostic(
				errorWritingPackageManagement,
				err.Error(),
			),
		}
	}

	change, err := r.client.CreateConfigurationChange(ctx, changeRequest)
	if err != nil {
		return diag.Diagnostics{
			diag.NewErrorDiagnostic(
				errorWritingPackageManagement,
				fmt.Sprintf("Error creating a package_management resource with qbee: %v", err),
			),
		}
	}

	_, err = r.client.CommitConfiguration(ctx, "terraform: create package_management")
	if err != nil {
		diags := diag.Diagnostics{}

		err = fmt.Errorf("error creating a commit for the package_management: %w", err)
		diags.AddError(errorWritingPackageManagement, err.Error())

		err = r.client.DeleteConfigurationChange(ctx, change.SHA)
		if err != nil {
			diags.AddError(
				errorWritingPackageManagement,
				fmt.Errorf("error deleting uncommitted package_management changes: %w", err).Error(),
			)
		}

		return diags
	}

	return nil
}
