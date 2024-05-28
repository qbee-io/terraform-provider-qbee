//go:build integration

package qbee

import (
	"crypto/sha256"
	"encoding/hex"
	"io"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func Test_file_lifecycle(t *testing.T) {
	client, err := CreateTestClient()
	if err != nil {
		t.Log(err)
		t.Fatalf("Could not create test Client")
	}
	filesService := FilesService{Client: client}

	file := "testfiles/upload_test.txt"

	targetFilename := "upload_test.txt"
	targetPath := "/files-integration-test/"
	uploadedPath := "/files-integration-test/upload_test.txt"

	fileDigest, err := Md5sum(file)
	if err != nil {
		t.Log(err)
		t.Fatalf("Could not determine checksum of file for test")
	}

	t.Run("it should be able to create a directory", func(t *testing.T) {
		result, err := filesService.CreateDir("/", targetPath)
		require.Nil(t, err)

		assert.Equal(t, "/files-integration-test/", result.Dir)
		assert.Equal(t, "/files-integration-test/", result.Path)
	})

	t.Run("it should be able to upload a file to that directory", func(t *testing.T) {
		got, err := filesService.Upload(file, targetPath, "")

		require.Nil(t, err)

		wants := FileUploadResponse{File: targetFilename, Path: "/files-integration-test"}
		assert.Equal(t, wants, *got)
	})

	t.Run("it should be able to download that file", func(t *testing.T) {
		got, err := filesService.Download(uploadedPath)

		require.Nil(t, err)

		f, err := os.Open(file)
		require.Nil(t, err)

		b, err := io.ReadAll(f)
		require.Nil(t, err)

		wants := FileDownloadResponse{Contents: string(b)}
		assert.Equal(t, wants, *got)
	})

	t.Run("it should be able to list files and find the file", func(t *testing.T) {
		got, err := filesService.List()

		require.Nil(t, err)

		contains := false
		for _, item := range got.Items {
			if item.Name == targetFilename {
				contains = true
			}
		}

		assert.True(t, contains)
	})

	t.Run("it should be able to get info on that specific file", func(t *testing.T) {
		got, err := filesService.GetMetadata(uploadedPath)

		require.Nil(t, err)

		assert.Equal(t, uploadedPath, got.Path)
		assert.Equal(t, false, got.IsDir)
		assert.Equal(t, "txt", got.Extension)
		assert.Equal(t, targetFilename, got.Name)
		assert.Equal(t, fileDigest, got.Digest)
	})

	t.Run("it should be able to delete a file", func(t *testing.T) {
		err := filesService.Delete(uploadedPath)
		assert.Nil(t, err)
	})

	t.Run("it should be able to delete the created directories", func(t *testing.T) {
		err := filesService.Delete("/files-integration-test")
		assert.Nil(t, err)
	})

	t.Run("it should no longer find the file afterward", func(t *testing.T) {
		_, err := filesService.GetMetadata(uploadedPath)

		assert.NotNil(t, err)
	})
}

func Md5sum(path string) (string, error) {
	f, err := os.Open(path)
	if err != nil {
		return "", err
	}
	defer f.Close()

	hasher := sha256.New()
	if _, err := io.Copy(hasher, f); err != nil {
		return "", err
	}

	return hex.EncodeToString(hasher.Sum(nil)), nil
}
