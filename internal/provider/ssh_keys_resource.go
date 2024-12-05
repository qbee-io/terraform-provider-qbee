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
	_ resource.Resource                     = &sshKeysResource{}
	_ resource.ResourceWithConfigure        = &sshKeysResource{}
	_ resource.ResourceWithConfigValidators = &sshKeysResource{}
	_ resource.ResourceWithImportState      = &sshKeysResource{}
)

const (
	errorImportingSSHKeys = "error importing ssh_keys resource"
	errorWritingSSHKeys   = "error writing ssh_keys resource"
	errorReadingSSHKeys   = "error reading ssh_keys resource"
	errorDeletingSSHKeys  = "error deleting ssh_keys resource"
)

// NewSSHKeysResource is a helper function to simplify the provider implementation.
func NewSSHKeysResource() resource.Resource {
	return &sshKeysResource{}
}

type sshKeysResource struct {
	client *client.Client
}

type sshKeysResourceModel struct {
	Node   types.String `tfsdk:"node"`
	Tag    types.String `tfsdk:"tag"`
	Extend types.Bool   `tfsdk:"extend"`
	Users  []sskKey     `tfsdk:"users"`
}

func (m sshKeysResourceModel) typeAndIdentifier() (config.EntityType, string) {
	return typeAndIdentifier(m.Tag, m.Node)
}

type sskKey struct {
	Username types.String   `tfsdk:"username"`
	Keys     []types.String `tfsdk:"keys"`
}

// Metadata returns the resource type name.
func (r *sshKeysResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_ssh_keys"
}

// Configure adds the provider configured client to the resource.
func (r *sshKeysResource) Configure(_ context.Context, req resource.ConfigureRequest, _ *resource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}

	r.client = req.ProviderData.(*client.Client)
}

// Schema defines the schema for the resource.
func (r *sshKeysResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "SSHKeys adds or removes authorized SSH keys for users.",
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
				Description: "The users to set SSH keys for.",
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"username": schema.StringAttribute{
							Required:    true,
							Description: "Username of the user for which the SSH keys are set.",
						},
						"keys": schema.ListAttribute{
							Required:    true,
							Description: "The SSH keys to set for the user.",
							ElementType: types.StringType,
						},
					},
				},
			},
		},
	}
}

func (r *sshKeysResource) ConfigValidators(ctx context.Context) []resource.ConfigValidator {
	return []resource.ConfigValidator{
		resourcevalidator.ExactlyOneOf(
			path.MatchRoot("tag"),
			path.MatchRoot("node"),
		),
	}
}

