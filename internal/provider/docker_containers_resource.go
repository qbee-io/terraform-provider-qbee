package provider

import (
	"context"
	"fmt"
	"github.com/hashicorp/terraform-plugin-framework-validators/resourcevalidator"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringdefault"
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
	_ resource.Resource                     = &dockerContainersResource{}
	_ resource.ResourceWithConfigure        = &dockerContainersResource{}
	_ resource.ResourceWithConfigValidators = &dockerContainersResource{}
	_ resource.ResourceWithImportState      = &dockerContainersResource{}
)

const (
	errorImportingDockerContainers = "error importing docker_containers resource"
	errorWritingDockerContainers   = "error writing docker_containers resource"
	errorReadingDockerContainers   = "error reading docker_containers resource"
	errorDeletingDockerContainers  = "error deleting docker_containers resource"
)

// NewDockerContainersResource is a helper function to simplify the provider implementation.
func NewDockerContainersResource() resource.Resource {
	return &dockerContainersResource{}
}

type dockerContainersResource struct {
	client *client.Client
}

type dockerContainersResourceModel struct {
	Node          types.String                   `tfsdk:"node"`
	Tag           types.String                   `tfsdk:"tag"`
	Extend        types.Bool                     `tfsdk:"extend"`
	Containers    []dockerContainerResourceModel `tfsdk:"containers"`
	RegistryAuths []registryAuthResourceModel    `tfsdk:"registry_auths"`
}

type dockerContainerResourceModel struct {
	containerResourceModel
	DockerArgs types.String `tfsdk:"docker_args"`
}

func (m dockerContainersResourceModel) typeAndIdentifier() (config.EntityType, string) {
	return typeAndIdentifier(m.Tag, m.Node)
}

// Metadata returns the resource type name.
func (r *dockerContainersResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_docker_containers"
}

// Configure adds the provider configured client to the resource.
func (r *dockerContainersResource) Configure(_ context.Context, req resource.ConfigureRequest, _ *resource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}

	r.client = req.ProviderData.(*client.Client)
}

