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
	"go.qbee.io/client"
	"go.qbee.io/client/config"
	"strings"
)

// Ensure the implementation satisfies the expected interfaces.
var (
	_ resource.Resource                     = &softwaremanagementResource{}
	_ resource.ResourceWithConfigure        = &softwaremanagementResource{}
	_ resource.ResourceWithConfigValidators = &softwaremanagementResource{}
	_ resource.ResourceWithImportState      = &softwaremanagementResource{}
)

const (
	errorImportingSoftwaremanagement = "error importing softwaremanagement resource"
	errorWritingSoftwaremanagement   = "error writing softwaremanagement resource"
	errorReadingSoftwaremanagement   = "error reading softwaremanagement resource"
	errorDeletingSoftwaremanagement  = "error deleting softwaremanagement resource"
)

func NewSoftwareManagementResource() resource.Resource {
	return &softwaremanagementResource{}
}

type softwaremanagementResource struct {
	client *client.Client
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

	r.client = req.ProviderData.(*client.Client)
}

type softwareManagementResourceModel struct {
	Node   types.String `tfsdk:"node"`
	Tag    types.String `tfsdk:"tag"`
	Extend types.Bool   `tfsdk:"extend"`
	Items  types.List   `tfsdk:"items"`
}

func (m softwareManagementResourceModel) typeAndIdentifier() (config.EntityType, string) {
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

func (m softwareManagementItemModel) toQbeeSoftwarePackage(ctx context.Context) config.SoftwarePackage {
	var fileModels []softwareManagementItemFile
	m.ConfigFiles.ElementsAs(ctx, &fileModels, false)

	var files []config.ConfigurationFile
	for _, model := range fileModels {
		files = append(files, config.ConfigurationFile{
			ConfigTemplate: model.Template.ValueString(),
			ConfigLocation: model.Location.ValueString(),
		})
	}

	var parameterModels []softwareManagementItemParameter
	m.Parameters.ElementsAs(ctx, &parameterModels, false)

	var parameters []config.ConfigurationFileParameter
	for _, model := range parameterModels {
		parameters = append(parameters, config.ConfigurationFileParameter{
			Key:   model.Key.ValueString(),
			Value: model.Value.ValueString(),
		})
	}

	return config.SoftwarePackage{
		Package:      m.Package.ValueString(),
		ServiceName:  m.ServiceName.ValueString(),
		PreCondition: m.PreCondition.ValueString(),
		ConfigFiles:  files,
		Parameters:   parameters,
	}
}

type softwareManagementItemFile struct {
	Template types.String `tfsdk:"template"`
	Location types.String `tfsdk:"location"`
}

func (f softwareManagementItemFile) attrTypes() map[string]attr.Type {
	return map[string]attr.Type{
		"template": types.StringType,
		"location": types.StringType,
	}
}

type softwareManagementItemParameter struct {
	Key   types.String `tfsdk:"key"`
	Value types.String `tfsdk:"value"`
}

func (p softwareManagementItemParameter) attrTypes() map[string]attr.Type {
	return map[string]attr.Type{
		"key":   types.StringType,
		"value": types.StringType,
	}
}

// Schema defines the schema for the resource.
func (r *softwaremanagementResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
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

	configType, identifier := state.typeAndIdentifier()

	// Read the real status
	activeConfig, err := r.client.GetActiveConfig(ctx, configType, identifier, config.EntityConfigScopeOwn)
	if err != nil {
		resp.Diagnostics.AddError(
			errorReadingSoftwaremanagement,
			"error reading the active configuration: "+err.Error(),
		)

		return
	}

	// Update the current state
	currentSoftwareManagement := activeConfig.BundleData.SoftwareManagement
	if currentSoftwareManagement == nil {
		resp.State.RemoveResource(ctx)
		return
	}

	state.Extend = types.BoolValue(currentSoftwareManagement.Extend)

	var items []softwareManagementItemModel
	for _, softwarePackage := range currentSoftwareManagement.Items {
		value, diags := readSoftwarePackage(ctx, softwarePackage)
		if diags.HasError() {
			resp.Diagnostics.Append(diags...)
			return
		}

		items = append(items, *value)
	}

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
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

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
	tflog.Info(ctx, fmt.Sprintf("Deleting softwaremanagement for %v %v", configType, identifier))

	content := config.SoftwareManagement{
		Metadata: config.Metadata{
			Version: "v1",
			Reset:   true,
		},
	}

	changeRequest, err := createChangeRequest(config.SoftwareManagementBundle, content, configType, identifier)
	if err != nil {
		resp.Diagnostics.AddError(
			errorDeletingSoftwaremanagement,
			err.Error())
		return
	}

	change, err := r.client.CreateConfigurationChange(ctx, changeRequest)
	if err != nil {
		resp.Diagnostics.AddError(
			errorDeletingSoftwaremanagement,
			"could not delete softwaremanagement, unexpected error: "+err.Error(),
		)
		return
	}

	_, err = r.client.CommitConfiguration(ctx, "terraform: create softwaremanagement_resource")
	if err != nil {
		resp.Diagnostics.AddError(
			errorDeletingSoftwaremanagement,
			"error creating a commit to delete the softwaremanagement resource: "+err.Error(),
		)

		err = r.client.DeleteConfigurationChange(ctx, change.SHA)
		if err != nil {
			resp.Diagnostics.AddError(
				errorDeletingSoftwaremanagement,
				"error deleting uncommitted softwaremanagement changes: "+err.Error(),
			)
		}

		return
	}
}

func (r *softwaremanagementResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	configType, identifier, found := strings.Cut(req.ID, ":")
	if !found || configType == "" || identifier == "" {
		resp.Diagnostics.AddError(
			errorImportingSoftwaremanagement,
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
			errorImportingSoftwaremanagement,
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

	var items []config.SoftwarePackage
	for _, model := range itemModels {
		items = append(items, model.toQbeeSoftwarePackage(ctx))
	}

	// Create the resource
	tflog.Info(ctx, fmt.Sprintf("Creating softwaremanagement for %v %v with %v items", configType, identifier, len(items)))

	content := config.SoftwareManagement{
		Metadata: config.Metadata{
			Enabled: true,
			Extend:  extend,
			Version: "v1",
		},
		Items: items,
	}

	changeRequest, err := createChangeRequest(config.SoftwareManagementBundle, content, configType, identifier)
	if err != nil {
		return diag.Diagnostics{
			diag.NewErrorDiagnostic(
				errorWritingSoftwaremanagement,
				err.Error(),
			),
		}
	}

	change, err := r.client.CreateConfigurationChange(ctx, changeRequest)
	if err != nil {
		return diag.Diagnostics{
			diag.NewErrorDiagnostic(
				errorWritingSoftwaremanagement,
				err.Error(),
			),
		}
	}

	_, err = r.client.CommitConfiguration(ctx, "terraform: create softwaremanagement_resource")
	if err != nil {
		diags = diag.Diagnostics{}

		err = fmt.Errorf("error creating a commit for the softwaremanagement: %w", err)
		diags.AddError(errorWritingSoftwaremanagement, err.Error())

		err = r.client.DeleteConfigurationChange(ctx, change.SHA)
		if err != nil {
			diags.AddError(
				errorWritingSoftwaremanagement,
				fmt.Errorf("could not delete uncommitted softwaremanagement changes: %w", err).Error(),
			)
		}

		return diags
	}

	return nil
}

func readSoftwarePackage(ctx context.Context, item config.SoftwarePackage) (*softwareManagementItemModel, diag.Diagnostics) {
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
