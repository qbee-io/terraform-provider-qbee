package provider

import (
	"go.qbee.io/client"
	"go.qbee.io/client/config"
)

func createChangeRequest(bundleName config.Bundle, change any, configType config.EntityType, identifier string) (client.ChangeRequest, error) {
	changeRequest := client.ChangeRequest{
		BundleName: bundleName,
		Content:    change,
	}
	if configType == config.EntityTypeNode {
		changeRequest.NodeID = identifier
	} else if configType == config.EntityTypeTag {
		changeRequest.Tag = identifier
	}

	return changeRequest, nil
}
