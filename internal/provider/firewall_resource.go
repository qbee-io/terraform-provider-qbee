package provider

import (
	"context"
	"fmt"
	"github.com/hashicorp/terraform-plugin-framework-validators/resourcevalidator"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	"go.qbee.io/client"
	"go.qbee.io/client/config"
	"strings"
)

// Ensure the implementation satisfies the expected interfaces.
var (
	_ resource.Resource                     = &firewallResource{}
	_ resource.ResourceWithConfigure        = &firewallResource{}
	_ resource.ResourceWithConfigValidators = &firewallResource{}
	_ resource.ResourceWithImportState      = &firewallResource{}
)

const (
	errorImportingFirewall = "error importing firewall resource"
	errorWritingFirewall   = "error writing firewall resource"
	errorReadingFirewall   = "error reading firewall resource"
	errorDeletingFirewall  = "error deleting firewall resource"
)

func NewFirewallResource() resource.Resource {
	return &firewallResource{}
}

type firewallResource struct {
	client *client.Client
}

// Metadata returns the resource type name.
func (r *firewallResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_firewall"
}

// Configure adds the provider configured client to the resource.
func (r *firewallResource) Configure(_ context.Context, req resource.ConfigureRequest, _ *resource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}

	r.client = req.ProviderData.(*client.Client)
}

// Schema defines the schema for the resource.
func (r *firewallResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
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
			"input": schema.SingleNestedAttribute{
				Required:    true,
				Description: "The definition of the firewall configuration.",
				Attributes: map[string]schema.Attribute{
					"policy": schema.StringAttribute{
						Required:    true,
						Description: "The default policy. Either DROP or ACCEPT.",
						Validators: []validator.String{
							stringvalidator.OneOf("DROP", "ACCEPT"),
						},
					},
					"rules": schema.ListNestedAttribute{
						Required: true,
						NestedObject: schema.NestedAttributeObject{
							Attributes: map[string]schema.Attribute{
								"proto": schema.StringAttribute{
									Required:    true,
									Description: "The protocol to match. Either UDP or TCP.",
									Validators: []validator.String{
										stringvalidator.OneOf("UDP", "TCP"),
									},
								},
								"target": schema.StringAttribute{
									Required:    true,
									Description: "The action to take when this rule is matched. Either DROP or ACCEPT.",
									Validators: []validator.String{
										stringvalidator.OneOf("DROP", "ACCEPT"),
									},
								},
								"src_ip": schema.StringAttribute{
									Required:    true,
									Description: "The source ip to match.",
								},
								"dst_port": schema.StringAttribute{
									Required:    true,
									Description: "The destination port to match.",
								},
							},
						},
					},
				},
			},
		},
	}
}

func (r *firewallResource) ConfigValidators(ctx context.Context) []resource.ConfigValidator {
	return []resource.ConfigValidator{
		resourcevalidator.ExactlyOneOf(
			path.MatchRoot("tag"),
			path.MatchRoot("node"),
		),
	}
}

type firewallResourceModel struct {
	Node   types.String   `tfsdk:"node"`
	Tag    types.String   `tfsdk:"tag"`
	Extend types.Bool     `tfsdk:"extend"`
	Input  *firewallInput `tfsdk:"input"`
}

func (m firewallResourceModel) typeAndIdentifier() (config.EntityType, string) {
	return typeAndIdentifier(m.Tag, m.Node)
}

func (i firewallInput) toQbeeFirewallChain() config.FirewallChain {
	var rules []config.FirewallRule
	for _, rule := range i.Rules {
		rules = append(rules, config.FirewallRule{
			Protocol:        config.Protocol(rule.Proto.ValueString()),
			Target:          config.Target(rule.Target.ValueString()),
			SourceIP:        rule.SrcIp.ValueString(),
			DestinationPort: rule.DstPort.ValueString(),
		})
	}
	return config.FirewallChain{
		Policy: config.Target(i.Policy.ValueString()),
		Rules:  rules,
	}
}

type firewallInput struct {
	Policy types.String   `tfsdk:"policy"`
	Rules  []firewallRule `tfsdk:"rules"`
}

type firewallRule struct {
	Proto   types.String `tfsdk:"proto"`
	Target  types.String `tfsdk:"target"`
	SrcIp   types.String `tfsdk:"src_ip"`
	DstPort types.String `tfsdk:"dst_port"`
}

