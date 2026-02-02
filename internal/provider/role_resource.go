package provider

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	"go.qbee.io/client"
)

// Ensure the implementation satisfies the expected interfaces.
var (
	_ resource.Resource                = &roleResource{}
	_ resource.ResourceWithConfigure   = &roleResource{}
	_ resource.ResourceWithImportState = &roleResource{}
)

const (
	errorImportingRole = "error importing role resource"
	errorCreatingRole  = "error creating role resource"
	errorUpdatingRole  = "error updating role resource"
	errorReadingRole   = "error reading role resource"
	errorDeletingRole  = "error deleting role resource"
)

// NewRoleResource is a helper function to simplify the provider implementation.
func NewRoleResource() resource.Resource {
	return &roleResource{}
}

type roleResource struct {
	client *client.Client
}

type roleResourceModel struct {
	Id          types.String `tfsdk:"id"`
	Name        types.String `tfsdk:"name"`
	Description types.String `tfsdk:"description"`
	Policies    []policy     `tfsdk:"policies"`
}

type policy struct {
	Permission types.String `tfsdk:"permission"`
	Resources  []string     `tfsdk:"resources"`
}

// Metadata returns the resource type name.
func (r *roleResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_role"
}

// Configure adds the provider configured client to the resource.
func (r *roleResource) Configure(_ context.Context, req resource.ConfigureRequest, _ *resource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}

	r.client = req.ProviderData.(*client.Client)
}

// Schema defines the schema for the resource.
func (r *roleResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed:    true,
				Optional:    true,
				Description: "The unique identifier of the role.",
			},
			"name": schema.StringAttribute{
				Required:    true,
				Description: "The short name of the role.",
			},
			"description": schema.StringAttribute{
				Optional:    true,
				Description: "The optional description of the role.",
			},
			"policies": schema.ListNestedAttribute{
				Optional:    true,
				Description: "The list of policies that are assigned to this role.",
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"permission": schema.StringAttribute{
							Required:    true,
							Description: "The permission that is granted by this policy.",
						},
						"resources": schema.ListAttribute{
							Optional:    true,
							Description: "The list of resources that are affected by this policy. Use `*` to match all resources.",
							ElementType: types.StringType,
						},
					},
				},
			},
		},
	}
}

// Create creates the resource and sets the initial Terraform state.
func (r *roleResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	// Retrieve values from the plan
	var plan roleResourceModel
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	tflog.Info(ctx, fmt.Sprintf("Creating role %v", plan.Name))

	payload := client.Role{
		Name:        plan.Name.ValueString(),
		Description: plan.Description.ValueString(),
		Policies:    make([]client.RolePolicy, len(plan.Policies)),
	}
	for i, p := range plan.Policies {
		payload.Policies[i] = client.RolePolicy{
			Permission: client.Permission(p.Permission.ValueString()),
			Resources:  p.Resources,
		}
	}

	role, err := r.client.CreateRole(ctx, payload)
	if err != nil {
		resp.Diagnostics.AddError(errorCreatingRole,
			"error writing the role: "+err.Error())
		return
	}

	// Update the current state with the actual values
	plan.Id = types.StringValue(role.ID)
	plan.Description = types.StringValue(role.Description)
	plan.Policies = make([]policy, len(role.Policies))
	for i, p := range role.Policies {
		plan.Policies[i] = policy{
			Permission: types.StringValue(string(p.Permission)),
			Resources:  p.Resources,
		}
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
func (r *roleResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	// Retrieve values from the plan
	var plan roleResourceModel
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)

	var state roleResourceModel
	diags = req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)

	if resp.Diagnostics.HasError() {
		return
	}

	tflog.Info(ctx, fmt.Sprintf("Updating role %v with id %v (or %v)", plan.Name, plan.Id, state.Id))

	payload := client.Role{
		ID:          state.Id.ValueString(),
		Name:        plan.Name.ValueString(),
		Description: plan.Description.ValueString(),
		Policies:    make([]client.RolePolicy, len(plan.Policies)),
	}
	for i, p := range plan.Policies {
		payload.Policies[i] = client.RolePolicy{
			Permission: client.Permission(p.Permission.ValueString()),
			Resources:  p.Resources,
		}
	}

	role, err := r.client.UpdateRole(ctx, payload)
	if err != nil {
		resp.Diagnostics.AddError(errorUpdatingRole,
			"error updating the role: "+err.Error())
		return
	}

	// Update the current state with the actual values
	plan.Id = types.StringValue(role.ID)
	plan.Description = types.StringValue(role.Description)
	plan.Policies = make([]policy, len(role.Policies))
	for i, p := range role.Policies {
		plan.Policies[i] = policy{
			Permission: types.StringValue(string(p.Permission)),
			Resources:  p.Resources,
		}
	}

	// Set state to fully populated data
	diags = resp.State.Set(ctx, plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}

// Read refreshes the Terraform state with the latest data.
func (r *roleResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	// Get the current state
	var state *roleResourceModel
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Read the real status
	roles, err := r.client.ListRoles(ctx)
	if err != nil {
		resp.Diagnostics.AddError(errorReadingRole,
			"error reading the role: "+err.Error())

		return
	}

	// Find the role in the list
	var activeRole *client.Role
	for _, role := range roles {
		if role.Name == state.Name.ValueString() {
			activeRole = &role
			break
		}
	}

	// Update the current state
	if activeRole == nil {
		resp.State.RemoveResource(ctx)
		return
	}

	state.Id = types.StringValue(activeRole.ID)
	state.Description = types.StringValue(activeRole.Description)
	state.Policies = make([]policy, len(activeRole.Policies))
	for i, p := range activeRole.Policies {
		state.Policies[i] = policy{
			Permission: types.StringValue(string(p.Permission)),
			Resources:  p.Resources,
		}
	}

	diags = resp.State.Set(ctx, state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}

// Delete deletes the resource and removes the Terraform state on success.
func (r *roleResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	// Retrieve values from the state
	var state roleResourceModel
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Delete the resource
	tflog.Info(ctx, fmt.Sprintf("Deleting role %v with id %v", state.Name, state.Id))

	err := r.client.DeleteRole(ctx, state.Id.ValueString())
	if err != nil {
		resp.Diagnostics.AddError(errorDeletingRole,
			"error deleting the role: "+err.Error())
		return
	}
}

// ImportState imports the resource state from the Terraform state.
func (r *roleResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	name := req.ID
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("name"), name)...)
}
