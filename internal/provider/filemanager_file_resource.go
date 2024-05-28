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
	_ resource.Resource                = &filemanagerFileResource{}
	_ resource.ResourceWithConfigure   = &filemanagerFileResource{}
	_ resource.ResourceWithImportState = &filemanagerFileResource{}
)

func NewFilemanagerFileResource() resource.Resource {
	return &filemanagerFileResource{}
}

type filemanagerFileResource struct {
	client *qbee.HttpClient
}

// Metadata returns the resource type name.
func (r *filemanagerFileResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_filemanager_file"
}

// Configure adds the provider configured client to the resource.
func (r *filemanagerFileResource) Configure(_ context.Context, req resource.ConfigureRequest, _ *resource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}

	r.client = req.ProviderData.(*qbee.HttpClient)
}

// Schema defines the schema for the resource.
func (r *filemanagerFileResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed:    true,
				Description: "Placeholder ID value",
			},
			"path": schema.StringAttribute{
				Required:            true,
				PlanModifiers:       []planmodifier.String{stringplanmodifier.RequiresReplace()},
				MarkdownDescription: "The full path of the uploaded file.",
			},
			"sourcefile": schema.StringAttribute{
				Required:      true,
				PlanModifiers: []planmodifier.String{stringplanmodifier.RequiresReplace()},
				Description:   "The source file to upload.",
			},
			"file_sha256": schema.StringAttribute{
				Required:      true,
				PlanModifiers: []planmodifier.String{stringplanmodifier.RequiresReplace()},
				Description: "The filebase64sha256 of the source file. Required to ensure resource " +
					"updates if the file changes.",
			},
		},
	}
}

type filemanagerFileResourceModel struct {
	ID         types.String `tfsdk:"id"`
	Path       types.String `tfsdk:"path"`
	SourceFile types.String `tfsdk:"sourcefile"`
	FileSha256 types.String `tfsdk:"file_sha256"`
}

// Create creates the resource and sets the initial Terraform state.
func (r *filemanagerFileResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	// Retrieve values from the plan
	var plan filemanagerFileResourceModel
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Upload the file
	uploadPath := plan.Path.ValueString()
	pathCleaned := filepath.Clean(uploadPath)
	fileDirectory := filepath.Dir(pathCleaned)
	fileName := filepath.Base(pathCleaned)
	sourceFile := plan.SourceFile.ValueString()
	tflog.Info(ctx, fmt.Sprintf("Uploading file %v to %v/%v", sourceFile, fileDirectory, fileName))
	_, err := r.client.Files.Upload(sourceFile, fileDirectory, fileName)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error creating filemanager_file",
			"could not create filemanager file, unexpected error: "+err.Error())
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
func (r *filemanagerFileResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	// Get the current state
	var state filemanagerFileResourceModel
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	filePath := state.Path.ValueString()

	// Get the current file from Qbee
	fileInfo, err := r.client.Files.GetMetadata(filePath)

	// If the file is not found, we have drift, and it was deleted from qbee
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

	// If the file was found, update the current state
	state.ID = types.StringValue("placeholder")
	state.FileSha256 = types.StringValue(fileInfo.Digest)

	resp.State.Set(ctx, state)
}

// Update updates the resource and sets the updated Terraform state on success.
func (r *filemanagerFileResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	resp.Diagnostics.AddError(
		"Error updating filemanager_file",
		"filemanager_file does not support updating resources inline")
	return
}

// Delete deletes the resource and removes the Terraform state on success.
func (r *filemanagerFileResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	// Retrieve values from the state
	var state filemanagerFileResourceModel
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Delete file
	filePath := state.Path.ValueString()
	tflog.Info(ctx, fmt.Sprintf("Deleting filemanager path '%v'", filePath))
	err := r.client.Files.Delete(filePath)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error deleting filemanager_file",
			fmt.Sprintf("could not delete filemanager_file with path '%v', unexpected error: %v", filePath, err.Error()))
		return
	}
}

func (r *filemanagerFileResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("path"), req, resp)
}
