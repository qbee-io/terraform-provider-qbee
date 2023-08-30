package provider

import (
	"bitbucket.org/booqsoftware/terraform-provider-qbee/internal/qbee"
	"context"
	"fmt"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
)

// Ensure the implementation satisfies the expected interfaces.
var (
	_ resource.Resource                = &tagFiledistributionResource{}
	_ resource.ResourceWithConfigure   = &tagFiledistributionResource{}
	_ resource.ResourceWithImportState = &tagFiledistributionResource{}
)

func NewTagFiledistributionResource() resource.Resource {
	return &tagFiledistributionResource{}
}

type tagFiledistributionResource struct {
	client *qbee.HttpClient
}

// Metadata returns the resource type name.
func (r *tagFiledistributionResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_tag_filedistribution"
}

// Configure adds the provider configured client to the resource.
func (r *tagFiledistributionResource) Configure(_ context.Context, req resource.ConfigureRequest, _ *resource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}

	r.client = req.ProviderData.(*qbee.HttpClient)
}

// Schema defines the schema for the resource.
func (r *tagFiledistributionResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed:    true,
				Description: "Placeholder ID value",
			},
			"tag": schema.StringAttribute{
				Required:      true,
				Description:   "The tag for which to set the configuration",
				PlanModifiers: []planmodifier.String{stringplanmodifier.RequiresReplace()},
			},
			"extend": schema.BoolAttribute{
				Required: true,
				Description: "If the tag configuration should extend configuration from the parent nodes of the node " +
					"the tag is applied to. If set to false, configuration from parent nodes is ignored.",
			},
			"files": schema.ListNestedAttribute{
				Required:    true,
				Description: "The filesets that must be distributed",
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"templates": schema.ListNestedAttribute{
							Required: true,
							NestedObject: schema.NestedAttributeObject{
								Attributes: map[string]schema.Attribute{
									"source": schema.StringAttribute{
										Required: true,
										Description: "The source of the file. Must correspond to a file in the qbee " +
											"filemanager.",
									},
									"destination": schema.StringAttribute{
										Required:    true,
										Description: "The destination of the file on the target device.",
									},
									"is_template": schema.BoolAttribute{
										Required: true,
										Description: "If this file is a template. If set to true, template " +
											"substitution of '\\{\\{ pattern \\}\\}' will be performed in the file contents, " +
											"using the parameters defined in this filedistribution config.",
									},
								},
							},
						},
						"parameters": schema.ListNestedAttribute{
							Optional: true,
							NestedObject: schema.NestedAttributeObject{
								Attributes: map[string]schema.Attribute{
									"key": schema.StringAttribute{
										Required: true,
									},
									"value": schema.StringAttribute{
										Required: true,
									},
								},
							},
						},
						"pre_condition": schema.StringAttribute{
							Optional: true,
							Description: "A command that must successfully execute on the device (return a non-zero " +
								"exit code) before this fileset can be distributed. Example: `/bin/true`.",
						},
						"command": schema.StringAttribute{
							Optional: true,
							Description: "A command that will be run on the device after this fileset is " +
								"distributed. Example: `/bin/true`.",
						},
					},
				},
			},
		},
	}
}

type tagFiledistributionResourceModel struct {
	ID     types.String `tfsdk:"id"`
	Tag    types.String `tfsdk:"tag"`
	Extend types.Bool   `tfsdk:"extend"`
	Files  types.List   `tfsdk:"files"`
}

// Create creates the resource and sets the initial Terraform state.
func (r *tagFiledistributionResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	// Retrieve values from the plan
	var plan tagFiledistributionResourceModel
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	err := r.writeFiledistribution(ctx, plan)
	if err != nil {
		resp.Diagnostics.AddError("Could not create tag_filedistribution",
			err.Error())
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

// Update updates the resource and sets the updated Terraform state on success.
func (r *tagFiledistributionResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	// Retrieve values from the plan
	var plan tagFiledistributionResourceModel
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	err := r.writeFiledistribution(ctx, plan)
	if err != nil {
		resp.Diagnostics.AddError("Could not update tag_filedistribution",
			err.Error())
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

func (r *tagFiledistributionResource) writeFiledistribution(ctx context.Context, plan tagFiledistributionResourceModel) error {
	tag := plan.Tag.ValueString()
	extend := plan.Extend.ValueBool()

	var files []FiledistributionFile
	diags := plan.Files.ElementsAs(ctx, &files, false)
	if diags.HasError() {
		// Note: this might silence some warnings... Redo at some point.
		return fmt.Errorf("%v: %v", diags.Errors()[0].Summary(), diags.Errors()[0].Detail())
	}

	filesets := PlanToQbeeFilesets(ctx, files)

	// Create the resource
	tflog.Info(ctx, fmt.Sprintf("Creating file distribution for tag %v with %v filesets", tag, len(filesets)))
	createResponse, err := r.client.Config.CreateTagFileDistribution(tag, filesets, extend)
	if err != nil {
		return fmt.Errorf("error creating a tag_filedistribution resource: %w", err)
	}

	_, err = r.client.Configuration.Commit("terraform: create tag_filedistribution_resource")
	if err != nil {
		err = fmt.Errorf("error creating a commit for the tag_filedistribution: %w", err)

		err = r.client.Configuration.DeleteUncommitted(createResponse.Sha)
		if err != nil {
			err = fmt.Errorf("error deleting uncommitted tag_filedistribution changes: %w", err)
		}

		return err
	}

	return nil
}

// Read refreshes the Terraform state with the latest data.
func (r *tagFiledistributionResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	// Get the current state
	var state tagFiledistributionResourceModel
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	tag := state.Tag.ValueString()

	// Read the real status
	distribution, err := r.client.Config.GetTagFiledistribution(tag)
	if err != nil {
		resp.Diagnostics.AddError("Could not read tag_filedistribution",
			"error reading the tag_filedistribution resource: "+err.Error())

		return
	}

	// Update the current state
	fileValues := filedistributionToListValue(ctx, distribution, resp)
	if resp.Diagnostics.HasError() {
		return
	}

	state.ID = types.StringValue("placeholder")
	state.Extend = types.BoolValue(distribution.Extend)
	state.Files = fileValues

	resp.State.Set(ctx, state)
}

// Delete deletes the resource and removes the Terraform state on success.
func (r *tagFiledistributionResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	// Retrieve values from the state
	var state tagFiledistributionResourceModel
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Delete the resource
	tagName := state.Tag.ValueString()
	tflog.Info(ctx, fmt.Sprintf("Deleting tag_filedistribution for tag %v", tagName))
	deleteResponse, err := r.client.Config.ClearTagFileDistribution(tagName)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error deleting tag_filedistribution",
			"could not delete tag filedistribution, unexpected error: "+err.Error())
		return
	}

	_, err = r.client.Configuration.Commit("terraform: create tag_filedistribution_resource")
	if err != nil {
		resp.Diagnostics.AddError("Could not commit deletion of tag_filedistribution",
			"error creating a commit to delete the tag_filedistribution resource: "+err.Error())

		err = r.client.Configuration.DeleteUncommitted(deleteResponse.Sha)
		if err != nil {
			resp.Diagnostics.AddError("Could not revert uncommitted tag_filedistribution changes",
				"error deleting uncommitted tag_filedistribution changes: "+err.Error())
		}

		return
	}
}

func (r *tagFiledistributionResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("tag"), req, resp)
}
