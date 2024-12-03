package provider

import (
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
	"github.com/qbee-io/terraform-provider-qbee/internal/qbee"
	"strings"
)

// Ensure the implementation satisfies the expected interfaces.
var (
	_ resource.Resource                     = &softwaremanagementResource{}
	_ resource.ResourceWithConfigure        = &softwaremanagementResource{}
	_ resource.ResourceWithConfigValidators = &softwaremanagementResource{}
	_ resource.ResourceWithImportState      = &softwaremanagementResource{}
)

func NewSoftwareManagementResource() resource.Resource {
	return &softwaremanagementResource{}
}

type softwaremanagementResource struct {
	client *qbee.HttpClient
}

// Metadata returns the resource type name.
func (r *softwaremanagementResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_softwaremanagement"
}

// Configure adds the provider configured client to the resource.
func (r *softwaremanagementResource) Configure(_ context.Context, req resource.ConfigureRequest, _ *resource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}

	r.client = req.ProviderData.(*qbee.HttpClient)
}

type softwareManagementResourceModel struct {
	Node   types.String `tfsdk:"node"`
	Tag    types.String `tfsdk:"tag"`
	ID     types.String `tfsdk:"id"`
	Extend types.Bool   `tfsdk:"extend"`
	Items  types.List   `tfsdk:"items"`
}

func (m softwareManagementResourceModel) typeAndIdentifier() (qbee.ConfigType, string) {
	return typeAndIdentifier(m.Tag, m.Node)
}

func (m softwareManagementResourceModel) identifierToString() string {
	t, i := m.typeAndIdentifier()
	return fmt.Sprintf("%v:%v", t, i)
}

type softwareManagementItemModel struct {
	Package      types.String `tfsdk:"package"`
	ServiceName  types.String `tfsdk:"service_name"`
	PreCondition types.String `tfsdk:"pre_condition"`
	ConfigFiles  types.List   `tfsdk:"config_files"`
	Parameters   types.List   `tfsdk:"parameters"`
}

func (m softwareManagementItemModel) attrTypes() map[string]attr.Type {
	return map[string]attr.Type{
		"package":       types.StringType,
		"service_name":  types.StringType,
		"pre_condition": types.StringType,
		"config_files": types.ListType{
			ElemType: types.ObjectType{
				AttrTypes: softwareManagementItemFile{}.attrTypes(),
			},
		},
		"parameters": types.ListType{
			ElemType: types.ObjectType{
				AttrTypes: softwareManagementItemParameter{}.attrTypes(),
			},
		},
	}
}

type softwareManagementItemFile struct {
	Template types.String `tfsdk:"template"`
	Location types.String `tfsdk:"location"`
}

func (_ softwareManagementItemFile) attrTypes() map[string]attr.Type {
	return map[string]attr.Type{
		"template": types.StringType,
		"location": types.StringType,
	}
}

type softwareManagementItemParameter struct {
	Key   types.String `tfsdk:"key"`
	Value types.String `tfsdk:"value"`
}

func (_ softwareManagementItemParameter) attrTypes() map[string]attr.Type {
	return map[string]attr.Type{
		"key":   types.StringType,
		"value": types.StringType,
	}
}

// Schema defines the schema for the resource.
func (r *softwaremanagementResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
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
			"items": schema.ListNestedAttribute{
				Required:    true,
				Description: "The filesets that must be distributed",
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"package": schema.StringAttribute{
							Required: true,
							Description: "Package name (with .deb) from package in File manager or package name " +
								"(without .deb ending) to install it from a apt repository configured on the device " +
								"(e.g. mc for midnight commander will install from repository)",
						},
						"service_name": schema.StringAttribute{
							Optional: true,
							Description: "Define a service name if it differs from the package name. " +
								"If empty then service name will be assumed to be the same as the package name",
						},
						"pre_condition": schema.StringAttribute{
							Optional: true,
							Description: "Script/executable that needs to return successfully before software " +
								"package is installed. We expect 0 or true. We assume true if left empty. For " +
								"example, call: /bin/true or finish with exit(0)",
						},
						"config_files": schema.ListNestedAttribute{
							Optional: true,
							NestedObject: schema.NestedAttributeObject{
								Attributes: map[string]schema.Attribute{
									"template": schema.StringAttribute{
										Required:    true,
										Description: "The source of the file. Must correspond to a file in the qbee filemanager.",
									},
									"location": schema.StringAttribute{
										Required:    true,
										Description: "The destination of the file on the target device.",
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
					},
				},
			},
		},
	}
}

