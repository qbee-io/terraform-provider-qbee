package provider

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringdefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"go.qbee.io/client/config"
)

// Ensure the implementation satisfies the expected interfaces.
var (
	_ resourceModelManager                  = &dockerContainersResourceModel{}
	_ resource.Resource                     = &dockerContainersResource{}
	_ resource.ResourceWithConfigure        = &dockerContainersResource{}
	_ resource.ResourceWithConfigValidators = &dockerContainersResource{}
	_ resource.ResourceWithImportState      = &dockerContainersResource{}
)

// NewDockerContainersResource is a helper function to simplify the provider implementation.
func NewDockerContainersResource() resource.Resource {
	return &dockerContainersResource{
		configurationResource: configurationResource{
			resourceBase: newResourceBase(config.DockerContainersBundle),
			modelFactory: func() any {
				return new(dockerContainersResourceModel)
			},
		},
	}
}

type dockerContainersResource struct {
	configurationResource
}

// Schema defines the schema for the resource.
func (r *dockerContainersResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Controls docker containers running in the system.",
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
			"containers": schema.ListNestedAttribute{
				Required:    true,
				Description: "The list of containers to be running in the system.",
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"name": schema.StringAttribute{
							Required:    true,
							Description: "The name used by the container",
						},
						"image": schema.StringAttribute{
							Required:    true,
							Description: "The image to be used by the container",
						},
						"docker_args": schema.StringAttribute{
							Optional:    true,
							Computed:    true,
							Default:     stringdefault.StaticString(""),
							Description: "Command line arguments for 'docker run'",
						},
						"env_file": schema.StringAttribute{
							Optional:    true,
							Computed:    true,
							Default:     stringdefault.StaticString(""),
							Description: "An env file (from file manager) to be used inside the container",
						},
						"command": schema.StringAttribute{
							Optional:    true,
							Computed:    true,
							Default:     stringdefault.StaticString(""),
							Description: "Command to be executed in the container",
						},
						"pre_condition": schema.StringAttribute{
							Optional:    true,
							Computed:    true,
							Default:     stringdefault.StaticString(""),
							Description: "A condition that must be met before the container is started",
						},
					},
				},
			},
			"registry_auths": schema.ListNestedAttribute{
				Optional:    true,
				Description: "Credentials for container registry authentication.",
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"server": schema.StringAttribute{
							Required:    true,
							Description: "Hostname of the registry",
						},
						"username": schema.StringAttribute{
							Required:    true,
							Description: "Username for the registry",
						},
						"password": schema.StringAttribute{
							Required:    true,
							Description: "Password for the registry",
						},
					},
				},
			},
		},
	}
}

type dockerContainerResourceModel struct {
	containerResourceModel
	DockerArgs types.String `tfsdk:"docker_args"`
}

type dockerContainersResourceModel struct {
	configurationResourceModel
	Containers    []dockerContainerResourceModel `tfsdk:"containers"`
	RegistryAuths []registryAuthResourceModel    `tfsdk:"registry_auths"`
}

func (m dockerContainersResourceModel) getConfigBundle() config.Bundle {
	return config.DockerContainersBundle
}

func (m *dockerContainersResourceModel) fromBundleData(bundleData config.BundleData) error {
	data := bundleData.DockerContainers
	if data == nil {
		return fmt.Errorf("docker_containers bundle data is nil")
	}

	m.Extend = types.BoolValue(data.Metadata.Extend)

	for _, container := range data.Containers {
		m.Containers = append(m.Containers, dockerContainerResourceModel{
			containerResourceModel: containerResourceModel{
				Name:         types.StringValue(container.Name),
				Image:        types.StringValue(container.Image),
				PreCondition: types.StringValue(container.PreCondition),
				EnvFile:      types.StringValue(container.EnvFile),
				Command:      types.StringValue(container.Command),
			},
			DockerArgs: types.StringValue(container.DockerArgs),
		})
	}

	for _, registryAuth := range data.RegistryAuths {
		m.RegistryAuths = append(m.RegistryAuths, registryAuthResourceModel{
			Server:   types.StringValue(registryAuth.Server),
			Username: types.StringValue(registryAuth.Username),
			Password: types.StringValue(registryAuth.Password),
		})
	}

	return nil
}

func (m dockerContainersResourceModel) toBundleData(metadata config.Metadata) any {
	bundleData := config.DockerContainers{
		Metadata: metadata,
	}

	if metadata.Reset {
		return bundleData
	}

	for _, container := range m.Containers {
		bundleData.Containers = append(bundleData.Containers, config.Container{
			Name:         container.Name.ValueString(),
			Image:        container.Image.ValueString(),
			PreCondition: container.PreCondition.ValueString(),
			EnvFile:      container.EnvFile.ValueString(),
			Command:      container.Command.ValueString(),
			DockerArgs:   container.DockerArgs.ValueString(),
		})
	}

	for _, registryAuth := range m.RegistryAuths {
		bundleData.RegistryAuths = append(bundleData.RegistryAuths, config.RegistryAuth{
			Server:   registryAuth.Server.ValueString(),
			Username: registryAuth.Username.ValueString(),
			Password: registryAuth.Password.ValueString(),
		})
	}

	return bundleData
}