// Create creates the resource and sets the initial Terraform state.
func (r *firewallResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	// Retrieve values from the plan
	var plan firewallResourceModel
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	diags = r.writeFirewall(ctx, plan)
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
func (r *firewallResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	// Get the current state
	var state *firewallResourceModel
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	configType, identifier := state.typeAndIdentifier()

	// Read the real status
	activeConfig, err := r.client.GetActiveConfig(ctx, configType, identifier, config.EntityConfigScopeOwn)
	if err != nil {
		resp.Diagnostics.AddError(
			errorReadingFirewall,
			"error reading the active configuration: "+err.Error(),
		)

		return
	}

	// Update the current state
	currentFirewall := activeConfig.BundleData.Firewall
	if currentFirewall == nil {
		resp.State.RemoveResource(ctx)
		return
	}

	var inputRules []firewallRule
	inputChain := currentFirewall.Tables[config.Filter][config.Input]
	for _, rule := range inputChain.Rules {
		inputRules = append(inputRules, firewallRule{
			Proto:   types.StringValue(string(rule.Protocol)),
			Target:  types.StringValue(string(rule.Target)),
			SrcIp:   types.StringValue(rule.SourceIP),
			DstPort: types.StringValue(rule.DestinationPort),
		})
	}
	input := firewallInput{
		Policy: types.StringValue(string(inputChain.Policy)),
		Rules:  inputRules,
	}

	state.Extend = types.BoolValue(currentFirewall.Extend)
	state.Input = &input

	diags = resp.State.Set(ctx, state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}

// Update updates the resource and sets the updated Terraform state on success.
func (r *firewallResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	// Retrieve values from the plan
	var plan firewallResourceModel
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	diags = r.writeFirewall(ctx, plan)
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
func (r *firewallResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	// Retrieve values from the state
	var state firewallResourceModel
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Delete the resource
	configType, identifier := state.typeAndIdentifier()
	tflog.Info(ctx, fmt.Sprintf("Deleting firewall for %v %v", configType, identifier))

	content := config.Firewall{
		Metadata: config.Metadata{
			Reset:   true,
			Version: "v1",
		},
	}

	changeRequest, err := createChangeRequest(config.FirewallBundle, content, configType, identifier)
	if err != nil {
		resp.Diagnostics.AddError(
			errorDeletingFirewall,
			err.Error(),
		)
		return
	}

	change, err := r.client.CreateConfigurationChange(ctx, changeRequest)
	if err != nil {
		resp.Diagnostics.AddError(
			errorDeletingFirewall,
			"could not delete firewall, unexpected error: "+err.Error())
		return
	}

	_, err = r.client.CommitConfiguration(ctx, "terraform: create firewall_resource")
	if err != nil {
		resp.Diagnostics.AddError(
			errorDeletingFirewall,
			"error creating a commit to delete the firewall resource: "+err.Error(),
		)

		err = r.client.DeleteConfigurationChange(ctx, change.SHA)
		if err != nil {
			resp.Diagnostics.AddError(
				errorDeletingFirewall,
				"error deleting uncommitted firewall changes: "+err.Error(),
			)
		}

		return
	}
}

func (r *firewallResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	configType, identifier, found := strings.Cut(req.ID, ":")
	if !found || configType == "" || identifier == "" {
		resp.Diagnostics.AddError(
			errorImportingFirewall,
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
			errorImportingFirewall,
			fmt.Sprintf("Import type must be either 'node' or 'tag'. Got: %q", configType),
		)
		return
	}
}

func (r *firewallResource) writeFirewall(ctx context.Context, plan firewallResourceModel) diag.Diagnostics {
	configType, identifier := plan.typeAndIdentifier()
	extend := plan.Extend.ValueBool()

	input := plan.Input
	tables := map[config.FirewallTableName]config.FirewallTable{
		config.Filter: {
			config.Input: input.toQbeeFirewallChain(),
		},
	}

	// Create the resource
	tflog.Info(ctx, fmt.Sprintf("Setting firewall for %v %v", configType, identifier))

	content := config.Firewall{
		Metadata: config.Metadata{
			Version: "v1",
			Enabled: true,
			Extend:  extend,
		},
		Tables: tables,
	}

	changeRequest, err := createChangeRequest(config.FirewallBundle, content, configType, identifier)
	if err != nil {
		return diag.Diagnostics{
			diag.NewErrorDiagnostic(
				errorWritingFirewall,
				err.Error(),
			),
		}
	}

	change, err := r.client.CreateConfigurationChange(ctx, changeRequest)
	if err != nil {
		return diag.Diagnostics{
			diag.NewErrorDiagnostic(
				errorWritingFirewall,
				err.Error(),
			),
		}
	}

	_, err = r.client.CommitConfiguration(ctx, "terraform: create firewall_resource")
	if err != nil {
		diags := diag.Diagnostics{}

		err = fmt.Errorf("error creating a commit for the firewall: %w", err)
		diags.AddError(errorWritingFirewall, err.Error())

		err = r.client.DeleteConfigurationChange(ctx, change.SHA)
		if err != nil {
			diags.AddError(
				errorWritingFirewall,
				fmt.Sprintf("error deleting uncommitted firewall changes: %v", err),
			)
		}

		return diags
	}

	return nil
}
