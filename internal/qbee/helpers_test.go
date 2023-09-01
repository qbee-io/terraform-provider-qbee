package qbee

import (
	"os"
)

type localTestConfig struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

func CreateTestClient() (*HttpClient, error) {
	config := localTestConfig{
		Username: os.Getenv("QBEE_USERNAME"),
		Password: os.Getenv("QBEE_PASSWORD"),
	}

	return NewClient(config.Username, config.Password)
}
