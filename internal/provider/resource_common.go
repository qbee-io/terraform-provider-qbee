package provider

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/resource"
	"go.qbee.io/client/config"
)

// resourceBase is a base struct that provides common functionality for all resources in the provider.
type resourceBase struct {
	// name is the name of the resource, e.g. "firewall", "docker_containers", etc. It is used for logging and error messages.
	name string

	// client is the provider configured client that can be used to interact with the Qbee API.
	client *Client
}

// newResourceBase is a helper function to create a new resourceBase with the given name.
func newResourceBase[T config.Bundle | string](name T) resourceBase {
	return resourceBase{
		name: string(name),
	}
}

// Configure adds the provider configured client to the resource.
func (r *resourceBase) Configure(_ context.Context, req resource.ConfigureRequest, _ *resource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}

	r.client = req.ProviderData.(*Client)
}

// Metadata returns the resource type name.
func (r *resourceBase) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = fmt.Sprintf("%s_%s", req.ProviderTypeName, r.name)
}
