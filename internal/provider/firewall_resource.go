package provider

import (
	"context"
	"fmt"
	"strings"

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

	"go.qbee.io/terraform/internal/qbee"
)

// Ensure the implementation satisfies the expected interfaces.
var (
	_ resource.Resource                     = &firewallResource{}
	_ resource.ResourceWithConfigure        = &firewallResource{}
	_ resource.ResourceWithConfigValidators = &firewallResource{}
	_ resource.ResourceWithImportState      = &firewallResource{}
)

func NewFirewallResource() resource.Resource {
	return &firewallResource{}
}

type firewallResource struct {
	client *qbee.HttpClient
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

	r.client = req.ProviderData.(*qbee.HttpClient)
}

// Schema defines the schema for the resource.
func (r *firewallResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
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
	ID     types.String   `tfsdk:"id"`
	Extend types.Bool     `tfsdk:"extend"`
	Input  *firewallInput `tfsdk:"input"`
}

func (m firewallResourceModel) typeAndIdentifier() (qbee.ConfigType, string) {
	return typeAndIdentifier(m.Tag, m.Node)
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
	plan.ID = types.StringValue("placeholder")

	// Set state to fully populated data
	diags = resp.State.Set(ctx, plan)
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
	plan.ID = types.StringValue("placeholder")

	// Set state to fully populated data
	diags = resp.State.Set(ctx, plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}

func (r *firewallResource) writeFirewall(ctx context.Context, plan firewallResourceModel) diag.Diagnostics {
	var diags diag.Diagnostics

	configType, identifier := plan.typeAndIdentifier()
	extend := plan.Extend.ValueBool()

	var mappedRules []qbee.FirewallRule
	for _, rule := range plan.Input.Rules {
		mappedRules = append(mappedRules, qbee.FirewallRule{
			Proto:   rule.Proto.ValueString(),
			Target:  rule.Target.ValueString(),
			SrcIp:   rule.SrcIp.ValueString(),
			DstPort: rule.DstPort.ValueString(),
		})
	}
	var firewallTables = qbee.FirewallTables{
		Filter: qbee.FirewallFilter{
			Input: qbee.FirewallConfig{
				Policy: plan.Input.Policy.ValueString(),
				Rules:  mappedRules,
			},
		},
	}

	// Create the resource
	tflog.Info(ctx, fmt.Sprintf("Setting firewall for %v %v", configType.String(), identifier))
	createResponse, err := r.client.Firewall.Create(configType, identifier, firewallTables, extend)
	if err != nil {
		diags.AddError(
			"Error creating firewall",
			fmt.Sprintf("Error creating a firewall resource with qbee: %v", err),
		)
		return diags
	}

	_, err = r.client.Configuration.Commit("terraform: create firewall_resource")
	if err != nil {
		err = fmt.Errorf("error creating a commit for the firewall: %w", err)

		err = r.client.Configuration.DeleteUncommitted(createResponse.Sha)
		if err != nil {
			err = fmt.Errorf("error deleting uncommitted firewall changes: %w", err)
		}

		diags.AddError(
			"Error creating firewall",
			fmt.Sprintf("Error while committing firewall change: %v", err),
		)
		return diags
	}

	return nil
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
	currentFirewall, err := r.client.Firewall.Get(configType, identifier)
	if err != nil {
		resp.Diagnostics.AddError("Could not read firewall",
			"error reading the firewall resource: "+err.Error())

		return
	}

	if currentFirewall == nil {
		resp.State.RemoveResource(ctx)
		return
	}

	// Update the current state
	var inputRules []firewallRule
	for _, rule := range currentFirewall.FirewallTables.Filter.Input.Rules {
		inputRules = append(inputRules, firewallRule{
			Proto:   types.StringValue(rule.Proto),
			Target:  types.StringValue(rule.Target),
			SrcIp:   types.StringValue(rule.SrcIp),
			DstPort: types.StringValue(rule.DstPort),
		})
	}
	input := firewallInput{
		Policy: types.StringValue(currentFirewall.FirewallTables.Filter.Input.Policy),
		Rules:  inputRules,
	}

	state.ID = types.StringValue("placeholder")
	state.Extend = types.BoolValue(currentFirewall.Extend)
	state.Input = &input

	diags = resp.State.Set(ctx, state)
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
	tflog.Info(ctx, fmt.Sprintf("Deleting firewall for %v %v", configType.String(), identifier))

	deleteResponse, err := r.client.Firewall.Clear(configType, identifier)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error deleting firewawll",
			"could not delete firewawll, unexpected error: "+err.Error())
		return
	}

	_, err = r.client.Configuration.Commit("terraform: create firewall_resource")
	if err != nil {
		resp.Diagnostics.AddError("Could not commit deletion of firewawll",
			"error creating a commit to delete the firewall resource: "+err.Error())

		err = r.client.Configuration.DeleteUncommitted(deleteResponse.Sha)
		if err != nil {
			resp.Diagnostics.AddError("Could not revert uncommitted firewall changes",
				"error deleting uncommitted firewall changes: "+err.Error())
		}

		return
	}
}

func (r *firewallResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
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
