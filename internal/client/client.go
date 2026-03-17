package client

import (
	"bytes"
	"crypto/hmac"
	"crypto/sha256"
	"crypto/tls"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/google/uuid"
)

// Client represents a Pritunl API client
type Client struct {
	baseURL    string
	token      string
	secret     string
	httpClient *http.Client
}

// NewClient creates a new Pritunl API client
func NewClient(baseURL, token, secret string, insecure bool) *Client {
	transport := &http.Transport{
		TLSClientConfig: &tls.Config{
			InsecureSkipVerify: insecure,
		},
	}

	return &Client{
		baseURL: strings.TrimSuffix(baseURL, "/"),
		token:   token,
		secret:  secret,
		httpClient: &http.Client{
			Transport: transport,
			Timeout:   30 * time.Second,
		},
	}
}

// generateAuth generates HMAC-SHA256 authentication headers
func (c *Client) generateAuth(method, path string) map[string]string {
	timestamp := fmt.Sprintf("%d", time.Now().Unix())
	nonce := strings.ReplaceAll(uuid.New().String(), "-", "") // Pritunl expects UUID hex without dashes

	// Build auth string: token + timestamp + nonce + method + path
	// Note: body is NOT included in the signature per Pritunl API docs
	authString := c.token + "&" + timestamp + "&" + nonce + "&" + strings.ToUpper(method) + "&" + path

	// Generate HMAC-SHA256 signature
	h := hmac.New(sha256.New, []byte(c.secret))
	h.Write([]byte(authString))
	signature := base64.StdEncoding.EncodeToString(h.Sum(nil))

	return map[string]string{
		"Auth-Token":     c.token,
		"Auth-Timestamp": timestamp,
		"Auth-Nonce":     nonce,
		"Auth-Signature": signature,
	}
}

// doRequest performs an authenticated HTTP request
func (c *Client) doRequest(method, path string, body interface{}) ([]byte, error) {
	var bodyBytes []byte
	var err error

	if body != nil {
		bodyBytes, err = json.Marshal(body)
		if err != nil {
			return nil, fmt.Errorf("failed to marshal request body: %w", err)
		}
	}

	url := c.baseURL + path
	req, err := http.NewRequest(method, url, bytes.NewReader(bodyBytes))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	// Set headers
	req.Header.Set("Content-Type", "application/json")
	for key, value := range c.generateAuth(method, path) {
		req.Header.Set(key, value)
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %w", err)
	}

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return nil, fmt.Errorf("API request failed with status %d: %s", resp.StatusCode, string(respBody))
	}

	return respBody, nil
}

// Get performs an authenticated GET request
func (c *Client) Get(path string) ([]byte, error) {
	return c.doRequest(http.MethodGet, path, nil)
}

// Post performs an authenticated POST request
func (c *Client) Post(path string, body interface{}) ([]byte, error) {
	return c.doRequest(http.MethodPost, path, body)
}

// Put performs an authenticated PUT request
func (c *Client) Put(path string, body interface{}) ([]byte, error) {
	return c.doRequest(http.MethodPut, path, body)
}

// Delete performs an authenticated DELETE request
func (c *Client) Delete(path string) ([]byte, error) {
	return c.doRequest(http.MethodDelete, path, nil)
}

// Status checks if the Pritunl API is accessible
func (c *Client) Status() error {
	_, err := c.Get("/status")
	return err
}
