package provider

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"go.qbee.io/client/config"
)

// Ensure the implementation satisfies the expected interfaces.
var (
	_ resourceModelManager                  = &sshKeysResourceModel{}
	_ resource.Resource                     = &sshKeysResource{}
	_ resource.ResourceWithConfigure        = &sshKeysResource{}
	_ resource.ResourceWithConfigValidators = &sshKeysResource{}
	_ resource.ResourceWithImportState      = &sshKeysResource{}
)

// NewSSHKeysResource is a helper function to simplify the provider implementation.
func NewSSHKeysResource() resource.Resource {
	return &sshKeysResource{
		configurationResource: configurationResource{
			resourceBase: newResourceBase("ssh_keys"),
			modelFactory: func() any {
				return new(sshKeysResourceModel)
			},
		},
	}
}

type sshKeysResource struct {
	configurationResource
}

// Schema defines the schema for the resource.
func (r *sshKeysResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "SSHKeys adds or removes authorized SSH keys for users.",
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
			"users": schema.ListNestedAttribute{
				Required:    true,
				Description: "The users to set SSH keys for.",
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"username": schema.StringAttribute{
							Required:    true,
							Description: "Username of the user for which the SSH keys are set.",
						},
						"keys": schema.ListAttribute{
							Required:    true,
							Description: "The SSH keys to set for the user.",
							ElementType: types.StringType,
						},
					},
				},
			},
		},
	}
}

type sshKeysResourceModel struct {
	configurationResourceModel
	Users []sskKey `tfsdk:"users"`
}

type sskKey struct {
	Username types.String   `tfsdk:"username"`
	Keys     []types.String `tfsdk:"keys"`
}

func (m sshKeysResourceModel) getConfigBundle() config.Bundle {
	return config.SSHKeysBundle
}

func (m *sshKeysResourceModel) fromBundleData(bundleData config.BundleData) error {
	data := bundleData.SSHKeys
	if data == nil {
		return fmt.Errorf("sshkeys bundle data is nil")
	}

	m.Extend = types.BoolValue(data.Metadata.Extend)

	for _, user := range data.Users {
		keys := make([]types.String, 0)
		for _, key := range user.Keys {
			keys = append(keys, types.StringValue(key))
		}

		m.Users = append(m.Users, sskKey{
			Username: types.StringValue(user.Username),
			Keys:     keys,
		})
	}

	return nil
}

func (m sshKeysResourceModel) toBundleData(metadata config.Metadata) any {
	bundleData := config.SSHKeys{
		Metadata: metadata,
	}

	if metadata.Reset {
		return bundleData
	}

	for _, user := range m.Users {
		var keys []string
		for _, key := range user.Keys {
			keys = append(keys, key.ValueString())
		}

		bundleData.Users = append(bundleData.Users, config.SSHKey{
			Username: user.Username.ValueString(),
			Keys:     keys,
		})
	}

	return bundleData
}
