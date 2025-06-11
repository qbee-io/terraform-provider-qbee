package provider

import (
	"context"
	"fmt"
	"github.com/hashicorp/terraform-plugin-framework-validators/resourcevalidator"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	"go.qbee.io/client"
	"go.qbee.io/client/config"
	"strings"

	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
)

// Ensure the implementation satisfies the expected interfaces.
var (
	_ resource.Resource                     = &parametersResource{}
	_ resource.ResourceWithConfigure        = &parametersResource{}
	_ resource.ResourceWithConfigValidators = &parametersResource{}
	_ resource.ResourceWithImportState      = &parametersResource{}
)

const (
	errorImportingParameters = "error importing parameters resource"
	errorWritingParameters   = "error writing parameters resource"
	errorReadingParameters   = "error reading parameters resource"
	errorDeletingParameters  = "error deleting parameters resource"
)

// NewParametersResource is a helper function to simplify the provider implementation.
func NewParametersResource() resource.Resource {
	return &parametersResource{}
}

type parametersResource struct {
	client *client.Client
}

type parametersResourceModel struct {
	Node       types.String `tfsdk:"node"`
	Tag        types.String `tfsdk:"tag"`
	Extend     types.Bool   `tfsdk:"extend"`
	Parameters []parameter  `tfsdk:"parameters"`
	Secrets    []secret     `tfsdk:"secrets"`
}

func (m parametersResourceModel) typeAndIdentifier() (config.EntityType, string) {
	return typeAndIdentifier(m.Tag, m.Node)
}

type parameter struct {
	Key   types.String `tfsdk:"key"`
	Value types.String `tfsdk:"value"`
}

type secret struct {
	Key          types.String `tfsdk:"key"`
	Value        types.String `tfsdk:"value_wo"`
	ValueVersion types.String `tfsdk:"value_wo_version"`
	SecretId     types.String `tfsdk:"secret_id"`
}

// Metadata returns the resource type name.
func (r *parametersResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_parameters"
}

// Configure adds the provider configured client to the resource.
func (r *parametersResource) Configure(_ context.Context, req resource.ConfigureRequest, _ *resource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}

	r.client = req.ProviderData.(*client.Client)
}

// Schema defines the schema for the resource.
func (r *parametersResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Parameters sets global configuration parameters.",
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
			"parameters": schema.ListNestedAttribute{
				Optional:    true,
				Description: "Parameters is a list of key/value pairs",
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
			"secrets": schema.ListNestedAttribute{
				Optional:    true,
				Description: "Secrets is a list of key/value pairs where value is write-only",
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"key": schema.StringAttribute{
							Required:    true,
							Description: "The key of the secret.",
						},
						"value_wo": schema.StringAttribute{
							Required:  true,
							Sensitive: true,
							WriteOnly: true,
							Description: "The value of the secret. This value is write-only and will not be stored " +
								"or returned in the state. If you need to overwrite the value, you must change the value_wo_version " +
								"attribute to a new value.",
						},
						"value_wo_version": schema.StringAttribute{
							Required: true,
							Description: "A version of the value_wo. This is used to detect changes in the value_wo. " +
								"If you need to overwrite the value, you must change this value to a new value.",
						},
						"secret_id": schema.StringAttribute{
							Computed: true,
						},
					},
				},
			},
		},
	}
}

func (r *parametersResource) ConfigValidators(ctx context.Context) []resource.ConfigValidator {
	return []resource.ConfigValidator{
		resourcevalidator.ExactlyOneOf(
			path.MatchRoot("tag"),
			path.MatchRoot("node"),
		),
	}
}

