package qbee

import (
	"encoding/json"
	"io"
)

type InventoryService struct {
	client *HttpClient
}

func (s InventoryService) GetDevices() (devices []Device, err error) {
	response, err := s.client.Get("/inventory", nil)
	if err != nil {
		return nil, err
	}

	responseData, err := io.ReadAll(response.Body)
	if err != nil {
		return nil, err
	}

	var responseObject getInventoryResponse
	err = json.Unmarshal(responseData, &responseObject)
	if err != nil {
		return nil, err
	}

	return devices, nil
}

type Device struct{}

type getInventoryResponse struct {
	Items []inventoryItem `json:"items"`
}

type inventoryItem struct {
	PubKeyDigest string `json:"pub_key_digest"`
	NodeId       string `json:"node_id"`
}
