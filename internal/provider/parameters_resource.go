package provider

import (
	"context"
	"crypto/sha256"
	"encoding/binary"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/hashicorp/terraform-plugin-framework-validators/resourcevalidator"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	"go.qbee.io/client"
	"go.qbee.io/client/config"

	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
)

// Ensure the implementation satisfies the expected interfaces.
var (
	_ resource.Resource                     = &parametersResource{}
	_ resource.ResourceWithConfigure        = &parametersResource{}
	_ resource.ResourceWithConfigValidators = &parametersResource{}
	_ resource.ResourceWithImportState      = &parametersResource{}
	_ resource.ResourceWithModifyPlan       = &parametersResource{}
)

const (
	errorImportingParameters = "error importing parameters resource"
	errorWritingParameters   = "error writing parameters resource"
	errorReadingParameters   = "error reading parameters resource"
	errorDeletingParameters  = "error deleting parameters resource"
	privateStateKey          = "private_state"
)

// NewParametersResource is a helper function to simplify the provider implementation.
func NewParametersResource() resource.Resource {
	return &parametersResource{}
}

type parametersResource struct {
	client *client.Client
}

// We use private state to keep a hash of the secret values we write.
// If, during a plan, we detect that the hash of the current values differs, we know
// that the planned change to SecretsHash should be "Unknown", and we can set it to
// what qbee returned after the apply.
type privateStateModel struct {
	// SecretsWoValuesHash is the SHA-256 hash of the write-only secret input values (secrets_wo)
	// from the Terraform configuration. It is stored in private state so we can detect when the
	// user-provided secret values changed and force SecretsHash to Unknown during planning.
	SecretsWoValuesHash string `json:"secrets_wo_values_hash"`

	// QbeeSecretIdsHash is the SHA-256 hash of the secrets as returned by qbee after the last
	// successful write. It is used for drift detection by comparing it with the current SecretsHash
	// derived from the remote state during Read/ModifyPlan.
	QbeeSecretIdsHash string `json:"qbee_secret_ids_hash"`
}

type parametersResourceModel struct {
	Node             types.String `tfsdk:"node"`
	Tag              types.String `tfsdk:"tag"`
	Extend           types.Bool   `tfsdk:"extend"`
	Parameters       []parameter  `tfsdk:"parameters"`
	SecretsWo        []secret     `tfsdk:"secrets_wo"`
	SecretsWoVersion types.Int64  `tfsdk:"secrets_wo_version"`
	// SecretsHash contains the hash based on the secret id's in Qbee.
	SecretsHash types.String `tfsdk:"secrets_hash"`
}

func (m parametersResourceModel) typeAndIdentifier() (config.EntityType, string) {
	return typeAndIdentifier(m.Tag, m.Node)
}

type parameter struct {
	Key   types.String `tfsdk:"key"`
	Value types.String `tfsdk:"value"`
}