// Create creates the resource and sets the initial Terraform state.
func (r *parametersResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	// Retrieve values from the plan
	var plan parametersResourceModel
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	diags = r.writeParameters(ctx, plan)
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

// Update updates the resource and sets the updated Terraform state on success.
func (r *parametersResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	// Retrieve values from the plan
	var plan parametersResourceModel
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	diags = r.writeParameters(ctx, plan)
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
func (r *parametersResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	// Get the current state
	var state *parametersResourceModel
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	configType, identifier := state.typeAndIdentifier()

	// Read the real status
	activeConfig, err := r.client.GetActiveConfig(ctx, configType, identifier, config.EntityConfigScopeOwn)
	if err != nil {
		resp.Diagnostics.AddError(errorReadingParameters,
			"error reading the active configuration: "+err.Error())

		return
	}

	// Update the current state
	currentParameters := activeConfig.BundleData.Parameters
	if currentParameters == nil {
		resp.State.RemoveResource(ctx)
		return
	}

	state.Extend = types.BoolValue(currentParameters.Extend)

	if len(currentParameters.Parameters) > 0 {
		mappedParameters := make([]parameter, len(currentParameters.Parameters))
		for i, p := range currentParameters.Parameters {
			mappedParameters[i] = parameter{
				Key:   types.StringValue(p.Key),
				Value: types.StringValue(p.Value),
			}
		}
		state.Parameters = mappedParameters
	}

	if len(currentParameters.Secrets) > 0 {
		mappedSecrets := make([]secret, len(currentParameters.Secrets))
		for i, s := range currentParameters.Secrets {
			// Find the secret value version in the current state
			valueVersion := ""
			for _, s2 := range state.Secrets {
				if s2.Key.ValueString() == s.Key {
					valueVersion = s2.ValueVersion.ValueString()
					break
				}
			}

			mappedSecrets[i] = secret{
				Key:          types.StringValue(s.Key),
				ValueVersion: types.StringValue(valueVersion),
				SecretId:     types.StringValue(s.Value),
			}
		}
		state.Secrets = mappedSecrets
	}

	diags = resp.State.Set(ctx, state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}

// Delete deletes the resource and removes the Terraform state on success.
func (r *parametersResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	// Retrieve values from the state
	var state parametersResourceModel
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Delete the resource
	configType, identifier := state.typeAndIdentifier()
	tflog.Info(ctx, fmt.Sprintf("Deleting parameters for %v %v", configType, identifier))

	content := config.Parameters{
		Metadata: config.Metadata{
			Reset:   true,
			Version: "v1",
		},
	}

	changeRequest, err := createChangeRequest(config.ParametersBundle, content, configType, identifier)
	if err != nil {
		resp.Diagnostics.AddError(
			errorDeletingParameters,
			err.Error(),
		)
		return
	}

	change, err := r.client.CreateConfigurationChange(ctx, changeRequest)
	if err != nil {
		resp.Diagnostics.AddError(
			errorDeletingParameters,
			err.Error(),
		)
		return
	}

	_, err = r.client.CommitConfiguration(ctx, "terraform: create parameters_resource")
	if err != nil {
		resp.Diagnostics.AddError(errorDeletingParameters,
			"error creating a commit to delete the parameters resource: "+err.Error(),
		)

		err = r.client.DeleteConfigurationChange(ctx, change.SHA)
		if err != nil {
			resp.Diagnostics.AddError(
				errorDeletingParameters,
				"error deleting uncommitted parameters changes: "+err.Error(),
			)
		}

		return
	}
}

// ImportState imports the resource state from the Terraform state.
func (r *parametersResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	configType, identifier, found := strings.Cut(req.ID, ":")
	if !found || configType == "" || identifier == "" {
		resp.Diagnostics.AddError(
			errorImportingParameters,
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
			errorImportingParameters,
			fmt.Sprintf("Import type must be either 'node' or 'tag'. Got: %q", configType),
		)
		return
	}
}

func (r *parametersResource) writeParameters(ctx context.Context, plan parametersResourceModel) diag.Diagnostics {
	configType, identifier := plan.typeAndIdentifier()
	extend := plan.Extend.ValueBool()

	// Create the resource
	tflog.Info(ctx, fmt.Sprintf("Creating parameters for %v %v", configType, identifier))

	mappedParameters := make([]config.Parameter, len(plan.Parameters))
	for i, p := range plan.Parameters {
		mappedParameters[i] = config.Parameter{
			Key:   p.Key.ValueString(),
			Value: p.Value.ValueString(),
		}
	}

	mappedSecrets := make([]config.Parameter, len(plan.Secrets))
	for i, s := range plan.Secrets {
		mappedSecrets[i] = config.Parameter{
			Key:   s.Key.ValueString(),
			Value: s.Value.ValueString(),
		}
	}

	content := config.Parameters{
		Metadata: config.Metadata{
			Enabled: true,
			Extend:  extend,
			Version: "v1",
		},
		Parameters: mappedParameters,
		Secrets:    mappedSecrets,
	}

	changeRequest, err := createChangeRequest(config.ParametersBundle, content, configType, identifier)
	if err != nil {
		return diag.Diagnostics{
			diag.NewErrorDiagnostic(
				errorWritingParameters,
				err.Error(),
			),
		}
	}

	change, err := r.client.CreateConfigurationChange(ctx, changeRequest)
	if err != nil {
		return diag.Diagnostics{
			diag.NewErrorDiagnostic(
				errorWritingParameters,
				fmt.Sprintf("Error creating a parameters resource with qbee: %v", err),
			),
		}
	}

	_, err = r.client.CommitConfiguration(ctx, "terraform: create parameters_resource")
	if err != nil {
		diags := diag.Diagnostics{}

		err = fmt.Errorf("error creating a commit for the parameters: %w", err)
		diags.AddError(errorWritingParameters, err.Error())

		err = r.client.DeleteConfigurationChange(ctx, change.SHA)
		if err != nil {
			diags.AddError(
				errorWritingParameters,
				fmt.Errorf("error deleting uncommitted parameters changes: %w", err).Error(),
			)
		}

		return diags
	}

	// Since the secrets are write-only, we need to get the secret ids from the created configuration
	// so we can detect drift in the future
	if len(plan.Secrets) != 0 {
		createdConfig, ok := change.Content.(map[string]interface{})["config"].(map[string]interface{})
		if !ok {
			return diag.Diagnostics{
				diag.NewErrorDiagnostic(
					errorWritingParameters,
					"error reading the created change content",
				),
			}
		}

		createdSecrets, ok := createdConfig["secrets"].([]interface{})
		if !ok {
			return diag.Diagnostics{
				diag.NewErrorDiagnostic(
					errorWritingParameters,
					"error reading the created secrets: 'secrets' is not a list",
				),
			}
		}

		for i, planSecret := range plan.Secrets {
			var secretId string
			for _, createdSecret := range createdSecrets {
				secretKey, _ := createdSecret.(map[string]interface{})["key"]
				if secretKey == planSecret.Key.ValueString() {
					secretId, _ = createdSecret.(map[string]interface{})["value"].(string)
					break
				}
			}

			if secretId == "" {
				return diag.Diagnostics{
					diag.NewErrorDiagnostic(
						errorWritingParameters,
						"error finding the created secret",
					),
				}
			}

			plan.Secrets[i].SecretId = types.StringValue(secretId)
		}
	}

	return nil
}
