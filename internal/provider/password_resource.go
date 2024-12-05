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
	_ resource.Resource                     = &passwordResource{}
	_ resource.ResourceWithConfigure        = &passwordResource{}
	_ resource.ResourceWithConfigValidators = &passwordResource{}
	_ resource.ResourceWithImportState      = &passwordResource{}
)

const (
	errorImportingPassword = "error importing password resource"
	errorWritingPassword   = "error writing password resource"
	errorReadingPassword   = "error reading password resource"
	errorDeletingPassword  = "error deleting password resource"
)

// NewPasswordResource is a helper function to simplify the provider implementation.
func NewPasswordResource() resource.Resource {
	return &passwordResource{}
}

type passwordResource struct {
	client *client.Client
}

type passwordResourceModel struct {
	Node   types.String   `tfsdk:"node"`
	Tag    types.String   `tfsdk:"tag"`
	Extend types.Bool     `tfsdk:"extend"`
	Users  []userPassword `tfsdk:"users"`
}

func (m passwordResourceModel) typeAndIdentifier() (config.EntityType, string) {
	return typeAndIdentifier(m.Tag, m.Node)
}

type userPassword struct {
	Username     types.String `tfsdk:"username"`
	PasswordHash types.String `tfsdk:"password_hash"`
}

// Metadata returns the resource type name.
func (r *passwordResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_password"
}

// Configure adds the provider configured client to the resource.
func (r *passwordResource) Configure(_ context.Context, req resource.ConfigureRequest, _ *resource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}

	r.client = req.ProviderData.(*client.Client)
}

// Schema defines the schema for the resource.
func (r *passwordResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Password bundle sets passwords for existing users.",
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
			"users": schema.ListNestedAttribute{
				Required:    true,
				Description: "A list of users and their password hashes.",
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"username": schema.StringAttribute{
							Required:    true,
							Description: "The username of the user for which the password hash is set.",
						},
						"password_hash": schema.StringAttribute{
							Required:    true,
							Description: "The password hash for the user. See https://qbee.io/docs/qbee-password.html for more information.",
						},
					},
				},
			},
		},
	}
}

func (r *passwordResource) ConfigValidators(ctx context.Context) []resource.ConfigValidator {
	return []resource.ConfigValidator{
		resourcevalidator.ExactlyOneOf(
			path.MatchRoot("tag"),
			path.MatchRoot("node"),
		),
	}
}

// Create creates the resource and sets the initial Terraform state.
func (r *passwordResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	// Retrieve values from the plan
	var plan passwordResourceModel
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	diags = r.writePassword(ctx, plan)
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
func (r *passwordResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	// Retrieve values from the plan
	var plan passwordResourceModel
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	diags = r.writePassword(ctx, plan)
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
func (r *passwordResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	// Get the current state
	var state *passwordResourceModel
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	configType, identifier := state.typeAndIdentifier()

	// Read the real status
	activeConfig, err := r.client.GetActiveConfig(ctx, configType, identifier, config.EntityConfigScopeOwn)
	if err != nil {
		resp.Diagnostics.AddError(errorReadingPassword,
			"error reading the active configuration: "+err.Error())

		return
	}

	// Update the current state
	currentPassword := activeConfig.BundleData.Password
	if currentPassword == nil {
		resp.State.RemoveResource(ctx)
		return
	}

	var mappedUsers []userPassword
	for _, user := range currentPassword.Users {
		mappedUsers = append(mappedUsers, userPassword{
			Username:     types.StringValue(user.Username),
			PasswordHash: types.StringValue(user.PasswordHash),
		})
	}
	state.Users = mappedUsers

	state.Extend = types.BoolValue(currentPassword.Extend)

	diags = resp.State.Set(ctx, state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}

// Delete deletes the resource and removes the Terraform state on success.
func (r *passwordResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	// Retrieve values from the state
	var state passwordResourceModel
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Delete the resource
	configType, identifier := state.typeAndIdentifier()
	tflog.Info(ctx, fmt.Sprintf("Deleting password for %v %v", configType, identifier))

	content := config.Password{
		Metadata: config.Metadata{
			Reset:   true,
			Version: "v1",
		},
	}

	changeRequest, err := createChangeRequest(config.PasswordBundle, content, configType, identifier)
	if err != nil {
		resp.Diagnostics.AddError(
			errorDeletingPassword,
			err.Error(),
		)
		return
	}

	change, err := r.client.CreateConfigurationChange(ctx, changeRequest)
	if err != nil {
		resp.Diagnostics.AddError(
			errorDeletingPassword,
			err.Error(),
		)
		return
	}

	_, err = r.client.CommitConfiguration(ctx, "terraform: create password")
	if err != nil {
		resp.Diagnostics.AddError(errorDeletingPassword,
			"error creating a commit to delete the password resource: "+err.Error(),
		)

		err = r.client.DeleteConfigurationChange(ctx, change.SHA)
		if err != nil {
			resp.Diagnostics.AddError(
				errorDeletingPassword,
				"error deleting uncommitted password changes: "+err.Error(),
			)
		}

		return
	}
}

// ImportState imports the resource state from the Terraform state.
func (r *passwordResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	configType, identifier, found := strings.Cut(req.ID, ":")
	if !found || configType == "" || identifier == "" {
		resp.Diagnostics.AddError(
			errorImportingPassword,
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
			errorImportingPassword,
			fmt.Sprintf("Import type must be either 'node' or 'tag'. Got: %q", configType),
		)
		return
	}
}

func (r *passwordResource) writePassword(ctx context.Context, plan passwordResourceModel) diag.Diagnostics {
	configType, identifier := plan.typeAndIdentifier()
	extend := plan.Extend.ValueBool()

	// Create the resource
	tflog.Info(ctx, fmt.Sprintf("Creating password for %v %v", configType, identifier))

	var mappedUsers []config.UserPassword
	for _, user := range plan.Users {
		mappedUsers = append(mappedUsers, config.UserPassword{
			Username:     user.Username.ValueString(),
			PasswordHash: user.PasswordHash.ValueString(),
		})
	}

	content := config.Password{
		Metadata: config.Metadata{
			Enabled: true,
			Extend:  extend,
			Version: "v1",
		},
		Users: mappedUsers,
	}

	changeRequest, err := createChangeRequest(config.PasswordBundle, content, configType, identifier)
	if err != nil {
		return diag.Diagnostics{
			diag.NewErrorDiagnostic(
				errorWritingPassword,
				err.Error(),
			),
		}
	}

	change, err := r.client.CreateConfigurationChange(ctx, changeRequest)
	if err != nil {
		return diag.Diagnostics{
			diag.NewErrorDiagnostic(
				errorWritingPassword,
				fmt.Sprintf("Error creating a password resource with qbee: %v", err),
			),
		}
	}

	_, err = r.client.CommitConfiguration(ctx, "terraform: create password")
	if err != nil {
		diags := diag.Diagnostics{}

		err = fmt.Errorf("error creating a commit for the password: %w", err)
		diags.AddError(errorWritingPassword, err.Error())

		err = r.client.DeleteConfigurationChange(ctx, change.SHA)
		if err != nil {
			diags.AddError(
				errorWritingPassword,
				fmt.Errorf("error deleting uncommitted password changes: %w", err).Error(),
			)
		}

		return diags
	}

	return nil
}