type secret struct {
	Key   types.String `tfsdk:"key"`
	Value types.String `tfsdk:"value"`
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
			"secrets_wo": schema.ListNestedAttribute{
				Optional:    true,
				Description: "A write-only list of key/value pairs.",
				WriteOnly:   true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"key": schema.StringAttribute{
							Required:    true,
							Description: "The key of the secret.",
							WriteOnly:   true,
						},
						"value": schema.StringAttribute{
							Required:  true,
							Sensitive: true,
							WriteOnly: true,
							Description: "The value of the secret. This value is write-only and will not be stored " +
								"or returned in the state.",
						},
					},
				},
			},
			"secrets_wo_version": schema.Int64Attribute{
				Optional:    true,
				Description: "Optional version for secrets_wo. If set, secrets are only rewritten when this version changes.",
			},
			"secrets_hash": schema.StringAttribute{
				Computed:    true,
				Description: "A computed hash based on secret IDs from qbee. This value changes when secrets are updated and is used to detect drift in remote secret values.",
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
	// Retrieve values from the plan.
	// Combine plan and configuration, since write-only values are not in the plan.
	var effective parametersResourceModel
	diags := req.Plan.Get(ctx, &effective)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	var configuration parametersResourceModel
	diags = req.Config.Get(ctx, &configuration)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	effective.SecretsWo = configuration.SecretsWo
	effective.SecretsWoVersion = configuration.SecretsWoVersion

	private, diags := r.writeParameters(ctx, &effective, privateStateModel{})
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	diags = resp.Private.SetKey(ctx, privateStateKey, private)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Map response body to schema and populate Computed attribute values
	// Set state to fully populated data
	var state parametersResourceModel
	state.Extend = effective.Extend
	state.Node = effective.Node
	state.Tag = effective.Tag
	state.Parameters = effective.Parameters
	state.SecretsHash = effective.SecretsHash
	state.SecretsWoVersion = effective.SecretsWoVersion

	diags = resp.State.Set(ctx, state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}

// Update updates the resource and sets the updated Terraform state on success.
func (r *parametersResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	// Retrieve values from the plan.
	// Combine plan and configuration, since write-only values are not in the plan.
	var plan parametersResourceModel
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	var configuration parametersResourceModel
	diags = req.Config.Get(ctx, &configuration)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	effective := plan
	// Only copy SecretsWo from configuration. Keep SecretsWoVersion from the plan.
	effective.SecretsWo = configuration.SecretsWo

	privateStateBytes, diags := req.Private.GetKey(ctx, privateStateKey)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	initialPrivateState, diags := unmarshalPrivateState(privateStateBytes)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	privateState, diags := r.writeParameters(ctx, &effective, initialPrivateState)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	diags = resp.Private.SetKey(ctx, privateStateKey, privateState)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Map response body to schema and populate Computed attribute values
	// Set state to fully populated data
	var state parametersResourceModel
	state.Extend = effective.Extend
	state.Node = effective.Node
	state.Tag = effective.Tag
	state.Parameters = effective.Parameters
	state.SecretsWoVersion = effective.SecretsWoVersion
	state.SecretsHash = effective.SecretsHash

	diags = resp.State.Set(ctx, state)
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

	// Parameters can be mapped directly
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

	// Secrets are not read themselves, but we decide the SecretsHash from it
	if len(currentParameters.Secrets) == 0 {
		// We can safely determine that the SecretsWoValuesHash should be null if there currently are no secrets.
		state.SecretsHash = types.StringNull()
	} else {
		currentHash := computeSecretsHash(currentParameters.Secrets)
		state.SecretsHash = types.StringValue(currentHash)
	}

	// Write the state
	diags = resp.State.Set(ctx, state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}

func (r *parametersResource) ModifyPlan(ctx context.Context, req resource.ModifyPlanRequest, resp *resource.ModifyPlanResponse) {
	// If either Plan or State is null, nothing to do (no resource instance)
	if req.Plan.Raw.IsNull() || req.State.Raw.IsNull() {
		return
	}

	// Load the current plan
	var plan *parametersResourceModel
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Load the current state
	var state parametersResourceModel
	diags = req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Load configuration to access write-only secrets
	var configuration parametersResourceModel
	diags = req.Config.Get(ctx, &configuration)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// If a secrets_wo_version is specified, only update when the version changes.
	// We would trigger a change by setting SecretsHash to Unknown, causing Terraform to trigger an update.
	if !configuration.SecretsWoVersion.IsNull() {
		newVersion := configuration.SecretsWoVersion.ValueInt64()

		// If the Version changed (either from null to any value, or from a value to another value),
		// set the hash to "Unknown" to trigger a resource update in the plan.
		if state.SecretsWoVersion.IsNull() || state.SecretsWoVersion.ValueInt64() != newVersion {
			plan.SecretsHash = types.StringUnknown()
		} else {
			plan.SecretsHash = state.SecretsHash
		}
	} else if len(configuration.SecretsWo) > 0 {
		// No SecretsWoVersion: Check if our inputs changed by comparing privateState.SecretsWoValueHash
		secrets := make([]config.Parameter, len(configuration.SecretsWo))
		for i, s := range configuration.SecretsWo {
			secrets[i] = config.Parameter{
				Key:   s.Key.ValueString(),
				Value: s.Value.ValueString(),
			}
		}

		currentHash := computeSecretsHash(secrets)

		privateStateBytes, diags := req.Private.GetKey(ctx, privateStateKey)
		resp.Diagnostics.Append(diags...)
		if resp.Diagnostics.HasError() {
			return
		}

		privateState, diags := unmarshalPrivateState(privateStateBytes)
		resp.Diagnostics.Append(diags...)
		if resp.Diagnostics.HasError() {
			return
		}

		inputValuesChanged := privateState.SecretsWoValuesHash != currentHash

		// State was already updated by Read; We can compare that to what it was last time we wrote,
		// which we stored in the private state.
		remoteStateDrifted := privateState.QbeeSecretIdsHash != state.SecretsHash.ValueString()

		if inputValuesChanged || remoteStateDrifted {
			plan.SecretsHash = types.StringUnknown()
		} else {
			plan.SecretsHash = state.SecretsHash
		}
	} else if !state.SecretsHash.IsNull() {
		// No SecretsWoVersion and no SecretsWo: SecretsHash is not null, which means we used to
		// have secrets. Set to Unknown so we trigger an update which will clear them in qbee.
		plan.SecretsHash = types.StringUnknown()
	} else {
		// No SecretsWoVersion and no SecretsWo: SecretsHash is null, so we don't need to update.
		plan.SecretsHash = types.StringNull()
	}

	diags = resp.Plan.Set(ctx, &plan)
	resp.Diagnostics.Append(diags...)
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

func (r *parametersResource) writeParameters(ctx context.Context, effective *parametersResourceModel, privateState privateStateModel) ([]byte, diag.Diagnostics) {
	configType, identifier := effective.typeAndIdentifier()
	extend := effective.Extend.ValueBool()

	// Create the resource
	tflog.Info(ctx, fmt.Sprintf("Creating parameters for %v %v", configType, identifier))

	mappedParameters := make([]config.Parameter, len(effective.Parameters))
	for i, p := range effective.Parameters {
		mappedParameters[i] = config.Parameter{
			Key:   p.Key.ValueString(),
			Value: p.Value.ValueString(),
		}
	}

	secretsToWrite := make([]config.Parameter, 0)
	secretValuesHash := ""
	if effective.SecretsHash.IsUnknown() {
		secretsToWrite = make([]config.Parameter, len(effective.SecretsWo))
		// We detected a reason to change the write-only secret during ModifyPlan
		for i, s := range effective.SecretsWo {
			secretsToWrite[i] = config.Parameter{
				Key:   s.Key.ValueString(),
				Value: s.Value.ValueString(),
			}
		}
		secretValuesHash = computeSecretsHash(secretsToWrite)
	} else {
		// We should not change, so copy the existing values
		activeConfig, err := r.client.GetActiveConfig(ctx, configType, identifier, config.EntityConfigScopeOwn)
		if err != nil {
			return nil, diag.Diagnostics{
				diag.NewErrorDiagnostic(
					errorReadingParameters,
					"error reading the active configuration: "+err.Error(),
				),
			}
		}

		currentParameters := activeConfig.BundleData.Parameters
		secretsToWrite = make([]config.Parameter, len(currentParameters.Secrets))
		for i, s := range currentParameters.Secrets {
			secretsToWrite[i] = config.Parameter{
				Key:   s.Key,
				Value: s.Value,
			}
		}

		secretValuesHash = privateState.SecretsWoValuesHash
	}

	content := config.Parameters{
		Metadata: config.Metadata{
			Enabled: true,
			Extend:  extend,
			Version: "v1",
		},
		Parameters: mappedParameters,
		Secrets:    secretsToWrite,
	}

	changeRequest, err := createChangeRequest(config.ParametersBundle, content, configType, identifier)
	if err != nil {
		return nil, diag.Diagnostics{
			diag.NewErrorDiagnostic(
				errorWritingParameters,
				err.Error(),
			),
		}
	}

	change, err := r.client.CreateConfigurationChange(ctx, changeRequest)
	if err != nil {
		return nil, diag.Diagnostics{
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

		return nil, diags
	}

	if len(secretsToWrite) == 0 {
		effective.SecretsHash = types.StringNull()
		return nil, nil
	}

	createdConfig, ok := change.Content.(map[string]interface{})["config"].(map[string]interface{})
	if !ok {
		return nil, diag.Diagnostics{
			diag.NewErrorDiagnostic(
				errorWritingParameters,
				"error reading the created change content",
			),
		}
	}

	createdSecretsStruct, ok := createdConfig["secrets"].([]interface{})
	if !ok {
		return nil, diag.Diagnostics{
			diag.NewErrorDiagnostic(
				errorWritingParameters,
				fmt.Sprintf("error reading the created secrets: could not read %v", createdConfig["secrets"]),
			),
		}
	}

	createdSecretsResponse := make([]config.Parameter, len(createdSecretsStruct))
	for i, createdSecret := range createdSecretsStruct {
		createdSecretsResponse[i] = config.Parameter{
			Key:   createdSecret.(map[string]interface{})["key"].(string),
			Value: createdSecret.(map[string]interface{})["value"].(string),
		}
	}

	// Secrets hash is based on the secret id's from qbee
	qbeeSecretIdsHash := computeSecretsHash(createdSecretsResponse)
	effective.SecretsHash = types.StringValue(qbeeSecretIdsHash)

	privateStateJson, err := marshalPrivateState(secretValuesHash, qbeeSecretIdsHash)
	if err != nil {
		return nil, diag.Diagnostics{
			diag.NewErrorDiagnostic(errorReadingParameters,
				fmt.Sprintf("error marshaling private state: %v", err),
			),
		}
	}

	return privateStateJson, nil
}

func marshalPrivateState(secretValuesHash string, qbeeSecretIdsHash string) ([]byte, error) {
	state := privateStateModel{
		SecretsWoValuesHash: secretValuesHash,
		QbeeSecretIdsHash:   qbeeSecretIdsHash,
	}

	j, err := json.Marshal(state)
	if err != nil {
		return nil, fmt.Errorf("could not marshal private state: %w", err)
	}

	return j, nil
}

func unmarshalPrivateState(private []byte) (privateStateModel, diag.Diagnostics) {
	if private == nil {
		return privateStateModel{}, nil
	}

	var p privateStateModel
	err := json.Unmarshal(private, &p)
	if err != nil {
		return privateStateModel{}, diag.Diagnostics{
			diag.NewErrorDiagnostic(
				errorReadingParameters,
				"error reading the private state: "+err.Error())}
	}

	return p, nil
}

func computeSecretsHash(secrets []config.Parameter) string {

	// Single growing buffer; avoids per-field []byte(string) allocations.
	// Capacity guess is optional; it just reduces re-allocations for typical sizes.
	buf := make([]byte, 0, len(secrets)*32)

	for _, s := range secrets {
		// Key
		buf = binary.AppendUvarint(buf, uint64(len(s.Key)))
		buf = append(buf, s.Key...)

		// Value
		buf = binary.AppendUvarint(buf, uint64(len(s.Value)))
		buf = append(buf, s.Value...)
	}

	sum := sha256.Sum256(buf)
	return hex.EncodeToString(sum[:])
}
