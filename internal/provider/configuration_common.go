package provider

import (
	"context"
	"fmt"
	"slices"
	"strings"

	"github.com/hashicorp/terraform-plugin-framework-validators/resourcevalidator"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"
	"go.qbee.io/client/config"
)

func nullableStringValue(value string) basetypes.StringValue {
	if value == "" {
		return types.StringNull()
	} else {
		return types.StringValue(value)
	}
}

// configurationResource is a base struct that is embedded in all configuration resources.
type configurationResource struct {
	resourceBase

	// modelFactory is a function that returns a pointer to an empty instance of the resource model struct.
	// It is used to create new instances of the model for each operation (create, read, update, delete)
	// without having to know the specific type of the model in the resource implementation.
	modelFactory func() any
}

func (r *configurationResource) ConfigValidators(ctx context.Context) []resource.ConfigValidator {
	return []resource.ConfigValidator{
		resourcevalidator.ExactlyOneOf(
			path.MatchRoot("tag"),
			path.MatchRoot("node"),
		),
	}
}

// Create creates the resource and sets the initial Terraform state.
func (r *configurationResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	model := r.modelFactory()

	// Get the model from the plan
	if resp.Diagnostics.Append(req.Plan.Get(ctx, model)...); resp.Diagnostics.HasError() {
		return
	}

	// Commit the configuration change
	if _, err := r.client.commitConfiguration(ctx, model.(resourceModelManager), false); err != nil {
		resp.Diagnostics.AddError(
			fmt.Sprintf("Error creating %s configuration", r.name),
			err.Error(),
		)
		return
	}

	// If commit was successful, we can assume the state now reflects the desired configuration, so we set the state to match the plan
	resp.Diagnostics.Append(resp.State.Set(ctx, model)...)
}

// Update updates the resource and sets the updated Terraform state on success.
func (r *configurationResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	model := r.modelFactory()

	// Get the model from the plan
	if resp.Diagnostics.Append(req.Plan.Get(ctx, model)...); resp.Diagnostics.HasError() {
		return
	}

	// Commit the configuration change
	if _, err := r.client.commitConfiguration(ctx, model.(resourceModelManager), false); err != nil {
		resp.Diagnostics.AddError(
			fmt.Sprintf("Error updating %s configuration", r.name),
			err.Error(),
		)
		return
	}

	// If commit was successful, we can assume the state now reflects the desired configuration, so we set the state to match the plan
	resp.Diagnostics.Append(resp.State.Set(ctx, model)...)
}

// Read refreshes the Terraform state with the latest data.
func (r *configurationResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var nodeID, tag *string

	resp.Diagnostics.Append(req.State.GetAttribute(ctx, path.Root("node"), &nodeID)...)
	resp.Diagnostics.Append(req.State.GetAttribute(ctx, path.Root("tag"), &tag)...)
	if resp.Diagnostics.HasError() {
		return
	}

	var entityType config.EntityType
	var entityID string
	if tag != nil && *tag != "" {
		entityType = config.EntityTypeTag
		entityID = *tag
	} else {
		entityType = config.EntityTypeNode
		entityID = *nodeID
	}

	// Retrieve the active configuration for the resource from the API
	activeConfig, err := r.client.GetActiveConfig(ctx, entityType, entityID, config.EntityConfigScopeOwn)
	if err != nil {
		resp.Diagnostics.AddError(
			fmt.Sprintf("Error reading %s configuration", r.name),
			err.Error())

		return
	}

	model := r.modelFactory()
	resourceManager := model.(resourceModelManager)
	resourceManager.setEntityID(activeConfig.Type, activeConfig.EntityID)

	// Remove the resource from the state if the active configuration does not contain the relevant bundle
	if !slices.Contains(activeConfig.Bundles, resourceManager.getConfigBundle()) {
		resp.State.RemoveResource(ctx)
		return
	}

	// Otherwise parse the bundle data and update the response state with the current configuration
	if err := resourceManager.fromBundleData(activeConfig.BundleData); err != nil {
		resp.Diagnostics.AddError(
			fmt.Sprintf("Error parsing %s configuration", r.name),
			err.Error(),
		)
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, model)...)
}

// Delete deletes the resource and removes the Terraform state on success.
func (r *configurationResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	model := r.modelFactory()

	// Get the model from the current state to identify which resource to reset
	if resp.Diagnostics.Append(req.State.Get(ctx, model)...); resp.Diagnostics.HasError() {
		return
	}

	// Commit the configuration change with reset=true.
	if _, err := r.client.commitConfiguration(ctx, model.(resourceModelManager), true); err != nil {
		resp.Diagnostics.AddError(
			fmt.Sprintf("Error deleting %s configuration", model.(resourceModelManager).getConfigBundle()),
			err.Error(),
		)
	}
}

// ImportState imports the resource state from the Terraform state.
func (r *configurationResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	entityType, entityID, found := strings.Cut(req.ID, ":")
	if !found || entityType == "" || entityID == "" {
		resp.Diagnostics.AddError(
			fmt.Sprintf("Error importing %s configuration", r.name),
			fmt.Sprintf("Expected import identifier with format: type:identifier. Got: %q", req.ID))
		return
	}

	switch entityType {
	case "tag":
		resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("tag"), entityID)...)
	case "node":
		resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("node"), entityID)...)
	default:
		resp.Diagnostics.AddError(
			fmt.Sprintf("Error importing %s configuration", r.name),
			fmt.Sprintf("Import type must be either 'node' or 'tag'. Got: %q", entityType),
		)
	}
}

// configurationResourceModel defines the common fields for all resources that are associated with a node or a tag.
type configurationResourceModel struct {
	Node   types.String `tfsdk:"node"`
	Tag    types.String `tfsdk:"tag"`
	Extend types.Bool   `tfsdk:"extend"`
}

// getBaseResourceModel returns the base resource model associated with the resource.
func (m configurationResourceModel) getBaseResourceModel() configurationResourceModel {
	return m
}

// setEntityID sets the node and tag IDs on the model based on the provided entity type and ID.
func (m *configurationResourceModel) setEntityID(entityType config.EntityType, entityID string) {
	if entityType == config.EntityTypeNode {
		m.Node = types.StringValue(entityID)
	} else {
		m.Tag = types.StringValue(entityID)
	}
}

// getEntityType returns the entity type (node or tag) associated with the resource model.
func (m configurationResourceModel) getEntityType() config.EntityType {
	if !m.Tag.IsNull() {
		return config.EntityTypeTag
	}

	return config.EntityTypeNode
}

// getEntityID returns the entity identifier (node ID or tag name) associated with the resource model.
func (m configurationResourceModel) getEntityID() string {
	if m.getEntityType() == config.EntityTypeTag {
		return m.Tag.ValueString()
	}

	return m.Node.ValueString()
}
