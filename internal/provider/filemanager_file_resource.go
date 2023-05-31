package provider

import (
	"context"
	"fmt"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/lesteenman/terraform-provider-qbee/internal/qbee"
)

// Ensure the implementation satisfies the expected interfaces.
var (
	_ resource.Resource              = &filemanagerFileResource{}
	_ resource.ResourceWithConfigure = &filemanagerFileResource{}
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
				Description: "Identifier attribute. Equal to 'path'.",
			},
			"path": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "The full path of the directory. Equal to `{parent}/{name}`.",
			},
			"parent": schema.StringAttribute{
				Required:      true,
				PlanModifiers: []planmodifier.String{stringplanmodifier.RequiresReplace()},
				Description: "The parent directory of the file. Must include a trailing slash. " +
					"The parent will be created as an unmanaged directory if it does not yet exist.",
			},
			"name": schema.StringAttribute{
				Computed:      true,
				Optional:      true,
				PlanModifiers: []planmodifier.String{stringplanmodifier.RequiresReplace()},
				Description:   "The name of the directory. Defaults to the name of the sourcefile if left empty.",
			},
			"sourcefile": schema.StringAttribute{
				Required:      true,
				PlanModifiers: []planmodifier.String{stringplanmodifier.RequiresReplace()},
				Description:   "The source file to upload.",
			},
			"file_hash": schema.StringAttribute{
				Required:      true,
				PlanModifiers: []planmodifier.String{stringplanmodifier.RequiresReplace()},
				Description: "The hash of the source file. Required to ensure resource updates if the file changes. " +
					"Should be equal to `filesha1(sourcefile)`.",
			},
		},
	}
}

type filemanagerFileResourceModel struct {
	ID         types.String `tfsdk:"id"`
	Path       types.String `tfsdk:"path"`
	Parent     types.String `tfsdk:"parent"`
	Name       types.String `tfsdk:"name"`
	SourceFile types.String `tfsdk:"sourcefile"`
	FileHash   types.String `tfsdk:"file_hash"`
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
	parent := plan.Parent.ValueString()
	sourceFile := plan.SourceFile.ValueString()
	filename := plan.Name.ValueString()
	uploadFileResponse, err := r.client.Files.Upload(sourceFile, parent, filename)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error creating filemanager_file",
			"could not create filemanager file, unexpected error: "+err.Error())
		return
	}

	// Map response body to schema and populate Computed attribute values
	fullFilePath := uploadFileResponse.Path + uploadFileResponse.File
	plan.ID = types.StringValue(fullFilePath)
	plan.Path = types.StringValue(fullFilePath)
	plan.Name = types.StringValue(uploadFileResponse.File)

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

	fullFilePath := state.ID.ValueString()

	// Get the current file from Qbee
	listFilesResponse, err := r.client.Files.List()
	tflog.Info(ctx, fmt.Sprintf("trying to find item %v", state.ID.ValueString()))
	if err != nil {
		resp.Diagnostics.AddError(
			"Error reading Qbee Filemanager data",
			"Could not read Filemanager data from Qbee: "+err.Error())
	}

	var fileName string

	exists := false
	for _, item := range listFilesResponse.Items {
		if item.Path == fullFilePath && !item.IsDir {
			exists = true
			fileName = item.Name
		}
	}

	// Delete from the state if it no longer exists
	if !exists {
		resp.State.RemoveResource(ctx)
	} else {
		fullFilePath = state.Parent.ValueString() + fileName
		//state.ID = types.StringValue(fullFilePath)
		state.Path = types.StringValue(fullFilePath)
	}

	resp.State.Set(ctx, state)
}

// Update updates the resource and sets the updated Terraform state on success.
func (r *filemanagerFileResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	resp.Diagnostics.AddError(
		"Error updating filemanager_file",
		"filemanager_file")
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
	path := state.Path.ValueString()
	tflog.Info(ctx, fmt.Sprintf("Deleting filemanager path '%v'", path))
	err := r.client.Files.Delete(path)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error deleting filemanager_path",
			"could not delete filemanager path, unexpected error: "+err.Error())
		return
	}
}
