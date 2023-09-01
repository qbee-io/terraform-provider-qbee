package provider

import (
	"bitbucket.org/booqsoftware/terraform-provider-qbee/internal/qbee"
	"context"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// Ensure the implementation satisfies the expected interfaces.
var (
	_ resource.Resource                = &nodeSoftwaremanagementResource{}
	_ resource.ResourceWithConfigure   = &nodeSoftwaremanagementResource{}
	_ resource.ResourceWithImportState = &nodeSoftwaremanagementResource{}
)

func NewTagConfigurationResource() resource.Resource {
	return &nodeSoftwaremanagementResource{}
}

type nodeSoftwaremanagementResource struct {
	client *qbee.HttpClient
}

// Metadata returns the resource type name.
func (r *nodeSoftwaremanagementResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_grouptree_group"
}

// Configure adds the provider configured client to the resource.
func (r *nodeSoftwaremanagementResource) Configure(_ context.Context, req resource.ConfigureRequest, _ *resource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}

	r.client = req.ProviderData.(*qbee.HttpClient)
}

// Schema defines the schema for the resource.
func (r *nodeSoftwaremanagementResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed:    true,
				Description: "Placeholder ID value",
			},
		},
	}
}

type nodeSoftwaremanagementResourceModel struct {
	ID types.String `tfsdk:"id"`
}

// Create creates the resource and sets the initial Terraform state.
func (r *nodeSoftwaremanagementResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	// Retrieve values from the plan
	var plan nodeSoftwaremanagementResourceModel
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Create the resource

	// Map response body to schema and populate Computed attribute values

	// Set state to fully populated data
}

// Read refreshes the Terraform state with the latest data.
func (r *nodeSoftwaremanagementResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	// Get the current state
	var state nodeSoftwaremanagementResourceModel
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Read the real status

	// Update the current state
	resp.State.Set(ctx, state)
}

// Update updates the resource and sets the updated Terraform state on success.
func (r *nodeSoftwaremanagementResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
}

// Delete deletes the resource and removes the Terraform state on success.
func (r *nodeSoftwaremanagementResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	// Retrieve values from the state
	var state nodeSoftwaremanagementResourceModel
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Delete the resource
}

func (r *nodeSoftwaremanagementResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}
