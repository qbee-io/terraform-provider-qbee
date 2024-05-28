package qbee

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"reflect"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestDownloadFile(t *testing.T) {
	mux, client := setup(t)

	mux.HandleFunc("/file", func(w http.ResponseWriter, r *http.Request) {
		t.Logf("handling %v", r.URL.String())
		assertMethod(t, r, http.MethodGet)
		assertParam(t, r, "path", "/target/test.txt")

		fmt.Fprint(w, "Test contents")
	})

	got, err := client.Files.Download("/target/test.txt")

	if err != nil {
		t.Fatalf("error from Files.Download: %v", err)
	}

	wants := FileDownloadResponse{Contents: "Test contents"}
	assert.Equal(t, wants, *got)
}

func TestUploadFile(t *testing.T) {
	mux, client := setup(t)

	mux.HandleFunc("/file", func(w http.ResponseWriter, r *http.Request) {
		assertMethod(t, r, http.MethodPost)

		if !strings.Contains(r.Header.Get("Content-Type"), "multipart/form-data;") {
			t.Fatalf("Projects.UploadFile request content-type %+v want multipart/form-data;", r.Header.Get("Content-Type"))
		}
		if r.ContentLength == -1 {
			t.Fatalf("Projects.UploadFile request content-length is -1")
		}

		fmt.Fprint(w, `{
    "file": "upload_test.txt",
    "path": "/target"
}`)
	})

	fileUpload, err := client.Files.Upload("testfiles/upload_test.txt", "/target", "")

	if err != nil {
		t.Fatalf("Upload threw error: %v", err)
	}

	want := FileUploadResponse{File: "upload_test.txt", Path: "/target"}

	if !reflect.DeepEqual(want, *fileUpload) {
		t.Fatalf("wanted %v, but got %v", want, *fileUpload)
	}
}

func TestDelete(t *testing.T) {
	mux, client := setup(t)

	mux.HandleFunc("/file", func(w http.ResponseWriter, r *http.Request) {
		assertMethod(t, r, http.MethodDelete)
		assertBody(t, r, `{"path": "/path/to/delete.txt"}`)

		w.WriteHeader(http.StatusNoContent)
		fmt.Fprint(w, "")
	})

	err := client.Files.Delete("/path/to/delete.txt")
	assert.Nil(t, err)
}

func TestCreateDir(t *testing.T) {
	mux, client := setup(t)

	pathToCreateAt := "/root"
	dirToCreate := "newdir"

	mux.HandleFunc("/file/createdir", func(w http.ResponseWriter, r *http.Request) {
		assertBody(t, r, `{"path": "/root", "name": "newdir"}`)

		fmt.Fprint(w, `{"dir": "newdir", "path": "/root/newdir/"}`)
	})

	response, err := client.Files.CreateDir(pathToCreateAt, dirToCreate)
	assert.Nil(t, err)
	assert.Equal(t, "newdir", response.Dir)
	assert.Equal(t, "/root/newdir/", response.Path)
}

func TestListFiles(t *testing.T) {
	mux, client := setup(t)

	mux.HandleFunc("/files", func(w http.ResponseWriter, r *http.Request) {
		assertMethod(t, r, http.MethodGet)

		f, err := os.Open("testfiles/list_files_response.json")
		assert.Nil(t, err)

		c, err := io.ReadAll(f)
		assert.Nil(t, err)

		fmt.Fprint(w, string(c))
	})

	got, err := client.Files.List()
	assert.Nil(t, err)

	wants := ListFilesResponse{Items: []FileMetadata{
		{Name: "test", Path: "/test/", IsDir: true, Created: 1685256027, Size: 0},
		{Name: "upload", Path: "/test/upload/", IsDir: true, Created: 1685256027, Size: 0},
		{Name: "upload_test.txt", Path: "/test/upload/upload_test.txt", IsDir: false, Created: 1685256027, Size: 4, Extension: "txt", Mime: "text/plain"},
	}}

	assert.Equal(t, wants, *got)
}