func (r *softwaremanagementResource) ConfigValidators(ctx context.Context) []resource.ConfigValidator {
	return []resource.ConfigValidator{
		resourcevalidator.ExactlyOneOf(
			path.MatchRoot("tag"),
			path.MatchRoot("node"),
		),
	}
}

// Create creates the resource and sets the initial Terraform state.
func (r *softwaremanagementResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	// Retrieve values from the plan
	var plan softwareManagementResourceModel
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	tflog.Info(ctx, fmt.Sprintf("creating softwaremanagement with identifier %v", plan.identifierToString()))

	diags = r.writeSoftwareManagement(ctx, plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
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
func (r *softwaremanagementResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	// Get the current state
	var state softwareManagementResourceModel
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	tflog.Info(ctx, fmt.Sprintf("reading softwaremanagement with identifier %v", state.identifierToString()))

	// Update the current state
	diags = r.readSoftwareManagement(ctx, &state, resp)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	state.ID = types.StringValue("placeholder")

	resp.State.Set(ctx, state)
}

// Update updates the resource and sets the updated Terraform state on success.
func (r *softwaremanagementResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	// Retrieve values from the plan
	var plan softwareManagementResourceModel
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	tflog.Info(ctx, fmt.Sprintf("updating softwaremanagement with identifier %v", plan.identifierToString()))

	diags = r.writeSoftwareManagement(ctx, plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
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

// Delete deletes the resource and removes the Terraform state on success.
func (r *softwaremanagementResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	// Retrieve values from the state
	var state softwareManagementResourceModel
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Delete the resource
	configType, identifier := typeAndIdentifier(state.Tag, state.Node)
	tflog.Info(ctx, fmt.Sprintf("Deleting softwaremanagement for %v %v", configType.String(), identifier))

	deleteResponse, err := r.client.SoftwareManagement.Clear(configType, identifier)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error deleting softwaremanagement",
			"could not delete softwaremanagement, unexpected error: "+err.Error())
		return
	}

	_, err = r.client.Configuration.Commit("terraform: create filedistribution_resource")
	if err != nil {
		resp.Diagnostics.AddError("Could not commit deletion of softwaremanagement",
			"error creating a commit to delete the softwaremanagement resource: "+err.Error())

		err = r.client.Configuration.DeleteUncommitted(deleteResponse.Sha)
		if err != nil {
			resp.Diagnostics.AddError("Could not revert uncommitted softwaremanagement changes",
				"error deleting uncommitted softwaremanagement changes: "+err.Error())
		}

		return
	}
}

func (r *softwaremanagementResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	configType, identifier, found := strings.Cut(req.ID, ":")
	if !found || configType == "" || identifier == "" {
		resp.Diagnostics.AddError(
			"Unexpected Import Identifier",
			fmt.Sprintf("Expected import identifier with format: type:identifier. Got: %q", req.ID),
		)
		return

	}

	tflog.Info(ctx, fmt.Sprintf("importing softwaremanagement with %v %v", configType, identifier))

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

func (r *softwaremanagementResource) writeSoftwareManagement(ctx context.Context, plan softwareManagementResourceModel) diag.Diagnostics {
	configType, identifier := plan.typeAndIdentifier()
	extend := plan.Extend.ValueBool()

	var itemModels []softwareManagementItemModel
	diags := plan.Items.ElementsAs(ctx, &itemModels, false)
	if diags.HasError() {
		return diags
	}

	var items []qbee.SoftwareManagementItem
	for _, model := range itemModels {
		items = append(items, model.toQbeeItem(ctx))
	}

	// Create the resource
	tflog.Info(ctx, fmt.Sprintf("Creating softwaremanagement for %v %v with %v items", configType.String(), identifier, len(items)))
	createResponse, err := r.client.SoftwareManagement.Create(configType, identifier, items, extend)
	if err != nil {
		return diag.Diagnostics{
			diag.NewErrorDiagnostic(
				"error creating a softwaremanagement resource",
				err.Error(),
			),
		}
	}

	_, err = r.client.Configuration.Commit("terraform: create softwaremanagement_resource")
	if err != nil {
		err = fmt.Errorf("error creating a commit for the softwaremanagement: %w", err)

		err = r.client.Configuration.DeleteUncommitted(createResponse.Sha)
		if err != nil {
			err = fmt.Errorf("error deleting uncommitted softwaremanagement changes: %w", err)
		}

		return diag.Diagnostics{
			diag.NewErrorDiagnostic("error creating a softwaremanagement resource", err.Error()),
		}
	}

	return nil
}

func (m softwareManagementItemModel) toQbeeItem(ctx context.Context) qbee.SoftwareManagementItem {
	var fileModels []softwareManagementItemFile
	m.ConfigFiles.ElementsAs(ctx, &fileModels, false)

	var files []qbee.SoftwareManagementConfigFile
	for _, model := range fileModels {
		files = append(files, model.toQbeeItem())
	}

	var parameterModels []softwareManagementItemParameter
	m.Parameters.ElementsAs(ctx, &parameterModels, false)

	var parameters []qbee.SoftwareManagementParameter
	for _, model := range parameterModels {
		parameters = append(parameters, model.toQbeeItem())
	}

	return qbee.SoftwareManagementItem{
		Package:      m.Package.ValueString(),
		ServiceName:  m.ServiceName.ValueString(),
		PreCondition: m.PreCondition.ValueString(),
		ConfigFiles:  files,
		Parameters:   parameters,
	}
}

func (f softwareManagementItemFile) toQbeeItem() qbee.SoftwareManagementConfigFile {
	return qbee.SoftwareManagementConfigFile{
		ConfigTemplate: f.Template.ValueString(),
		ConfigLocation: f.Location.ValueString(),
	}
}

func (p softwareManagementItemParameter) toQbeeItem() qbee.SoftwareManagementParameter {
	return qbee.SoftwareManagementParameter{
		Key:   p.Key.ValueString(),
		Value: p.Value.ValueString(),
	}
}

func (r *softwaremanagementResource) readSoftwareManagement(ctx context.Context, state *softwareManagementResourceModel, resp *resource.ReadResponse) diag.Diagnostics {
	configType, identifier := state.typeAndIdentifier()

	// Read the real status
	currentState, err := r.client.SoftwareManagement.Get(configType, identifier)
	if err != nil {
		return diag.Diagnostics{
			diag.NewErrorDiagnostic("error reading softwaremanagement state", err.Error()),
		}
	}

	if currentState == nil {
		resp.State.RemoveResource(ctx)
		return nil
	}

	state.Extend = types.BoolValue(currentState.Extend)

	var items []softwareManagementItemModel
	for _, item := range currentState.SoftwareItems {
		value, diags := fromQbeeItem(ctx, item)
		if diags.HasError() {
			return diags
		}

		items = append(items, *value)
	}

	var diags diag.Diagnostics
	if len(items) == 0 {
		state.Items = types.ListNull(
			basetypes.ObjectType{
				AttrTypes: softwareManagementItemModel{}.attrTypes(),
			},
		)
	} else {
		var itemsValue types.List
		itemsValue, diags = types.ListValueFrom(
			ctx,
			basetypes.ObjectType{
				AttrTypes: softwareManagementItemModel{}.attrTypes(),
			},
			items,
		)
		state.Items = itemsValue
	}

	return diags
}

func fromQbeeItem(ctx context.Context, item qbee.SoftwareManagementItem) (*softwareManagementItemModel, diag.Diagnostics) {
	var configFiles []softwareManagementItemFile
	for _, file := range item.ConfigFiles {
		configFiles = append(configFiles, softwareManagementItemFile{
			Template: nullableStringValue(file.ConfigTemplate),
			Location: nullableStringValue(file.ConfigLocation),
		})
	}
	configFilesValue, diags := listFromStructs(ctx, configFiles, basetypes.ObjectType{AttrTypes: softwareManagementItemFile{}.attrTypes()})
	if diags.HasError() {
		return nil, diags
	}

	var parameters []softwareManagementItemParameter
	for _, parameter := range item.Parameters {
		parameters = append(parameters, softwareManagementItemParameter{
			Key:   nullableStringValue(parameter.Key),
			Value: nullableStringValue(parameter.Value),
		})
	}
	parametersValue, diags := listFromStructs(ctx, parameters, basetypes.ObjectType{AttrTypes: softwareManagementItemParameter{}.attrTypes()})
	if diags.HasError() {
		return nil, diags
	}

	return &softwareManagementItemModel{
		Package:      nullableStringValue(item.Package),
		ServiceName:  nullableStringValue(item.ServiceName),
		PreCondition: nullableStringValue(item.PreCondition),
		ConfigFiles:  configFilesValue,
		Parameters:   parametersValue,
	}, nil
}