// Create creates the resource and sets the initial Terraform state.
func (r *sshKeysResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	// Retrieve values from the plan
	var plan sshKeysResourceModel
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	diags = r.writeSSHKeys(ctx, plan)
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
func (r *sshKeysResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	// Retrieve values from the plan
	var plan sshKeysResourceModel
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	diags = r.writeSSHKeys(ctx, plan)
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
func (r *sshKeysResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	// Get the current state
	var state *sshKeysResourceModel
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	configType, identifier := state.typeAndIdentifier()

	// Read the real status
	activeConfig, err := r.client.GetActiveConfig(ctx, configType, identifier, config.EntityConfigScopeOwn)
	if err != nil {
		resp.Diagnostics.AddError(errorReadingSSHKeys,
			"error reading the active configuration: "+err.Error())

		return
	}

	// Update the current state
	currentSSHKeys := activeConfig.BundleData.SSHKeys
	if currentSSHKeys == nil {
		resp.State.RemoveResource(ctx)
		return
	}

	state.Extend = types.BoolValue(currentSSHKeys.Extend)

	var mappedUsers []sskKey
	for _, user := range currentSSHKeys.Users {
		keys := make([]types.String, 0)
		for _, key := range user.Keys {
			keys = append(keys, types.StringValue(key))
		}

		mappedUsers = append(mappedUsers, sskKey{
			Username: types.StringValue(user.Username),
			Keys:     keys,
		})
	}
	state.Users = mappedUsers

	diags = resp.State.Set(ctx, state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}

// Delete deletes the resource and removes the Terraform state on success.
func (r *sshKeysResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	// Retrieve values from the state
	var state sshKeysResourceModel
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Delete the resource
	configType, identifier := state.typeAndIdentifier()
	tflog.Info(ctx, fmt.Sprintf("Deleting ssh_keys for %v %v", configType, identifier))

	content := config.SSHKeys{
		Metadata: config.Metadata{
			Reset:   true,
			Version: "v1",
		},
	}

	changeRequest, err := createChangeRequest(config.SSHKeysBundle, content, configType, identifier)
	if err != nil {
		resp.Diagnostics.AddError(
			errorDeletingSSHKeys,
			err.Error(),
		)
		return
	}

	change, err := r.client.CreateConfigurationChange(ctx, changeRequest)
	if err != nil {
		resp.Diagnostics.AddError(
			errorDeletingSSHKeys,
			err.Error(),
		)
		return
	}

	_, err = r.client.CommitConfiguration(ctx, "terraform: create ssh_keys")
	if err != nil {
		resp.Diagnostics.AddError(errorDeletingSSHKeys,
			"error creating a commit to delete the ssh_keys resource: "+err.Error(),
		)

		err = r.client.DeleteConfigurationChange(ctx, change.SHA)
		if err != nil {
			resp.Diagnostics.AddError(
				errorDeletingSSHKeys,
				"error deleting uncommitted ssh_keys changes: "+err.Error(),
			)
		}

		return
	}
}

// ImportState imports the resource state from the Terraform state.
func (r *sshKeysResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	configType, identifier, found := strings.Cut(req.ID, ":")
	if !found || configType == "" || identifier == "" {
		resp.Diagnostics.AddError(
			errorImportingSSHKeys,
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
			errorImportingSSHKeys,
			fmt.Sprintf("Import type must be either 'node' or 'tag'. Got: %q", configType),
		)
		return
	}
}

func (r *sshKeysResource) writeSSHKeys(ctx context.Context, plan sshKeysResourceModel) diag.Diagnostics {
	configType, identifier := plan.typeAndIdentifier()
	extend := plan.Extend.ValueBool()

	// Create the resource
	tflog.Info(ctx, fmt.Sprintf("Creating ssh_keys for %v %v", configType, identifier))

	var mappedUsers []config.SSHKey
	for _, user := range plan.Users {
		var keys []string
		for _, key := range user.Keys {
			keys = append(keys, key.ValueString())
		}

		mappedUsers = append(mappedUsers, config.SSHKey{
			Username: user.Username.ValueString(),
			Keys:     keys,
		})
	}

	content := config.SSHKeys{
		Metadata: config.Metadata{
			Enabled: true,
			Extend:  extend,
			Version: "v1",
		},
		Users: mappedUsers,
	}

	changeRequest, err := createChangeRequest(config.SSHKeysBundle, content, configType, identifier)
	if err != nil {
		return diag.Diagnostics{
			diag.NewErrorDiagnostic(
				errorWritingSSHKeys,
				err.Error(),
			),
		}
	}

	change, err := r.client.CreateConfigurationChange(ctx, changeRequest)
	if err != nil {
		return diag.Diagnostics{
			diag.NewErrorDiagnostic(
				errorWritingSSHKeys,
				fmt.Sprintf("Error creating a ssh_keys resource with qbee: %v", err),
			),
		}
	}

	_, err = r.client.CommitConfiguration(ctx, "terraform: create ssh_keys")
	if err != nil {
		diags := diag.Diagnostics{}

		err = fmt.Errorf("error creating a commit for the ssh_keys: %w", err)
		diags.AddError(errorWritingSSHKeys, err.Error())

		err = r.client.DeleteConfigurationChange(ctx, change.SHA)
		if err != nil {
			diags.AddError(
				errorWritingSSHKeys,
				fmt.Errorf("error deleting uncommitted ssh_keys changes: %w", err).Error(),
			)
		}

		return diags
	}

	return nil
}
