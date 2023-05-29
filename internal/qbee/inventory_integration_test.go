//go:build integration

package qbee

import (
	"fmt"
	"github.com/stretchr/testify/assert"
	"log"
	"os"
	"testing"
)

func Test_get_test_device_from_inventory(t *testing.T) {
	t.Run("it should find the device in the inventory", func(t *testing.T) {
		var client = HttpClient{
			Username: "TODO: Username",
			Password: "TODO: Password",
		}
		var inventoryService = InventoryService{client: &client}

		devices, err := inventoryService.GetDevices()
		if err != nil {
			log.Fatal(err)
			os.Exit(1)
		}

		fmt.Print(devices)

		assert.Len(t, devices, 1)
	})
}
