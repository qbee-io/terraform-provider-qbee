package provider

import (
	"context"
	"fmt"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/lesteenman/terraform-provider-qbee/internal/qbee"
)

// Ensure the implementation satisfies the expected interfaces.
var (
	_ resource.Resource              = &tagFiledistributionResource{}
	_ resource.ResourceWithConfigure = &tagFiledistributionResource{}
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
				Required:    true,
				Description: "The tag for which to set the configuration",
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

type tagFiledistributionFile struct {
	Command      types.String `tfsdk:"command"`
	PreCondition types.String `tfsdk:"pre_condition"`
	Templates    []struct {
		Source      types.String `tfsdk:"source"`
		Destination types.String `tfsdk:"destination"`
		IsTemplate  types.Bool   `tfsdk:"is_template"`
	} `tfsdk:"templates"`
	Parameters *[]struct {
		Key   types.String `tfsdk:"key"`
		Value types.String `tfsdk:"value"`
	} `tfsdk:"parameters"`
}

type tagFiledistributionResourceModel struct {
	ID     types.String              `tfsdk:"id"`
	Tag    types.String              `tfsdk:"tag"`
	Extend types.Bool                `tfsdk:"extend"`
	Files  []tagFiledistributionFile `tfsdk:"files"`
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

	tag := plan.Tag.ValueString()
	extend := plan.Extend.ValueBool()
	filesets := r.createFilesets(plan.Files)

	// Create the resource
	tflog.Info(ctx, fmt.Sprintf("Creating file distribution for tag %v with %v filesets", tag, len(filesets)))
	createResponse, err := r.client.TagConfig.CreateFileDistribution(tag, filesets, extend)
	if err != nil {
		resp.Diagnostics.AddError("Could not create tag_filedistribution",
			"error creating a tag_filedistribution resource: "+err.Error())

		return
	}

	_, err = r.client.Configuration.Commit("terraform: create tag_filedistribution_resource")
	if err != nil {
		resp.Diagnostics.AddError("Could not commit tag_filedistribution",
			"error creating a commit for the tag_filedistribution resource: "+err.Error())

		err = r.client.Configuration.DeleteUncommitted(createResponse.Sha)
		if err != nil {
			resp.Diagnostics.AddError("Could not revert uncommitted tag_filedistribution changes",
				"error deleting uncommitted tag_filedistribution changes: "+err.Error())
		}

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

func (r *tagFiledistributionResource) createFilesets(files []tagFiledistributionFile) []qbee.FilesetConfig {
	var filesets []qbee.FilesetConfig

	for _, file := range files {
		var fsc qbee.FilesetConfig

		if !file.PreCondition.IsNull() {
			fsc.PreCondition = file.PreCondition.ValueString()
		}

		if !file.Command.IsNull() {
			fsc.Command = file.Command.ValueString()
		}

		var params []qbee.FilesetParameter
		if file.Parameters != nil {
			for _, p := range *file.Parameters {
				params = append(params, qbee.FilesetParameter{
					Key:   p.Key.ValueString(),
					Value: p.Value.ValueString(),
				})
			}
		}

		var templates []qbee.FilesetTemplate
		for _, template := range file.Templates {
			templates = append(templates, qbee.FilesetTemplate{
				Source:      template.Source.ValueString(),
				Destination: template.Destination.ValueString(),
				IsTemplate:  template.IsTemplate.ValueBool(),
			})
		}

		filesets = append(filesets, qbee.FilesetConfig{
			PreCondition: file.PreCondition.ValueString(),
			Command:      file.Command.ValueString(),
			Templates:    templates,
			Parameters:   params,
		})
	}

	return filesets
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
	_, err := r.client.TagConfig.GetFiledistribution(tag)
	if err != nil {
		resp.Diagnostics.AddError("Could not read tag_filedistribution",
			"error reading the tag_filedistribution resource: "+err.Error())

		return
	}

	// Update the current state
	resp.State.Set(ctx, state)
}

// Update updates the resource and sets the updated Terraform state on success.
func (r *tagFiledistributionResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	// TODO: Implement
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
	deleteResponse, err := r.client.TagConfig.ClearFileDistribution(tagName)
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
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}
