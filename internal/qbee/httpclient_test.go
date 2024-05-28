package qbee

import (
	"bytes"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/stretchr/testify/require"
)

func setup(t *testing.T) (*http.ServeMux, *HttpClient) {
	// mux is the HTTP request multiplexer used with the test server.
	mux := http.NewServeMux()

	// server is a test HTTP server used to provide mock API responses.
	server := httptest.NewServer(mux)
	t.Cleanup(server.Close)

	// Client is the Gitlab Client being tested.
	client, err := NewClient("",
		"",
		WithBaseURL(server.URL),
		//// Disable backoff to speed up tests that expect errors.
		//WithCustomBackoff(func(_, _ time.Duration, _ int, _ *http.Response) time.Duration {
		//	return 0
		//}),
	)
	if err != nil {
		t.Fatalf("Failed to create Client: %v", err)
	}

	mux.HandleFunc("/login", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprint(w, `{"token": "unit-test-login-token"}`)
	})

	return mux, client
}

func assertURL(t *testing.T, r *http.Request, want string) {
	if got := r.RequestURI; got != want {
		t.Errorf("Request url: %+v, want %s", got, want)
	}
}

func assertMethod(t *testing.T, r *http.Request, want string) {
	if got := r.Method; got != want {
		t.Errorf("Request method: %s, want %s", got, want)
	}
}

func assertBody(t *testing.T, r *http.Request, want string) {
	buffer := new(bytes.Buffer)
	_, err := buffer.ReadFrom(r.Body)
	if err != nil {
		t.Fatalf("Failed to Read Body: %v", err)
	}

	got := buffer.String()
	require.JSONEq(t, want, got)
}

func assertParams(t *testing.T, r *http.Request, want string) {
	if got := r.URL.RawQuery; got != want {
		t.Errorf("Request query: %s, want %s", got, want)
	}
}

func assertParam(t *testing.T, r *http.Request, key string, want string) {
	if got := r.URL.Query().Get(key); got != want {
		t.Errorf("Request query %s: %s, want %s", key, got, want)
	}
}

func mustWriteHTTPResponse(t *testing.T, w io.Writer, fixturePath string) {
	f, err := os.Open(fixturePath)
	if err != nil {
		t.Fatalf("error opening fixture file: %v", err)
	}

	if _, err = io.Copy(w, f); err != nil {
		t.Fatalf("error writing response: %v", err)
	}
}

//func errorOption(*retryablehttp.Request) error {
//	return errors.New("RequestOptionFunc returns an error")
//}
