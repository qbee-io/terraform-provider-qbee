package provider

import (
	"bitbucket.org/booqsoftware/terraform-provider-qbee/internal/qbee"
	"context"
	"fmt"
	"github.com/hashicorp/terraform-plugin-framework-validators/resourcevalidator"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	"strings"
)

// Ensure the implementation satisfies the expected interfaces.
var (
	_ resource.Resource                     = &filedistributionResource{}
	_ resource.ResourceWithConfigure        = &filedistributionResource{}
	_ resource.ResourceWithConfigValidators = &filedistributionResource{}
	_ resource.ResourceWithImportState      = &filedistributionResource{}
)

func NewFiledistributionResource() resource.Resource {
	return &filedistributionResource{}
}

type filedistributionResource struct {
	client *qbee.HttpClient
}

// Metadata returns the resource type name.
func (r *filedistributionResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_filedistribution"
}

// Configure adds the provider configured client to the resource.
func (r *filedistributionResource) Configure(_ context.Context, req resource.ConfigureRequest, _ *resource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}

	r.client = req.ProviderData.(*qbee.HttpClient)
}

// Schema defines the schema for the resource.
func (r *filedistributionResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed:    true,
				Description: "Placeholder ID value",
			},
			"tag": schema.StringAttribute{
				Optional:      true,
				Description:   "The tag for which to set the configuration. Either tag or node is required.",
				PlanModifiers: []planmodifier.String{stringplanmodifier.RequiresReplace()},
			},
			"node": schema.StringAttribute{
				Optional:      true,
				Description:   "The node for which to set the configuration. Either tag or node is required.",
				PlanModifiers: []planmodifier.String{stringplanmodifier.RequiresReplace()},
			},
			"extend": schema.BoolAttribute{
				Required: true,
				Description: "If the configuration should extend configuration from the parent nodes of the node " +
					"the configuration is applied to. If set to false, configuration from parent nodes is ignored.",
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

func (r *filedistributionResource) ConfigValidators(ctx context.Context) []resource.ConfigValidator {
	return []resource.ConfigValidator{
		resourcevalidator.ExactlyOneOf(
			path.MatchRoot("tag"),
			path.MatchRoot("node"),
		),
	}
}

type filedistributionResourceModel struct {
	Node   types.String `tfsdk:"node"`
	Tag    types.String `tfsdk:"tag"`
	ID     types.String `tfsdk:"id"`
	Extend types.Bool   `tfsdk:"extend"`
	Files  types.List   `tfsdk:"files"`
}

func (m filedistributionResourceModel) typeAndIdentifier() (qbee.ConfigType, string) {
	return typeAndIdentifier(m.Tag, m.Node)
}

// Create creates the resource and sets the initial Terraform state.
func (r *filedistributionResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	// Retrieve values from the plan
	var plan filedistributionResourceModel
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	err := r.writeFiledistribution(ctx, plan)
	if err != nil {
		resp.Diagnostics.AddError("Could not create filedistribution",
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
func (r *filedistributionResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	// Retrieve values from the plan
	var plan filedistributionResourceModel
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	err := r.writeFiledistribution(ctx, plan)
	if err != nil {
		resp.Diagnostics.AddError("Could not update filedistribution",
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

// Read refreshes the Terraform state with the latest data.
func (r *filedistributionResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	// Get the current state
	var state filedistributionResourceModel
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	configType, identifier := state.typeAndIdentifier()

	// Read the real status
	currentFiledistribution, err := r.client.FileDistribution.Get(configType, identifier)
	if err != nil {
		resp.Diagnostics.AddError("Could not read filedistribution",
			"error reading the filedistribution resource: "+err.Error())

		return
	}

	if currentFiledistribution == nil {
		resp.State.RemoveResource(ctx)
		return
	}

	// Update the current state
	fileValues := filedistributionToListValue(ctx, currentFiledistribution, resp)
	if resp.Diagnostics.HasError() {
		return
	}

	state.ID = types.StringValue("placeholder")
	state.Extend = types.BoolValue(currentFiledistribution.Extend)
	state.Files = *fileValues

	resp.State.Set(ctx, state)
}

// Delete deletes the resource and removes the Terraform state on success.
func (r *filedistributionResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	// Retrieve values from the state
	var state filedistributionResourceModel
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Delete the resource
	configType, identifier := state.typeAndIdentifier()
	tflog.Info(ctx, fmt.Sprintf("Deleting filedistribution for %v %v", configType.String(), identifier))

	deleteResponse, err := r.client.FileDistribution.Clear(configType, identifier)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error deleting filedistribution",
			"could not delete filedistribution, unexpected error: "+err.Error())
		return
	}

	_, err = r.client.Configuration.Commit("terraform: create filedistribution_resource")
	if err != nil {
		resp.Diagnostics.AddError("Could not commit deletion of filedistribution",
			"error creating a commit to delete the filedistribution resource: "+err.Error())

		err = r.client.Configuration.DeleteUncommitted(deleteResponse.Sha)
		if err != nil {
			resp.Diagnostics.AddError("Could not revert uncommitted filedistribution changes",
				"error deleting uncommitted filedistribution changes: "+err.Error())
		}

		return
	}
}

func (r *filedistributionResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	configType, identifier, found := strings.Cut(req.ID, ":")
	if !found || configType == "" || identifier == "" {
		resp.Diagnostics.AddError(
			"Unexpected Import Identifier",
			fmt.Sprintf("Expected import identifier with format: type:identifier. Got: %q", req.ID),
		)
		return

	}

	if configType == "tag" {
		resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("tag"), identifier)...)
	} else if configType == "node" {
		resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("node"), identifier)...)
	} else {
		resp.Diagnostics.AddError(
			"Unexpected Import Identifier",
			fmt.Sprintf("Import type must be either 'node' or 'tag'. Got: %q", configType),
		)
		return
	}
}

func (r *filedistributionResource) writeFiledistribution(ctx context.Context, plan filedistributionResourceModel) error {
	configType, identifier := plan.typeAndIdentifier()
	extend := plan.Extend.ValueBool()

	var files []filedistributionFile
	diags := plan.Files.ElementsAs(ctx, &files, false)
	if diags.HasError() {
		// Note: this might silence some warnings... Redo at some point.
		return fmt.Errorf("%v: %v", diags.Errors()[0].Summary(), diags.Errors()[0].Detail())
	}

	filesets := planToQbeeFilesets(ctx, files)

	// Create the resource
	tflog.Info(ctx, fmt.Sprintf("Creating file distribution for %v %v with %v filesets", configType.String(), identifier, len(filesets)))
	createResponse, err := r.client.FileDistribution.Create(configType, identifier, filesets, extend)
	if err != nil {
		return fmt.Errorf("error creating a filedistribution resource: %w", err)
	}

	_, err = r.client.Configuration.Commit("terraform: create filedistribution_resource")
	if err != nil {
		err = fmt.Errorf("error creating a commit for the filedistribution: %w", err)

		err = r.client.Configuration.DeleteUncommitted(createResponse.Sha)
		if err != nil {
			err = fmt.Errorf("error deleting uncommitted filedistribution changes: %w", err)
		}

		return err
	}

	return nil
}

func planToQbeeFilesets(ctx context.Context, files []filedistributionFile) []qbee.FiledistributionFile {
	var filesets []qbee.FiledistributionFile

	for _, file := range files {
		var paramValues []filedistributionParameter
		file.Parameters.ElementsAs(ctx, &paramValues, false)
		var params []qbee.FiledistributionParameter
		for _, value := range paramValues {
			params = append(params, qbee.FiledistributionParameter{
				Key:   value.Key.ValueString(),
				Value: value.Value.ValueString(),
			})
		}

		var templateValues []filedistributionTemplate
		file.Templates.ElementsAs(ctx, &templateValues, false)
		var templates []qbee.FiledistributionTemplate
		for _, value := range templateValues {
			templates = append(templates, qbee.FiledistributionTemplate{
				Source:      value.Source.ValueString(),
				Destination: value.Destination.ValueString(),
				IsTemplate:  value.IsTemplate.ValueBool(),
			})
		}

		filesets = append(filesets, qbee.FiledistributionFile{
			PreCondition: file.PreCondition.ValueString(),
			Command:      file.Command.ValueString(),
			Templates:    templates,
			Parameters:   params,
		})
	}

	return filesets
}

func filedistributionToListValue(ctx context.Context, filedistribution *qbee.FileDistribution, resp *resource.ReadResponse) *basetypes.ListValue {
	var files []filedistributionFile

	for _, file := range filedistribution.Files {
		value, diags := fromQbeeFile(ctx, file)
		if diags.HasError() {
			resp.Diagnostics.Append(diags...)
			return nil
		}

		files = append(files, *value)
	}

	fileValues, diags := types.ListValueFrom(ctx, types.ObjectType{
		AttrTypes: filedistributionFile{}.attrTypes(),
	}, files)
	resp.Diagnostics.Append(diags...)
	return &fileValues
}

func fromQbeeFile(ctx context.Context, item qbee.FiledistributionFile) (*filedistributionFile, diag.Diagnostics) {
	var templates []filedistributionTemplate
	for _, template := range item.Templates {
		templates = append(templates, filedistributionTemplate{
			Source:      nullableStringValue(template.Source),
			Destination: nullableStringValue(template.Destination),
			IsTemplate:  types.BoolValue(template.IsTemplate),
		})
	}
	templatesValue, diags := listFromStructs(ctx, templates, basetypes.ObjectType{AttrTypes: filedistributionTemplate{}.attrTypes()})
	if diags.HasError() {
		return nil, diags
	}

	var parameters []filedistributionParameter
	for _, parameter := range item.Parameters {
		parameters = append(parameters, filedistributionParameter{
			Key:   nullableStringValue(parameter.Key),
			Value: nullableStringValue(parameter.Value),
		})
	}
	parametersValue, diags := listFromStructs(ctx, parameters, basetypes.ObjectType{AttrTypes: filedistributionParameter{}.attrTypes()})
	if diags.HasError() {
		return nil, diags
	}

	return &filedistributionFile{
		Command:      nullableStringValue(item.Command),
		PreCondition: nullableStringValue(item.PreCondition),
		Templates:    templatesValue,
		Parameters:   parametersValue,
	}, nil
}

type filedistributionTemplate struct {
	Source      types.String `tfsdk:"source"`
	Destination types.String `tfsdk:"destination"`
	IsTemplate  types.Bool   `tfsdk:"is_template"`
}

func (f filedistributionTemplate) attrTypes() map[string]attr.Type {
	return map[string]attr.Type{
		"source":      types.StringType,
		"destination": types.StringType,
		"is_template": types.BoolType,
	}
}

type filedistributionFile struct {
	Command      types.String `tfsdk:"command"`
	PreCondition types.String `tfsdk:"pre_condition"`
	Templates    types.List   `tfsdk:"templates"`
	Parameters   types.List   `tfsdk:"parameters"`
}

func (f filedistributionFile) attrTypes() map[string]attr.Type {
	return map[string]attr.Type{
		"command":       types.StringType,
		"pre_condition": types.StringType,
		"templates": types.ListType{
			ElemType: types.ObjectType{
				AttrTypes: filedistributionTemplate{}.attrTypes(),
			},
		},
		"parameters": types.ListType{
			ElemType: types.ObjectType{
				AttrTypes: filedistributionParameter{}.attrTypes(),
			},
		},
	}
}

type filedistributionParameter struct {
	Key   types.String `tfsdk:"key"`
	Value types.String `tfsdk:"value"`
}

func (f filedistributionParameter) attrTypes() map[string]attr.Type {
	return map[string]attr.Type{
		"key":   types.StringType,
		"value": types.StringType,
	}
}
