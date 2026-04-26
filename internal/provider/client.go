package provider

import (
	"context"
	"fmt"
	"io"
	"sync"

	"github.com/hashicorp/terraform-plugin-log/tflog"
	"go.qbee.io/client"
	"go.qbee.io/client/config"
)

// Client is a wrapper around the Qbee API client that can be used as the provider data and resource data.
// It allows to easily access the Qbee API client from the resources and data sources.
type Client struct {
	sync.Mutex
	*client.Client
}

// NewClient creates a new Client instance with the given Qbee API client.
func NewClient() *Client {
	return &Client{
		Client: client.New(),
	}
}

// UploadFile uploads a file to the Qbee API using the provided path, name, and reader.
// It locks the client to prevent concurrent uploads and ensures that only one upload operation is performed at a time.
func (cli *Client) UploadFile(ctx context.Context, path, name string, reader io.Reader) error {
	cli.Lock()
	defer cli.Unlock()

	return cli.Client.UploadFile(ctx, path, name, reader)
}

// DeleteFile deletes a file from the Qbee API using the provided name.
// It locks the client to prevent concurrent deletions and ensures that only one delete operation is performed at a time.
func (cli *Client) DeleteFile(ctx context.Context, name string) error {
	cli.Lock()
	defer cli.Unlock()

	return cli.Client.DeleteFile(ctx, name)
}

// CommitConfiguration commits a configuration change to the Qbee API with the given message and change request.
// It locks the client to prevent concurrent commits and ensures that only one commit operation is performed at a time.
func (cli *Client) CommitConfiguration(ctx context.Context, message string, changes ...client.ChangeRequest) (*client.Commit, error) {
	cli.Lock()
	defer cli.Unlock()

	return cli.Client.CommitConfiguration(ctx, message, changes...)
}

// resourceModelManager defines common methods for managing resource models.
type resourceModelManager interface {
	// getBaseResourceModel returns the base resource model containing common fields like Node, Tag, and Extend.
	getBaseResourceModel() configurationResourceModel

	// setEntityID sets the entity ID (node or tag) on the model based on the provided entity type and ID.
	setEntityID(entityType config.EntityType, entityID string)	

	// getConfigBundle returns the configuration bundle associated with the resource model.
	getConfigBundle() config.Bundle

	// toBundleData converts the resource model to the corresponding bundle data for API requests.
	toBundleData(metadata config.Metadata) any

	// fromBundleData populates the resource model from the given bundle data retrieved from API responses.
	fromBundleData(bundleData config.BundleData) error
}

// commitConfiguration commits a configuration change for the given resource.
// If reset is true, it will commit a reset operation, otherwise it will commit a set operation.
func (cli *Client) commitConfiguration(ctx context.Context, model resourceModelManager, reset bool) (*client.Commit, error) {
	baseModel := model.getBaseResourceModel()

	// Log the operation being performed with relevant details
	message := "Setting "
	if reset {
		message = "Deleting "
	}

	message += fmt.Sprintf(
		"%s configuration for %s %s",
		model.getConfigBundle(), baseModel.getEntityType(), baseModel.getEntityID())

	tflog.Info(ctx, message)

	// Commit change request based on the resource model and the operation type (set or reset)
	changeReequest := client.ChangeRequest{
		BundleName: model.getConfigBundle(),
		Content: model.toBundleData(config.Metadata{
			Version: "v1",
			Enabled: true,
			Reset:   reset,
			Extend:  baseModel.Extend.ValueBool(),
		}),
	}

	entityType := baseModel.getEntityType()
	switch entityType {
	case config.EntityTypeNode:
		changeReequest.NodeID = baseModel.Node.ValueString()
	case config.EntityTypeTag:
		changeReequest.Tag = baseModel.Tag.ValueString()
	default:
		return nil, fmt.Errorf("unsupported entity type: %s", entityType)
	}

	return cli.CommitConfiguration(ctx, "terraform: "+message, changeReequest)
}
