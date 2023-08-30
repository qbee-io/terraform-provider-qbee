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
	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"
	"github.com/hashicorp/terraform-plugin-log/tflog"
)

// Ensure the implementation satisfies the expected interfaces.
var (
	_ resource.Resource                = &nodeFiledistributionResource{}
	_ resource.ResourceWithConfigure   = &nodeFiledistributionResource{}
	_ resource.ResourceWithImportState = &nodeFiledistributionResource{}
)

func NewNodeFiledistributionResource() resource.Resource {
	return &nodeFiledistributionResource{}
}

type nodeFiledistributionResource struct {
	client *qbee.HttpClient
}

// Metadata returns the resource type name.
func (r *nodeFiledistributionResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_node_filedistribution"
}

// Configure adds the provider configured client to the resource.
func (r *nodeFiledistributionResource) Configure(_ context.Context, req resource.ConfigureRequest, _ *resource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}

	r.client = req.ProviderData.(*qbee.HttpClient)
}

// Schema defines the schema for the resource.
func (r *nodeFiledistributionResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed:    true,
				Description: "Placeholder ID value",
			},
			"node": schema.StringAttribute{
				Required:      true,
				Description:   "The node for which to set the configuration",
				PlanModifiers: []planmodifier.String{stringplanmodifier.RequiresReplace()},
			},
			"extend": schema.BoolAttribute{
				Required: true,
				Description: "If the node configuration should extend configuration from the parent nodes of the node " +
					"the node is applied to. If set to false, configuration from parent nodes is ignored.",
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

type nodeFiledistributionResourceModel struct {
	ID     types.String `tfsdk:"id"`
	Node   types.String `tfsdk:"node"`
	Extend types.Bool   `tfsdk:"extend"`
	Files  types.List   `tfsdk:"files"`
}

// Create creates the resource and sets the initial Terraform state.
func (r *nodeFiledistributionResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	// Retrieve values from the plan
	var plan nodeFiledistributionResourceModel
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	err := r.writeFiledistribution(ctx, plan)
	if err != nil {
		resp.Diagnostics.AddError("Could not create node_filedistribution",
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
func (r *nodeFiledistributionResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	// Retrieve values from the plan
	var plan nodeFiledistributionResourceModel
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	err := r.writeFiledistribution(ctx, plan)
	if err != nil {
		resp.Diagnostics.AddError("Could not update node_filedistribution",
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

func (r *nodeFiledistributionResource) writeFiledistribution(ctx context.Context, plan nodeFiledistributionResourceModel) error {
	nodeId := plan.Node.ValueString()
	extend := plan.Extend.ValueBool()

	var files []FiledistributionFile
	diags := plan.Files.ElementsAs(ctx, &files, false)
	if diags.HasError() {
		// Note: this might silence some warnings... Redo at some point.
		return fmt.Errorf("%v: %v", diags.Errors()[0].Summary(), diags.Errors()[0].Detail())
	}

	filesets := PlanToQbeeFilesets(ctx, files)

	// Create the resource
	tflog.Info(ctx, fmt.Sprintf("Creating file distribution for nodeId %v with %v filesets", nodeId, len(filesets)))
	createResponse, err := r.client.Config.CreateNodeFileDistribution(nodeId, filesets, extend)
	if err != nil {
		return fmt.Errorf("error creating a node_filedistribution resource: %w", err)
	}

	_, err = r.client.Configuration.Commit("terraform: create node_filedistribution_resource")
	if err != nil {
		err = fmt.Errorf("error creating a commit for the node_filedistribution: %w", err)

		err = r.client.Configuration.DeleteUncommitted(createResponse.Sha)
		if err != nil {
			err = fmt.Errorf("error deleting uncommitted node_filedistribution changes: %w", err)
		}

		return err
	}

	return nil
}

func (r *nodeFiledistributionResource) createFilesets(ctx context.Context, files []FiledistributionFile) []qbee.FilesetConfig {
	var filesets []qbee.FilesetConfig

	for _, file := range files {
		var fsc qbee.FilesetConfig

		if !file.PreCondition.IsNull() {
			fsc.PreCondition = file.PreCondition.ValueString()
		}

		if !file.Command.IsNull() {
			fsc.Command = file.Command.ValueString()
		}

		var paramValues []FiledistributionParameter
		file.Parameters.ElementsAs(ctx, &paramValues, false)
		var params []qbee.FilesetParameter
		for _, value := range paramValues {
			params = append(params, qbee.FilesetParameter{
				Key:   value.Key.ValueString(),
				Value: value.Value.ValueString(),
			})
		}

		var templateValues []FiledistributionTemplate
		file.Templates.ElementsAs(ctx, &templateValues, false)
		var templates []qbee.FilesetTemplate
		for _, value := range templateValues {
			templates = append(templates, qbee.FilesetTemplate{
				Source:      value.Source.ValueString(),
				Destination: value.Destination.ValueString(),
				IsTemplate:  value.IsTemplate.ValueBool(),
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
func (r *nodeFiledistributionResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	// Get the current state
	var state nodeFiledistributionResourceModel
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	nodeId := state.Node.ValueString()

	// Read the real status
	distribution, err := r.client.Config.GetNodeFiledistribution(nodeId)
	if err != nil {
		resp.Diagnostics.AddError("Could not read node_filedistribution",
			"error reading the node_filedistribution resource: "+err.Error())

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

func filedistributionToListValue(ctx context.Context, filedistribution *qbee.GetFileDistributionResponse, resp *resource.ReadResponse) basetypes.ListValue {
	var files []FiledistributionFile

	for _, file := range filedistribution.Files {
		files = append(files, FiledistributionFile{
			Command:      types.StringValue(file.Command),
			PreCondition: types.StringValue(file.PreCondition),
			Templates:    templatesToListValue(ctx, file.Templates, resp),
			Parameters:   parametersToListValue(ctx, file.Parameters, resp),
		})
	}

	fileValues, diags := types.ListValueFrom(ctx, types.ObjectType{
		AttrTypes: FiledistributionFile{}.attrTypes(),
	}, files)
	resp.Diagnostics.Append(diags...)
	return fileValues
}

func templatesToListValue(ctx context.Context, templates []qbee.FiledistributionFileTemplateResponse, resp *resource.ReadResponse) types.List {
	if templates == nil {
		return types.ListNull(basetypes.ObjectType{
			AttrTypes: FiledistributionTemplate{}.attrTypes(),
		})
	}

	var result []FiledistributionTemplate
	for _, template := range templates {
		result = append(result, FiledistributionTemplate{
			Source:      types.StringValue(template.Source),
			Destination: types.StringValue(template.Destination),
			IsTemplate:  types.BoolValue(template.IsTemplate),
		})
	}

	templatesValue, diags := types.ListValueFrom(ctx, basetypes.ObjectType{AttrTypes: FiledistributionTemplate{}.attrTypes()}, result)
	resp.Diagnostics.Append(diags...)
	return templatesValue
}

func parametersToListValue(ctx context.Context, parameters []qbee.FiledistributionFileParameterResponse, resp *resource.ReadResponse) basetypes.ListValue {
	if parameters == nil {
		return types.ListNull(basetypes.ObjectType{
			AttrTypes: FiledistributionParameter{}.attrTypes(),
		})
	}

	var result []FiledistributionParameter
	for _, parameter := range parameters {
		result = append(result, FiledistributionParameter{
			Key:   types.StringValue(parameter.Key),
			Value: types.StringValue(parameter.Value),
		})
	}

	parametersValue, diags := types.ListValueFrom(ctx, basetypes.ObjectType{AttrTypes: FiledistributionParameter{}.attrTypes()}, result)
	resp.Diagnostics.Append(diags...)
	return parametersValue
}

// Delete deletes the resource and removes the Terraform state on success.
func (r *nodeFiledistributionResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	// Retrieve values from the state
	var state nodeFiledistributionResourceModel
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Delete the resource
	nodeId := state.Node.ValueString()
	tflog.Info(ctx, fmt.Sprintf("Deleting node_filedistribution for node %v", nodeId))
	deleteResponse, err := r.client.Config.ClearNodeFileDistribution(nodeId)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error deleting node_filedistribution",
			"could not delete node filedistribution, unexpected error: "+err.Error())
		return
	}

	_, err = r.client.Configuration.Commit("terraform: create node_filedistribution_resource")
	if err != nil {
		resp.Diagnostics.AddError("Could not commit deletion of node_filedistribution",
			"error creating a commit to delete the node_filedistribution resource: "+err.Error())

		err = r.client.Configuration.DeleteUncommitted(deleteResponse.Sha)
		if err != nil {
			resp.Diagnostics.AddError("Could not revert uncommitted node_filedistribution changes",
				"error deleting uncommitted node_filedistribution changes: "+err.Error())
		}

		return
	}
}

func (r *nodeFiledistributionResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("node"), req, resp)
}
