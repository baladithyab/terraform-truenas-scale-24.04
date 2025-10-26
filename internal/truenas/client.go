package truenas

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"
)

const (
	defaultTimeout = 30 * time.Second
	apiVersion     = "v2.0"
)

// Client is the TrueNAS API client
type Client struct {
	BaseURL    string
	APIKey     string
	HTTPClient *http.Client
}

// NewClient creates a new TrueNAS API client
func NewClient(baseURL, apiKey string) (*Client, error) {
	if baseURL == "" {
		return nil, fmt.Errorf("base URL cannot be empty")
	}
	if apiKey == "" {
		return nil, fmt.Errorf("API key cannot be empty")
	}

	// Ensure baseURL doesn't have trailing slash
	baseURL = strings.TrimSuffix(baseURL, "/")

	return &Client{
		BaseURL: baseURL,
		APIKey:  apiKey,
		HTTPClient: &http.Client{
			Timeout: defaultTimeout,
		},
	}, nil
}

// DoRequest performs an HTTP request to the TrueNAS API
func (c *Client) DoRequest(method, endpoint string, body interface{}) ([]byte, error) {
	var reqBody io.Reader
	if body != nil {
		jsonData, err := json.Marshal(body)
		if err != nil {
			return nil, fmt.Errorf("error marshaling request body: %w", err)
		}
		reqBody = bytes.NewBuffer(jsonData)
	}

	url := fmt.Sprintf("%s/api/%s%s", c.BaseURL, apiVersion, endpoint)
	req, err := http.NewRequest(method, url, reqBody)
	if err != nil {
		return nil, fmt.Errorf("error creating request: %w", err)
	}

	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", c.APIKey))
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("error performing request: %w", err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("error reading response body: %w", err)
	}

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return nil, fmt.Errorf("API request failed with status %d: %s", resp.StatusCode, string(respBody))
	}

	return respBody, nil
}

// Get performs a GET request
func (c *Client) Get(endpoint string) ([]byte, error) {
	return c.DoRequest(http.MethodGet, endpoint, nil)
}

// Post performs a POST request
func (c *Client) Post(endpoint string, body interface{}) ([]byte, error) {
	return c.DoRequest(http.MethodPost, endpoint, body)
}

// Put performs a PUT request
func (c *Client) Put(endpoint string, body interface{}) ([]byte, error) {
	return c.DoRequest(http.MethodPut, endpoint, body)
}

// Delete performs a DELETE request
func (c *Client) Delete(endpoint string) ([]byte, error) {
	return c.DoRequest(http.MethodDelete, endpoint, nil)
}

// Patch performs a PATCH request
func (c *Client) Patch(endpoint string, body interface{}) ([]byte, error) {
	return c.DoRequest(http.MethodPatch, endpoint, body)
}

