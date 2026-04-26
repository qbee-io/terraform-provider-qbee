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
	_ resourceModelManager                  = &usersResourceModel{}
	_ resource.Resource                     = &usersResource{}
	_ resource.ResourceWithConfigure        = &usersResource{}
	_ resource.ResourceWithConfigValidators = &usersResource{}
	_ resource.ResourceWithImportState      = &usersResource{}
)

// NewUsersResource is a helper function to simplify the provider implementation.
func NewUsersResource() resource.Resource {
	return &usersResource{
		configurationResource: configurationResource{
			resourceBase: newResourceBase(config.UsersBundle),
			modelFactory: func() any {
				return new(usersResourceModel)
			},
		},
	}
}

type usersResource struct {
	configurationResource
}

// Schema defines the schema for the resource.
func (r *usersResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Users adds or removes users.",
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
				Description: "The users to add or remove.",
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"username": schema.StringAttribute{
							Required:    true,
							Description: "The username of the user to add or remove.",
						},
						"action": schema.StringAttribute{
							Required:    true,
							Description: "The action to perform on the user. Either 'add' or 'remove'.",
							Validators: []validator.String{
								stringvalidator.OneOf("add", "remove"),
							},
						},
					},
				},
			},
		},
	}
}

type usersResourceModel struct {
	configurationResourceModel
	Users []user `tfsdk:"users"`
}

type user struct {
	Username types.String `tfsdk:"username"`
	Action   types.String `tfsdk:"action"`
}

func (m usersResourceModel) getConfigBundle() config.Bundle {
	return config.UsersBundle
}

func (m *usersResourceModel) fromBundleData(bundleData config.BundleData) error {
	data := bundleData.Users
	if data == nil {
		return fmt.Errorf("users bundle data is nil")
	}

	m.Extend = types.BoolValue(data.Metadata.Extend)

	for _, u := range data.Users {
		m.Users = append(m.Users, user{
			Username: types.StringValue(u.Username),
			Action:   types.StringValue(string(u.Action)),
		})
	}

	return nil
}

func (m usersResourceModel) toBundleData(metadata config.Metadata) any {
	bundleData := config.Users{
		Metadata: metadata,
	}

	if metadata.Reset {
		return bundleData
	}

	for _, u := range m.Users {
		bundleData.Users = append(bundleData.Users, config.User{
			Username: u.Username.ValueString(),
			Action:   config.UserAction(u.Action.ValueString()),
		})
	}

	return bundleData
}
