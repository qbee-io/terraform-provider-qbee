package qbee

import (
	"fmt"
	"github.com/stretchr/testify/assert"
	"io"
	"net/http"
	"os"
	"testing"
)

//mux.HandleFunc("/files", func(w http.ResponseWriter, r *http.Request) {
//	assertMethod(t, r, http.MethodGet)
//
//	f, err := os.Open("testfiles/list_files_response.json")
//	assert.Nil(t, err)
//
//	c, err := io.ReadAll(f)
//	assert.Nil(t, err)
//
//	fmt.Fprint(w, string(c))
//})

func TestGetFiledistribution(t *testing.T) {
	mux, client := setup(t)

	tagId := "some:test-tag"

	mux.HandleFunc(fmt.Sprintf("/config/tag/%s", tagId), func(w http.ResponseWriter, r *http.Request) {
		assertMethod(t, r, http.MethodGet)

		f, err := os.Open("testfiles/get_tag_configuration_response.json")
		assert.Nil(t, err)

		c, err := io.ReadAll(f)
		assert.Nil(t, err)

		fmt.Fprint(w, string(c))
	})

	got, err := client.TagConfig.GetFiledistribution(tagId)
	assert.Nil(t, err)

	wants := GetFileDistributionResponse{
		Enabled:        true,
		Extend:         true,
		Version:        "v1",
		BundleCommitId: "filedist_bundlecommitid",
		Files: []FileDistributionFileResponse{
			{
				Command:      "echo \"done!\"",
				PreCondition: "",
				Templates: []FiledistributionFileTemplateResponse{
					{
						Source:      "sources/file1",
						Destination: "/destinations/file1",
						IsTemplate:  true,
					},
				},
				Parameters: []FiledistributionFileParameterResponse{
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
			{
				Command:      "",
				PreCondition: "/bin/true",
				Templates: []FiledistributionFileTemplateResponse{
					{
						Source:      "sources/file2",
						Destination: "/destinations/file2",
						IsTemplate:  false,
					},
				},
				Parameters: nil,
			},
		},
	}

	assert.Equal(t, wants, *got)
}
