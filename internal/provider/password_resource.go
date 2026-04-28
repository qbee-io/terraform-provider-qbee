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
	_ resourceModelManager                  = &passwordResourceModel{}
	_ resource.Resource                     = &passwordResource{}
	_ resource.ResourceWithConfigure        = &passwordResource{}
	_ resource.ResourceWithConfigValidators = &passwordResource{}
	_ resource.ResourceWithImportState      = &passwordResource{}
)

// NewPasswordResource is a helper function to simplify the provider implementation.
func NewPasswordResource() resource.Resource {
	return &passwordResource{
		configurationResource: configurationResource{
			resourceBase: newResourceBase(config.PasswordBundle),
			modelFactory: func() any {
				return new(passwordResourceModel)
			},
		},
	}
}

type passwordResource struct {
	configurationResource
}

// Schema defines the schema for the resource.
func (r *passwordResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Password bundle sets passwords for existing users.",
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
				Description: "A list of users and their password hashes.",
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"username": schema.StringAttribute{
							Required:    true,
							Description: "The username of the user for which the password hash is set.",
						},
						"password_hash": schema.StringAttribute{
							Required:    true,
							Description: "The password hash for the user. See https://qbee.io/docs/qbee-password.html for more information.",
						},
					},
				},
			},
		},
	}
}

type passwordResourceModel struct {
	configurationResourceModel
	Users []userPassword `tfsdk:"users"`
}

type userPassword struct {
	Username     types.String `tfsdk:"username"`
	PasswordHash types.String `tfsdk:"password_hash"`
}

func (m passwordResourceModel) getConfigBundle() config.Bundle {
	return config.PasswordBundle
}

func (m *passwordResourceModel) fromBundleData(bundleData config.BundleData) error {
	data := bundleData.Password
	if data == nil {
		return fmt.Errorf("password bundle data is nil")
	}

	m.Extend = types.BoolValue(data.Metadata.Extend)

	for _, user := range data.Users {
		m.Users = append(m.Users, userPassword{
			Username:     types.StringValue(user.Username),
			PasswordHash: types.StringValue(user.PasswordHash),
		})
	}

	return nil
}

func (m passwordResourceModel) toBundleData(metadata config.Metadata) any {
	bundleData := config.Password{
		Metadata: metadata,
	}

	if metadata.Reset {
		return bundleData
	}

	for _, user := range m.Users {
		bundleData.Users = append(bundleData.Users, config.UserPassword{
			Username:     user.Username.ValueString(),
			PasswordHash: user.PasswordHash.ValueString(),
		})
	}

	return bundleData
}
