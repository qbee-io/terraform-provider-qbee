//go:build integration

package qbee

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func Test_grouptree_lifecycle(t *testing.T) {
	client, err := CreateTestClient()
	if err != nil {
		t.Log(err)
		t.Fatalf("Could not create test Client")
	}
	grouptreeService := GrouptreeService{Client: client}

	groupParent := "integrationtests"
	secondGroupParent := "root"
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

	t.Run("it should be able to move the group", func(t *testing.T) {
		err := grouptreeService.Move(groupId, groupParent, secondGroupParent)

		assert.Nil(t, err)

		if !t.Failed() {
			// Change the groupParent, so we delete it from the correct group in the next test
			groupParent = secondGroupParent
		}
	})

	t.Run("it should be able to delete the group", func(t *testing.T) {
		err := grouptreeService.Delete(groupId, groupParent)

		assert.Nil(t, err)
	})
}
