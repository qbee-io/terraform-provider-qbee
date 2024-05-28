//go:build integration

package qbee

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func Test_grouptree_lifecycle(t *testing.T) {
	client, err := CreateTestClient()
	if err != nil {
		t.Log(err)
		t.Fatalf("Could not create test Client")
	}
	grouptreeService := GrouptreeService{Client: client}

	groupParent := "integrationtests"
	groupId := "undertest"
	groupTitle := "Group under test"

	t.Run("it should be able to create a group", func(t *testing.T) {
		err := grouptreeService.Create(groupId, groupParent, groupTitle)

		assert.Nil(t, err)
	})

	t.Run("it should be able to describe the group", func(t *testing.T) {
		resp, err := grouptreeService.Get(groupId)

		assert.Nil(t, err)

		wants := GetGrouptreeResponse{
			Type:   "group",
			NodeId: groupId,
			Title:  groupTitle,
			Ancestors: []string{
				"root",
				groupParent,
				groupId,
			},
		}
		assert.Equal(t, wants, *resp)
	})

	t.Run("it should be able to rename the group", func(t *testing.T) {
		err := grouptreeService.Rename(groupId, groupParent, "Some new title")

		assert.Nil(t, err)
	})

	t.Run("it should be able to delete the group", func(t *testing.T) {
		err := grouptreeService.Delete(groupId, groupParent)

		assert.Nil(t, err)
	})
}
