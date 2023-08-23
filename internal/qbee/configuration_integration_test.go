//go:build integration

package qbee

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func Test_change_lifecycle(t *testing.T) {
	client, err := CreateTestClient()
	if err != nil {
		t.Log(err)
		t.Fatalf("Could not create test Client")
	}

	changeService := ConfigurationService{Client: client}

	tagName := "qbeeclient:acctest"

	t.Run("it should show an empty configuration for an unknown tag", func(t *testing.T) {
		resp, err := changeService.GetTagConfiguration(tagName)
		assert.Nil(t, err)

		assert.NotNil(t, resp)
	})

	t.Run("it should be able to create and delete configuration for a tag", func(t *testing.T) {
		// Change
		// Commit
		// Verify
		// Clear
		// Commit
		//response, err := changeService.ChangeTagFiledistribution()
	})

	t.Run("it should be able to delete uncommitted changes", func(t *testing.T) {
		// Change
		// Clear uncommitted
		// Verify
	})
}