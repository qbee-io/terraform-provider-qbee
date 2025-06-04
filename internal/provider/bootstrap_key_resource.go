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
	_ resource.Resource                = &bootstrapKeyResource{}
	_ resource.ResourceWithConfigure   = &bootstrapKeyResource{}
	_ resource.ResourceWithImportState = &bootstrapKeyResource{}
)

const (
	errorImportingBootstrapKey = "error importing bootstrap_key resource"
	errorWritingBootstrapKey   = "error writing bootstrap_key resource"
	errorReadingBootstrapKey   = "error reading bootstrap_key resource"
	errorDeletingBootstrapKey  = "error deleting bootstrap_key resource"
)

// NewBootstrapKey is a helper function to simplify the provider implementation.
func NewBootstrapKeyResource() resource.Resource {
	return &bootstrapKeyResource{}
}

type bootstrapKeyResource struct {
	client *client.Client
}

type bootstrapKeyResourceModel struct {
	Id         types.String `tfsdk:"id"`
	GroupId    types.String `tfsdk:"group_id"`
	AutoAccept types.Bool   `tfsdk:"auto_accept"`
}

// Metadata returns the resource type name.
func (r *bootstrapKeyResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_bootstrap_key"
}

// Configure adds the provider configured client to the resource.
func (r *bootstrapKeyResource) Configure(_ context.Context, req resource.ConfigureRequest, _ *resource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}

	r.client = req.ProviderData.(*client.Client)
}

// Schema defines the schema for the resource.
func (r *bootstrapKeyResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed:    true,
				Optional:    true,
				Description: "The actual bootstrap key.",
				Sensitive:   true,
			},
			"group_id": schema.StringAttribute{
				Required:    true,
				Description: "The group ID associated with the bootstrap key.",
			},
			"auto_accept": schema.BoolAttribute{
				Required:    true,
				Description: "Indicates whether the bootstrap key is auto accepted.",
			},
		},
	}
}

// Create creates the resource and sets the initial Terraform state.
func (r *bootstrapKeyResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	// Retrieve values from the plan
	var plan bootstrapKeyResourceModel
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	tflog.Info(ctx, fmt.Sprintf("Creating bootstrap key associated with group ID: %s", plan.GroupId.ValueString()))

	bootstrapKey, err := r.client.NewBootstrapKey(ctx)
	if err != nil {
		resp.Diagnostics.AddError(errorWritingBootstrapKey,
			"error writing the bootstrapKey configuration: "+err.Error())
		return
	}

	// New bootstrap key is created, so we need to set the group ID and auto accept values
	payload := client.BootstrapKey{
		ID:         bootstrapKey.ID,
		GroupID:    plan.GroupId.ValueString(),
		AutoAccept: plan.AutoAccept.ValueBool(),
	}
	err = r.client.UpdateBoostrapKey(ctx, &payload)
	if err != nil {
		resp.Diagnostics.AddError(errorWritingBootstrapKey,
			"error writing the bootstrapKey: "+err.Error())
		return
	}

	bootstrapKey, err = r.client.GetBootstrapKey(ctx, bootstrapKey.ID)
	if err != nil {
		resp.Diagnostics.AddError(errorReadingBootstrapKey,
			"error reading the bootstrapKey: "+err.Error())
		return
	}

	// Update the current state with the actual values
	plan.Id = types.StringValue(bootstrapKey.ID)
	plan.GroupId = types.StringValue(bootstrapKey.GroupID)
	plan.AutoAccept = types.BoolValue(bootstrapKey.AutoAccept)

	// Set state to fully populated data
	diags = resp.State.Set(ctx, plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}

// Update updates the resource and sets the updated Terraform state on success.
func (r *bootstrapKeyResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	// Retrieve values from the plan
	var plan bootstrapKeyResourceModel
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)

	var state bootstrapKeyResourceModel
	diags = req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)

	if resp.Diagnostics.HasError() {
		return
	}

	tflog.Info(ctx, fmt.Sprintf("Updating bootstrap key associated with group ID: %s", plan.GroupId.ValueString()))
	payload := client.BootstrapKey{
		ID:         state.Id.ValueString(),
		GroupID:    plan.GroupId.ValueString(),
		AutoAccept: plan.AutoAccept.ValueBool(),
	}

	err := r.client.UpdateBoostrapKey(ctx, &payload)
	if err != nil {
		resp.Diagnostics.AddError(errorWritingBootstrapKey,
			"error writing the bootstrap key: "+err.Error())
		return
	}

	// Update the current state with the actual values
	plan.Id = types.StringValue(payload.ID)
	plan.GroupId = types.StringValue(payload.GroupID)
	plan.AutoAccept = types.BoolValue(payload.AutoAccept)

	// Map response body to schema and populate Computed attribute values
	// Set state to fully populated data
	diags = resp.State.Set(ctx, plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}

// Read refreshes the Terraform state with the latest data.
func (r *bootstrapKeyResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	// Get the current state
	var state *bootstrapKeyResourceModel
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Read the real status
	bootstrapKey, err := r.client.GetBootstrapKey(ctx, state.Id.ValueString())
	if err != nil {
		resp.Diagnostics.AddError(errorReadingBootstrapKey,
			"error reading the bootstrap_key: "+err.Error())

		return
	}

	// Update the current state
	if bootstrapKey == nil {
		resp.State.RemoveResource(ctx)
		return
	}

	state.Id = types.StringValue(bootstrapKey.ID)
	state.GroupId = types.StringValue(bootstrapKey.GroupID)
	state.AutoAccept = types.BoolValue(bootstrapKey.AutoAccept)

	diags = resp.State.Set(ctx, state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}

// Delete deletes the resource and removes the Terraform state on success.
func (r *bootstrapKeyResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	// Retrieve values from the state
	var state bootstrapKeyResourceModel
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Delete the resource
	tflog.Info(ctx, fmt.Sprintf("Deleting bootstrap key associated with group ID: %s", state.GroupId.ValueString()))

	err := r.client.DeleteBootstrapKey(ctx, state.Id.ValueString())
	if err != nil {
		resp.Diagnostics.AddError(errorDeletingBootstrapKey,
			"error deleting the bootstrap key: "+err.Error())
		return
	}
}

// ImportState imports the resource state from the Terraform state.
func (r *bootstrapKeyResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	name := req.ID
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("id"), name)...)
}
