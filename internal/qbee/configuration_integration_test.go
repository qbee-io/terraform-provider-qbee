//go:build integration

package qbee

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func Test_change_lifecycle(t *testing.T) {
	client, err := CreateTestClient()
	require.Nil(t, err)

	tagName := "terraform:configuration-integrationtest"

	testFiles := []FiledistributionFile{
		{
			PreCondition: "",
			Command:      "",
			Templates: []FiledistributionTemplate{
				{
					Source:      "/some/source",
					Destination: "/other/dest",
					IsTemplate:  true,
				},
			},
			Parameters: []FiledistributionParameter{
				{
					Key:   "key",
					Value: "value",
				},
			},
		},
	}

	var testSoftwareItems = []SoftwareManagementItem{
		{
			Package:      "TestPackage",
			ServiceName:  "TestName",
			PreCondition: "/bin/true",
			ConfigFiles: []SoftwareManagementConfigFile{
				{
					ConfigTemplate: "/source",
					ConfigLocation: "/destination",
				},
			},
			Parameters: []SoftwareManagementParameter{
				{
					Key:   "some-key",
					Value: "the-value",
				},
			},
		},
	}

	t.Run("test configuration lifecycle", func(t *testing.T) {
		//
		// If we create a change, we should see it as being uncommitted
		//
		_, err := client.FileDistribution.Create(ConfigForTag, tagName, testFiles, false)
		require.Nil(t, err)

		pendingChanges, err := client.Configuration.GetUncommitted()
		require.Nil(t, err)

		assert.Len(t, pendingChanges, 1)

		//
		// Clear all uncommitted changes, and verify the list is empty
		//
		err = client.Configuration.ClearUncommitted()
		require.Nil(t, err)

		pendingChanges, err = client.Configuration.GetUncommitted()
		require.Nil(t, err)
		require.Empty(t, pendingChanges)

		//
		// Add and commit FileDistribution and verify we see no uncommitted changes, and the FileDistribution as config
		//
		_, err = client.FileDistribution.Create(ConfigForTag, tagName, testFiles, false)
		require.Nil(t, err)

		_, err = client.Configuration.Commit("Terraform provider - add filedistribution for integration test")
		require.Nil(t, err)

		pendingChanges, err = client.Configuration.GetUncommitted()
		require.Nil(t, err)
		assert.Empty(t, pendingChanges)

		resp, err := client.Configuration.GetConfiguration(ConfigForTag, tagName)
		require.Nil(t, err)
		require.NotNil(t, resp.Config.BundleData.FileDistribution)

		//
		// Add SoftwareManagement on the tag and verify we see both
		//
		_, err = client.SoftwareManagement.Create(ConfigForTag, tagName, testSoftwareItems, false)
		require.Nil(t, err)

		_, err = client.Configuration.Commit("Terraform provider - add filedistribution for integration test")
		require.Nil(t, err)

		resp, err = client.Configuration.GetConfiguration(ConfigForTag, tagName)
		require.Nil(t, err)

		require.NotNil(t, t, resp.Config.BundleData)
		var bd = resp.Config.BundleData
		require.NotNil(t, bd.FileDistribution)
		require.NotNil(t, bd.SoftwareManagement)

		var actualFiles = bd.FileDistribution.FiledistributionFiles
		require.Len(t, actualFiles, 1)
		require.Len(t, actualFiles[0].Templates, 1)
		assert.Equal(t, actualFiles[0].Templates[0].Source, "/some/source")
		assert.Equal(t, actualFiles[0].Templates[0].Destination, "/other/dest")
		assert.Equal(t, actualFiles[0].Templates[0].IsTemplate, true)
		require.Len(t, actualFiles[0].Parameters, 1)
		assert.Equal(t, actualFiles[0].Parameters[0].Key, "key")
		assert.Equal(t, actualFiles[0].Parameters[0].Value, "value")

		var actualItems = bd.SoftwareManagement.SoftwareItems
		require.Len(t, actualItems, 1)
		assert.Equal(t, actualItems[0].Package, "TestPackage")
		assert.Equal(t, actualItems[0].ServiceName, "TestName")
		assert.Equal(t, actualItems[0].PreCondition, "/bin/true")
		require.Len(t, actualItems[0].ConfigFiles, 1)
		assert.Equal(t, actualItems[0].ConfigFiles[0].ConfigTemplate, "/source")
		assert.Equal(t, actualItems[0].ConfigFiles[0].ConfigLocation, "/destination")
		require.Len(t, actualItems[0].Parameters, 1)
		assert.Equal(t, actualItems[0].Parameters[0].Key, "some-key")
		assert.Equal(t, actualItems[0].Parameters[0].Value, "the-value")

		// Clear FileDistribution from the tag
		_, err = client.FileDistribution.Clear(ConfigForTag, tagName)
		require.Nil(t, err)

		_, err = client.Configuration.Commit("Terraform provider - cleared filedistribution from tag")
		require.Nil(t, err)

		resp, err = client.Configuration.GetConfiguration(ConfigForTag, tagName)
		require.Nil(t, err)

		bd = resp.Config.BundleData
		require.Nil(t, bd.FileDistribution)
		require.NotNil(t, bd.SoftwareManagement)

		// Clear SoftwareManagement from the tag
		_, err = client.SoftwareManagement.Clear(ConfigForTag, tagName)
		require.Nil(t, err)

		_, err = client.Configuration.Commit("Terraform provider - cleared filedistribution from tag")
		require.Nil(t, err)

		resp, err = client.Configuration.GetConfiguration(ConfigForTag, tagName)
		require.Nil(t, err)

		bd = resp.Config.BundleData
		assert.Nil(t, bd.FileDistribution)
		assert.Nil(t, bd.SoftwareManagement)
	})

	t.Cleanup(func() {
		err = client.Configuration.ClearUncommitted()
		require.Nil(t, err)

		_, err = client.FileDistribution.Clear(ConfigForTag, tagName)
		require.Nil(t, err)

		_, err = client.SoftwareManagement.Clear(ConfigForTag, tagName)
		require.Nil(t, err)

		changes, err := client.Configuration.GetUncommitted()
		require.Nil(t, err)

		if len(changes) > 0 {
			_, err = client.Configuration.Commit("Terraform provider - Clean up after configuration integration test")
			require.Nil(t, err)
		}
	})
}
