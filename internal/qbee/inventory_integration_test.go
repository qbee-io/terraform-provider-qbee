//go:build integration

package qbee

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func Test_get_test_device_from_inventory(t *testing.T) {
	t.Run("it should find the device in the inventory", func(t *testing.T) {
		client, err := CreateTestClient()
		require.Nil(t, err)

		devices, err := client.Inventory.GetDevices()
		require.Nil(t, err)

		fmt.Print(devices)

		assert.Len(t, devices, 1)
	})
}
