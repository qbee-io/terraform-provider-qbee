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
	_ resourceModelManager                  = &podmanContainersResourceModel{}
	_ resource.Resource                     = &podmanContainersResource{}
	_ resource.ResourceWithConfigure        = &podmanContainersResource{}
	_ resource.ResourceWithConfigValidators = &podmanContainersResource{}
	_ resource.ResourceWithImportState      = &podmanContainersResource{}
)

// NewPodmanContainersResource is a helper function to simplify the provider implementation.
func NewPodmanContainersResource() resource.Resource {
	return &podmanContainersResource{
		configurationResource: configurationResource{
			resourceBase: newResourceBase(config.PodmanContainersBundle),
			modelFactory: func() any {
				return new(podmanContainersResourceModel)
			},
		},
	}
}

type podmanContainersResource struct {
	configurationResource
}

// Schema defines the schema for the resource.
func (r *podmanContainersResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "podman_containers controls podman containers running in the system.",
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
						"podman_args": schema.StringAttribute{
							Optional:    true,
							Computed:    true,
							Default:     stringdefault.StaticString(""),
							Description: "Command line arguments for 'podman run'",
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

type podmanContainersResourceModel struct {
	configurationResourceModel
	Containers    []podmanContainerResourceModel `tfsdk:"containers"`
	RegistryAuths []registryAuthResourceModel    `tfsdk:"registry_auths"`
}

type podmanContainerResourceModel struct {
	containerResourceModel
	PodmanArgs types.String `tfsdk:"podman_args"`
}

func (m podmanContainersResourceModel) getConfigBundle() config.Bundle {
	return config.PodmanContainersBundle
}

func (m *podmanContainersResourceModel) fromBundleData(bundleData config.BundleData) error {
	data := bundleData.PodmanContainers
	if data == nil {
		return fmt.Errorf("podman containers bundle data is nil")
	}

	m.Extend = types.BoolValue(data.Metadata.Extend)

	for _, container := range data.Containers {
		m.Containers = append(m.Containers, podmanContainerResourceModel{
			containerResourceModel: containerResourceModel{
				Name:         types.StringValue(container.Name),
				Image:        types.StringValue(container.Image),
				PreCondition: types.StringValue(container.PreCondition),
				EnvFile:      types.StringValue(container.EnvFile),
				Command:      types.StringValue(container.Command),
			},
			PodmanArgs: types.StringValue(container.DockerArgs),
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

func (m podmanContainersResourceModel) toBundleData(metadata config.Metadata) any {
	bundleData := config.PodmanContainers{
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
			DockerArgs:   container.PodmanArgs.ValueString(),
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
