package provider

import (
	"context"
	"fmt"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/lesteenman/terraform-provider-qbee/internal/qbee"

	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
)

// Ensure the implementation satisfies the expected interfaces.
var (
	_ resource.Resource                = &filemanagerDirectoryResource{}
	_ resource.ResourceWithConfigure   = &filemanagerDirectoryResource{}
	_ resource.ResourceWithImportState = &filemanagerDirectoryResource{}
)

func NewFilemanagerDirectoryResource() resource.Resource {
	return &filemanagerDirectoryResource{}
}

type filemanagerDirectoryResource struct {
	client *qbee.HttpClient
}

// Configure adds the provider configured client to the resource.
func (r *filemanagerDirectoryResource) Configure(_ context.Context, req resource.ConfigureRequest, _ *resource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}

	r.client = req.ProviderData.(*qbee.HttpClient)
}

// Metadata returns the resource type name.
func (r *filemanagerDirectoryResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_filemanager_directory"
}

// Schema defines the schema for the resource.
func (r *filemanagerDirectoryResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed:    true,
				Description: "Identifier attribute. Equal to 'path'.",
			},
			"path": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "The full path of the directory. Equal to `{parent}/{name}/`.",
			},
			"parent": schema.StringAttribute{
				Required:      true,
				PlanModifiers: []planmodifier.String{stringplanmodifier.RequiresReplace()},
				Description: "The parent of the directory. Must include a trailing slash. " +
					"The parent will be created as an unmanaged directory if it does not yet exist.",
			},
			"name": schema.StringAttribute{
				Required:      true,
				PlanModifiers: []planmodifier.String{stringplanmodifier.RequiresReplace()},
				Description:   "The name of the directory.",
			},
		},
	}
}

type filemanagerDirectoryResourceModel struct {
	ID     types.String `tfsdk:"id"`
	Path   types.String `tfsdk:"path"`
	Parent types.String `tfsdk:"parent"`
	Name   types.String `tfsdk:"name"`
}

// Create creates the resource and sets the initial Terraform state.
func (r *filemanagerDirectoryResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	// Retrieve values from the plan
	var plan filemanagerDirectoryResourceModel
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Create new directory
	parent := plan.Parent.ValueString()
	name := plan.Name.ValueString()
	createDirResponse, err := r.client.Files.CreateDir(parent, name)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error creating filemanager_directory",
			"could not create filemanager directory, unexpected error: "+err.Error())
		return
	}

	// Map response body to schema and populate Computed attribute values
	plan.ID = types.StringValue(createDirResponse.Path)
	plan.Path = types.StringValue(createDirResponse.Path)

	// Set state to fully populated data
	diags = resp.State.Set(ctx, plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}

// Read refreshes the Terraform state with the latest data.
func (r *filemanagerDirectoryResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	// Get the current state
	var state filemanagerDirectoryResourceModel
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Get the refreshed directory value from Qbee
	expectedPath := state.ID.ValueString()

	listFilesResponse, err := r.client.Files.List()
	tflog.Info(ctx, fmt.Sprintf("trying to find item %v", state.ID.ValueString()))
	if err != nil {
		resp.Diagnostics.AddError(
			"Error reading Qbee Filemanager data",
			"Could not read Filemanager data from Qbee: "+err.Error())
	}

	exists := false
	for _, item := range listFilesResponse.Items {
		if item.Path == expectedPath && item.IsDir {
			exists = true
		}
	}

	// Delete from the state if it no longer exists
	if !exists {
		resp.State.RemoveResource(ctx)
	} else {
		state.ID = types.StringValue(expectedPath)
		state.Path = types.StringValue(expectedPath)
	}

	resp.State.Set(ctx, state)
}

// Update updates the resource and sets the updated Terraform state on success.
func (r *filemanagerDirectoryResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	resp.Diagnostics.AddError(
		"Error updating filemanager_directory",
		"filemanager_directory does not support in-place updates")
	return
}

// Delete deletes the resource and removes the Terraform state on success.
func (r *filemanagerDirectoryResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	// Retrieve values from the state
	var state filemanagerDirectoryResourceModel
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Delete old directory
	path := state.Path.ValueString()
	tflog.Info(ctx, fmt.Sprintf("Deleting filemanager directory '%v'", path))
	err := r.client.Files.Delete(path)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error deleting filemanager_directory",
			"could not delete filemanager directory, unexpected error: "+err.Error())
		return
	}
}

func (r *filemanagerDirectoryResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	id := req.ID
	if id[len(id)-1:] != "/" {
		resp.Diagnostics.AddError("Error importing filemanager_directory",
			fmt.Sprintf("Invalid directory ID '%v': a directory ID must have a trailing slash.", id))
		return
	}

	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}
