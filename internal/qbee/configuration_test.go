package qbee

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGetConfigurationFiledistributionAndSoftware(t *testing.T) {
	mux, client := setup(t)

	mux.HandleFunc("/config/tag/tagname", func(w http.ResponseWriter, r *http.Request) {
		t.Logf("handling %v", r.URL.String())
		assertMethod(t, r, http.MethodGet)

		f, err := os.Open("testfiles/get_tag_configuration_response.json")
		require.Nil(t, err)

		c, err := io.ReadAll(f)
		require.Nil(t, err)

		_, err = fmt.Fprint(w, string(c))
		require.Nil(t, err)
	})

	got, err := client.Configuration.GetConfiguration(ConfigForTag, "tagname")

	if err != nil {
		t.Fatalf("error from Configuration.GetConfiguration: %v", err)
	}

	wants := GetConfigurationResponse{
		Config: GetConfigurationConfig{
			Id:            "tagname",
			Type:          "tag",
			CommitId:      "abcdefghijklmnopqrstuvwxyz",
			CommitCreated: 1688388581822251355,
			Bundles:       []string{"file_distribution", "software_management"},
			BundleData: ConfigurationBundleData{
				FileDistribution: &BundleConfiguration{
					Enabled:        true,
					Extend:         true,
					Version:        "v1",
					BundleCommitId: "filedist_bundlecommitid",
					FiledistributionFiles: []FiledistributionFile{
						{
							Command: "echo \"done!\"",
							Templates: []FiledistributionTemplate{
								{
									Source:      "sources/file1",
									Destination: "/destinations/file1",
									IsTemplate:  true,
								},
							},
							Parameters: []FiledistributionParameter{
								{
									Key:   "parameter-1",
									Value: "parameter-value-1",
								},
								{
									Key:   "parameter-2",
									Value: "parameter-value-2",
								},
							},
						},
					},
				},
				SoftwareManagement: &BundleConfiguration{
					Enabled:        true,
					Extend:         true,
					Version:        "v1",
					BundleCommitId: "softwareman_bundlecommitid",
					SoftwareItems: []SoftwareManagementItem{
						{
							Package:     "/some/package.deb",
							ServiceName: "somepackage",
						},
					},
				},
			},
		},
		Status: "OK",
	}

	assert.Equal(t, wants, *got)
}

func TestGetConfigurationEmptyBundledata(t *testing.T) {
	mux, client := setup(t)

	mux.HandleFunc("/config/tag/tag:withemptybundledata", func(w http.ResponseWriter, r *http.Request) {
		t.Logf("handling %v", r.URL.String())
		assertMethod(t, r, http.MethodGet)

		f, err := os.Open("testfiles/get_tag_configuration_response_empty_bundledata.json")
		require.Nil(t, err)

		c, err := io.ReadAll(f)
		require.Nil(t, err)

		_, err = fmt.Fprint(w, string(c))
		require.Nil(t, err)
	})

	got, err := client.Configuration.GetConfiguration(ConfigForTag, "tag:withemptybundledata")

	if err != nil {
		t.Fatalf("error from Configuration.GetConfiguration: %v", err)
	}

	wants := GetConfigurationResponse{
		Config: GetConfigurationConfig{
			Id:            "tag:withemptybundledata",
			Type:          "tag",
			CommitId:      "abcdefghijklmnopqrstuvwxyz",
			CommitCreated: 1686643585663473286,
			Bundles:       nil,
			BundleData: ConfigurationBundleData{
				FileDistribution:   nil,
				SoftwareManagement: nil,
			},
		},
		Status: "OK",
	}

	assert.Equal(t, wants, *got)
}