// Schema defines the schema for the resource.
func (r *dockerContainersResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
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

func (r *dockerContainersResource) ConfigValidators(ctx context.Context) []resource.ConfigValidator {
	return []resource.ConfigValidator{
		resourcevalidator.ExactlyOneOf(
			path.MatchRoot("tag"),
			path.MatchRoot("node"),
		),
	}
}

// Create creates the resource and sets the initial Terraform state.
func (r *dockerContainersResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	// Retrieve values from the plan
	var plan dockerContainersResourceModel
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	diags = r.writeDockerContainers(ctx, plan)
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
func (r *dockerContainersResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	// Retrieve values from the plan
	var plan dockerContainersResourceModel
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	diags = r.writeDockerContainers(ctx, plan)
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
func (r *dockerContainersResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	// Get the current state
	var state *dockerContainersResourceModel
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	configType, identifier := state.typeAndIdentifier()

	// Read the real status
	activeConfig, err := r.client.GetActiveConfig(ctx, configType, identifier, config.EntityConfigScopeOwn)
	if err != nil {
		resp.Diagnostics.AddError(errorReadingDockerContainers,
			"error reading the active configuration: "+err.Error())

		return
	}

	// Update the current state
	currentDockerContainers := activeConfig.BundleData.DockerContainers
	if currentDockerContainers == nil {
		resp.State.RemoveResource(ctx)
		return
	}

	state.Extend = types.BoolValue(currentDockerContainers.Extend)

	var mappedContainers []dockerContainerResourceModel
	for _, container := range currentDockerContainers.Containers {
		mappedContainers = append(mappedContainers, dockerContainerResourceModel{
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
	state.Containers = mappedContainers

	var mappedRegistryAuths []registryAuthResourceModel
	for _, registryAuth := range currentDockerContainers.RegistryAuths {
		mappedRegistryAuths = append(mappedRegistryAuths, registryAuthResourceModel{
			Server:   types.StringValue(registryAuth.Server),
			Username: types.StringValue(registryAuth.Username),
			Password: types.StringValue(registryAuth.Password),
		})
	}
	if len(mappedRegistryAuths) == 0 {
		state.RegistryAuths = nil
	} else {
		state.RegistryAuths = mappedRegistryAuths
	}

	diags = resp.State.Set(ctx, state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}

// Delete deletes the resource and removes the Terraform state on success.
func (r *dockerContainersResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	// Retrieve values from the state
	var state dockerContainersResourceModel
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Delete the resource
	configType, identifier := state.typeAndIdentifier()
	tflog.Info(ctx, fmt.Sprintf("Deleting docker_containers for %v %v", configType, identifier))

	content := config.DockerContainers{
		Metadata: config.Metadata{
			Reset:   true,
			Version: "v1",
		},
	}

	changeRequest, err := createChangeRequest(config.DockerContainersBundle, content, configType, identifier)
	if err != nil {
		resp.Diagnostics.AddError(
			errorDeletingDockerContainers,
			err.Error(),
		)
		return
	}

	change, err := r.client.CreateConfigurationChange(ctx, changeRequest)
	if err != nil {
		resp.Diagnostics.AddError(
			errorDeletingDockerContainers,
			err.Error(),
		)
		return
	}

	_, err = r.client.CommitConfiguration(ctx, "terraform: create dockerContainers_resource")
	if err != nil {
		resp.Diagnostics.AddError(errorDeletingDockerContainers,
			"error creating a commit to delete the docker_containers resource: "+err.Error(),
		)

		err = r.client.DeleteConfigurationChange(ctx, change.SHA)
		if err != nil {
			resp.Diagnostics.AddError(
				errorDeletingDockerContainers,
				"error deleting uncommitted docker_containers changes: "+err.Error(),
			)
		}

		return
	}
}

// ImportState imports the resource state from the Terraform state.
func (r *dockerContainersResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	configType, identifier, found := strings.Cut(req.ID, ":")
	if !found || configType == "" || identifier == "" {
		resp.Diagnostics.AddError(
			errorImportingDockerContainers,
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
			errorImportingDockerContainers,
			fmt.Sprintf("Import type must be either 'node' or 'tag'. Got: %q", configType),
		)
		return
	}
}

func (r *dockerContainersResource) writeDockerContainers(ctx context.Context, plan dockerContainersResourceModel) diag.Diagnostics {
	configType, identifier := plan.typeAndIdentifier()
	extend := plan.Extend.ValueBool()

	// Create the resource
	tflog.Info(ctx, fmt.Sprintf("Creating docker_containers for %v %v", configType, identifier))

	var mappedContainers []config.Container
	for _, container := range plan.Containers {
		mappedContainers = append(mappedContainers, config.Container{
			Name:         container.Name.ValueString(),
			Image:        container.Image.ValueString(),
			PreCondition: container.PreCondition.ValueString(),
			EnvFile:      container.EnvFile.ValueString(),
			Command:      container.Command.ValueString(),
			DockerArgs:   container.DockerArgs.ValueString(),
		})
	}

	var mappedRegistryAuths []config.RegistryAuth
	for _, registryAuth := range plan.RegistryAuths {
		mappedRegistryAuths = append(mappedRegistryAuths, config.RegistryAuth{
			Server:   registryAuth.Server.ValueString(),
			Username: registryAuth.Username.ValueString(),
			Password: registryAuth.Password.ValueString(),
		})
	}

	content := config.DockerContainers{
		Metadata: config.Metadata{
			Enabled: true,
			Extend:  extend,
			Version: "v1",
		},
		Containers:    mappedContainers,
		RegistryAuths: mappedRegistryAuths,
	}

	changeRequest, err := createChangeRequest(config.DockerContainersBundle, content, configType, identifier)
	if err != nil {
		return diag.Diagnostics{
			diag.NewErrorDiagnostic(
				errorWritingDockerContainers,
				err.Error(),
			),
		}
	}

	change, err := r.client.CreateConfigurationChange(ctx, changeRequest)
	if err != nil {
		return diag.Diagnostics{
			diag.NewErrorDiagnostic(
				errorWritingDockerContainers,
				fmt.Sprintf("Error creating a docker_containers resource with qbee: %v", err),
			),
		}
	}

	_, err = r.client.CommitConfiguration(ctx, "terraform: create dockerContainers_resource")
	if err != nil {
		diags := diag.Diagnostics{}

		err = fmt.Errorf("error creating a commit for the docker_containers: %w", err)
		diags.AddError(errorWritingDockerContainers, err.Error())

		err = r.client.DeleteConfigurationChange(ctx, change.SHA)
		if err != nil {
			diags.AddError(
				errorWritingDockerContainers,
				fmt.Errorf("error deleting uncommitted docker_containers changes: %w", err).Error(),
			)
		}

		return diags
	}

	return nil
}
