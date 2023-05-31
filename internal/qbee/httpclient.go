package qbee

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/google/go-querystring/query"
	"io"
	"net/http"
	"net/url"
	"path"
	"strings"
)

const (
	DefaultBaseURL = "https://www.app.qbee.io/api/v2"
)

type HttpClient struct {
	Username string
	Password string

	authToken string
	baseURL   *url.URL

	Inventory *InventoryService
	Files     *FilesService
}

type ClientOptionFunc func(*HttpClient) error

func WithBaseURL(baseURL string) ClientOptionFunc {
	return func(c *HttpClient) error {
		return c.setBaseURL(baseURL)
	}
}

func (c *HttpClient) setBaseURL(urlStr string) error {
	// Make sure the given URL end with a slash
	if !strings.HasSuffix(urlStr, "/") {
		urlStr += "/"
	}

	baseURL, err := url.Parse(urlStr)
	if err != nil {
		return err
	}

	// Update the base URL of the Client.
	c.baseURL = baseURL

	return nil
}

func NewClient(username string, password string, options ...ClientOptionFunc) (*HttpClient, error) {
	var c = &HttpClient{
		Username: username,
		Password: password,
	}

	err := c.setBaseURL(DefaultBaseURL)
	if err != nil {
		return nil, fmt.Errorf("could not set baseURL %v: %w", DefaultBaseURL, err)
	}

	// Apply any given Client options.
	for _, fn := range options {
		if fn == nil {
			continue
		}
		if err := fn(c); err != nil {
			return nil, err
		}
	}

	c.Inventory = &InventoryService{client: c}
	c.Files = &FilesService{Client: c}

	return c, nil
}

func (c *HttpClient) Get(path string, q interface{}) (*http.Response, error) {
	u := c.buildURL(path)
	if q != nil {
		q, err := query.Values(q)
		if err != nil {
			return nil, err
		}
		u.RawQuery = q.Encode()
	}

	req, _ := http.NewRequest(http.MethodGet, u.String(), nil)

	response, err := c.AuthenticatedRequest(req)
	if err != nil {
		return nil, fmt.Errorf("httpclient.Get(%v): %w", path, err)
	}

	err = checkResponse(*response)
	if err != nil {
		return nil, fmt.Errorf("httpclient.Get(%v): %w", path, err)
	}

	return response, nil
}

func (c *HttpClient) Post(path string, body interface{}) (*http.Response, error) {
	var req *http.Request
	u := c.buildURL(path)

	if body != nil {
		buffer := new(bytes.Buffer)
		err := json.NewEncoder(buffer).Encode(body)
		if err != nil {
			return nil, fmt.Errorf("HttpClient.Post(%v) Marshal: ", path)
		}

		req, err = http.NewRequest(http.MethodPost, u.String(), buffer)
		if err != nil {
			return nil, fmt.Errorf("HttpClient.Post(%v) NewRequest: ", path)
		}
	} else {
		var err error
		req, err = http.NewRequest(http.MethodPost, u.String(), nil)
		if err != nil {
			return nil, fmt.Errorf("HttpClient.Post(%v) NewRequest: ", path)
		}
	}

	response, err := c.AuthenticatedRequest(req)
	if err != nil {
		return nil, fmt.Errorf("httpclient.Post(%v): %w", path, err)
	}

	err = checkResponse(*response)
	if err != nil {
		return nil, fmt.Errorf("httpclient.Post(%v): %w", path, err)
	}

	return response, nil
}

func (c *HttpClient) Delete(path string, body interface{}) (*http.Response, error) {
	var req *http.Request
	u := c.buildURL(path)

	if body != nil {
		buffer := new(bytes.Buffer)
		err := json.NewEncoder(buffer).Encode(body)
		if err != nil {
			return nil, fmt.Errorf("HttpClient.Delete(%v) Marshal: ", path)
		}

		req, err = http.NewRequest(http.MethodDelete, u.String(), buffer)
		if err != nil {
			return nil, fmt.Errorf("HttpClient.Delete(%v) NewRequest: ", path)
		}
	} else {
		var err error
		req, err = http.NewRequest(http.MethodDelete, u.String(), nil)
		if err != nil {
			return nil, fmt.Errorf("HttpClient.Delete(%v) NewRequest: ", path)
		}
	}

	response, err := c.AuthenticatedRequest(req)
	if err != nil {
		return nil, fmt.Errorf("httpclient.Delete(%v): %w", path, err)
	}

	err = checkResponse(*response)
	if err != nil {
		return nil, fmt.Errorf("httpclient.Delete(%v): %w", path, err)
	}

	return response, nil
}

func (c *HttpClient) UploadFile(path string, body io.Reader, contentType string) (*http.Response, error) {
	u := c.buildURL(path)
	req, _ := http.NewRequest(http.MethodPost, u.String(), body)
	req.Header.Add("Content-Type", contentType)

	response, err := c.AuthenticatedRequest(req)
	if err != nil {
		return nil, fmt.Errorf("failed to make authenticated request to %s: %w", path, err)
	}

	err = checkResponse(*response)
	if err != nil {
		return nil, fmt.Errorf("httpclient.UploadFile: %w", err)
	}

	return response, nil
}

func checkResponse(r http.Response) error {
	s := r.StatusCode
	if s >= 300 {
		b, err := io.ReadAll(r.Body)
		if err != nil {
			return fmt.Errorf("could not read body: %w", err)
		}

		return fmt.Errorf("non-OK status code '%v' (body='%v')", s, string(b))
	}

	return nil
}

func (c *HttpClient) AuthenticatedRequest(req *http.Request) (*http.Response, error) {
	auth, err := c.AuthToken()
	if err != nil {
		return nil, fmt.Errorf("could not retrieve Qbee auth token: %w", err)
	}

	req.Header.Add("Authorization", "Bearer "+auth)

	client := http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("error while executing HTTP request %v: %w", req, err)
	}

	return resp, nil
}

type authResponse struct {
	Token string `json:"token"`
}

type loginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

func (c *HttpClient) AuthToken() (string, error) {
	if c.authToken == "" {
		loginReq := loginRequest{
			Email:    c.Username,
			Password: c.Password,
		}

		b, err := json.Marshal(loginReq)
		if err != nil {
			return "", fmt.Errorf("could not create login request: %w", err)
		}

		resp, err := http.DefaultClient.Post(c.buildURL("/login").String(), "application/json", bytes.NewBuffer(b))
		if err != nil {
			return "", fmt.Errorf("could not execute login request: %w", err)
		}

		b, err = io.ReadAll(resp.Body)
		if err != nil {
			return "", fmt.Errorf("could not ready response body from bytes %s: %w", string(b), err)
		}

		err = resp.Body.Close()
		if err != nil {
			return "", fmt.Errorf("error closing response body: %w", err)
		}

		if resp.StatusCode != 200 {
			return "", fmt.Errorf("unexpected status code '%v' returned by qbee login (email=%v): '%v'", resp.StatusCode, c.Username, string(b))
		}

		var auth authResponse
		err = json.Unmarshal(b, &auth)
		if err != nil {
			return "", fmt.Errorf("could not parse response body to auth response: '%v', %w", string(b), err)
		}

		c.authToken = auth.Token
	}

	return c.authToken, nil
}

func (c *HttpClient) buildURL(p string) *url.URL {
	u := *c.baseURL
	u.Path = path.Join(u.Path, p)
	return &u
}

func (c *HttpClient) BaseURL() *url.URL {
	u := *c.baseURL
	return &u
}
