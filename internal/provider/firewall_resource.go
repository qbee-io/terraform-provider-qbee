package provider

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"go.qbee.io/client/config"
)

// Ensure the implementation satisfies the expected interfaces.
var (
	_ resourceModelManager                  = &firewallResourceModel{}
	_ resource.Resource                     = &firewallResource{}
	_ resource.ResourceWithConfigure        = &firewallResource{}
	_ resource.ResourceWithConfigValidators = &firewallResource{}
	_ resource.ResourceWithImportState      = &firewallResource{}
)

func NewFirewallResource() resource.Resource {
	return &firewallResource{
		configurationResource: configurationResource{
			resourceBase: newResourceBase(config.FirewallBundle),
			modelFactory: func() any {
				return new(firewallResourceModel)
			},
		},
	}
}

type firewallResource struct {
	configurationResource
}

// Schema defines the schema for the resource.
func (r *firewallResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Firewall configures system firewall.",
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
									Description: "The protocol to match. Either udp or tcp.",
									Validators: []validator.String{
										stringvalidator.OneOf("udp", "tcp"),
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

type firewallResourceModel struct {
	configurationResourceModel
	Input *firewallInput `tfsdk:"input"`
}

func (m firewallResourceModel) getConfigBundle() config.Bundle {
	return config.FirewallBundle
}

func (i firewallInput) toBundleData() config.FirewallChain {
	firewallChain := config.FirewallChain{
		Policy: config.Target(i.Policy.ValueString()),
	}

	for _, rule := range i.Rules {
		firewallChain.Rules = append(firewallChain.Rules, config.FirewallRule{
			Protocol:        config.Protocol(rule.Proto.ValueString()),
			Target:          config.Target(rule.Target.ValueString()),
			SourceIP:        rule.SrcIp.ValueString(),
			DestinationPort: rule.DstPort.ValueString(),
		})
	}

	return firewallChain
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

func (m *firewallResourceModel) fromBundleData(bundleData config.BundleData) error {
	data := bundleData.Firewall
	if data == nil {
		return fmt.Errorf("firewall bundle data is nil")
	}

	m.Extend = types.BoolValue(data.Metadata.Extend)

	inputChain := data.Tables[config.Filter][config.Input]

	m.Input = &firewallInput{
		Policy: types.StringValue(string(inputChain.Policy)),
	}

	for _, rule := range inputChain.Rules {
		m.Input.Rules = append(m.Input.Rules, firewallRule{
			Proto:   types.StringValue(string(rule.Protocol)),
			Target:  types.StringValue(string(rule.Target)),
			SrcIp:   types.StringValue(rule.SourceIP),
			DstPort: types.StringValue(rule.DestinationPort),
		})
	}

	return nil
}

func (m firewallResourceModel) toBundleData(metadata config.Metadata) any {
	bundleData := config.Firewall{
		Metadata: metadata,
	}

	if metadata.Reset {
		return bundleData
	}

	bundleData.Tables = map[config.FirewallTableName]config.FirewallTable{
		config.Filter: {
			config.Input: m.Input.toBundleData(),
		},
	}

	return bundleData
}
