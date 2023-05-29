package qbee

import (
	"encoding/json"
	"io"
	"os"
)

type localTestConfig struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

func CreateTestClient() (*HttpClient, error) {
	f, err := os.Open("local-testing.json")
	if err != nil {
		return nil, err
	}

	byteValue, _ := io.ReadAll(f)

	var config localTestConfig

	err = json.Unmarshal(byteValue, &config)
	if err != nil {
		return nil, err
	}

	client, err := NewClient(config.Username, config.Password)
	return client, nil
}
