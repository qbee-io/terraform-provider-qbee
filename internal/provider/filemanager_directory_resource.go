package provider

import (
	"context"
	"errors"
	"fmt"
	"path/filepath"

	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"

	"go.qbee.io/terraform/internal/qbee"
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
				Description: "Placeholder ID value",
			},
			"path": schema.StringAttribute{
				Required:            true,
				PlanModifiers:       []planmodifier.String{stringplanmodifier.RequiresReplace()},
				MarkdownDescription: "The full path of the directory. Must not include a trailing slash. Example: /parent/directory",
			},
		},
	}
}

type filemanagerDirectoryResourceModel struct {
	ID   types.String `tfsdk:"id"`
	Path types.String `tfsdk:"path"`
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
	directoryPath := plan.Path.ValueString()
	pathCleaned := filepath.Clean(directoryPath)
	pathParent := filepath.Dir(pathCleaned)
	pathName := filepath.Base(pathCleaned)
	tflog.Info(ctx, fmt.Sprintf("Creating filemanager directory %v/%v", pathParent, pathName))
	_, err := r.client.Files.CreateDir(pathParent, pathName)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error creating filemanager_directory",
			"could not create filemanager directory, unexpected error: "+err.Error())
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
func (r *filemanagerDirectoryResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	// Get the current state
	var state filemanagerDirectoryResourceModel
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Get the refreshed directory value from Qbee
	directoryPath := state.Path.ValueString()

	listFilesResponse, err := r.client.Files.GetMetadata(directoryPath)

	// If the directory is not found, we have drift, and it was deleted from qbee
	if errors.Is(err, qbee.ErrFileNotFound) {
		resp.State.RemoveResource(ctx)
		return
	}

	// Any other error is unexpected
	if err != nil {
		resp.Diagnostics.AddError(
			"Error reading Qbee Filemanager data",
			"Could not read Filemanager data from Qbee: "+err.Error())
		return
	}

	// Delete from the state if it no longer exists
	if !listFilesResponse.IsDir {
		resp.State.RemoveResource(ctx)
		return
	}

	// Update the state
	state.ID = types.StringValue("placeholder")
	state.Path = types.StringValue(directoryPath)
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
	directoryPath := state.Path.ValueString()
	tflog.Info(ctx, fmt.Sprintf("Deleting filemanager directory '%v'", directoryPath))

	err := r.client.Files.Delete(directoryPath)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error deleting filemanager_directory",
			"could not delete filemanager directory, unexpected error: "+err.Error())
		return
	}
}

func (r *filemanagerDirectoryResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	cleanPath := filepath.Clean(req.ID)

	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("id"), "placeholder")...)
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("path"), cleanPath)...)
}
