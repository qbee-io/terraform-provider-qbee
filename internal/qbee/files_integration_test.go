//go:build integration

package qbee

import (
	"github.com/stretchr/testify/assert"
	"io"
	"os"
	"testing"
)

func Test_file_lifecycle(t *testing.T) {
	client, err := CreateTestClient()
	if err != nil {
		t.Log(err)
		t.Fatalf("Could not create test Client")
	}
	filesService := FilesService{Client: client}

	file := "testfiles/upload_test.txt"
	targetPath := "/test/"
	uploadedPath := "/test/upload_test.txt"

	t.Run("it should be able to create a directory", func(t *testing.T) {
		t.Fatalf("not implemented")
	})

	t.Run("it should be able to upload a file to that directory", func(t *testing.T) {
		got, err := filesService.Upload(file, targetPath)

		assert.Nil(t, err)

		wants := FileUploadResponse{File: "upload_test.txt", Path: "/test/"}
		assert.Equal(t, wants, got)
	})

	t.Run("it should be able to describe a file", func(t *testing.T) {
		got, err := filesService.Download(uploadedPath)

		assert.Nil(t, err)

		f, err := os.Open(file)
		assert.Nil(t, err)

		b, err := io.ReadAll(f)
		assert.Nil(t, err)

		wants := FileDownloadResponse{Contents: string(b)}
		assert.Equal(t, wants, *got)
	})

	t.Run("it should be able to list files", func(t *testing.T) {

	})

	t.Run("it should be able to delete a file", func(t *testing.T) {

	})

	t.Run("it should be able to delete the created directories", func(t *testing.T) {

	})

	t.Run("it should no longer find the file afterward", func(t *testing.T) {

	})
}
