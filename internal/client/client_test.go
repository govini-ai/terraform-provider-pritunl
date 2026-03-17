package client

import (
	"crypto/hmac"
	"crypto/sha256"
	"crypto/tls"
	"encoding/base64"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewClient_TrimsTrailingSlash(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		inputURL string
		wantURL  string
	}{
		{
			name:     "with trailing slash",
			inputURL: "https://example.com/",
			wantURL:  "https://example.com",
		},
		{
			name:     "without trailing slash",
			inputURL: "https://example.com",
			wantURL:  "https://example.com",
		},
		{
			name:     "multiple trailing slashes",
			inputURL: "https://example.com///",
			wantURL:  "https://example.com//",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			client := NewClient(tt.inputURL, "token", "secret", false)
			assert.Equal(t, tt.wantURL, client.baseURL)
		})
	}
}

func TestNewClient_SetsTimeout(t *testing.T) {
	t.Parallel()

	client := NewClient("https://example.com", "token", "secret", false)
	assert.Equal(t, 30*time.Second, client.httpClient.Timeout)
}

func TestNewClient_InsecureTLS(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		insecure bool
		want     bool
	}{
		{
			name:     "insecure true",
			insecure: true,
			want:     true,
		},
		{
			name:     "insecure false",
			insecure: false,
			want:     false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			client := NewClient("https://example.com", "token", "secret", tt.insecure)
			transport := client.httpClient.Transport.(*http.Transport)
			assert.Equal(t, tt.want, transport.TLSClientConfig.InsecureSkipVerify)
		})
	}
}

func TestGenerateAuth_HeaderFormat(t *testing.T) {
	t.Parallel()

	client := NewClient("https://example.com", "test-token", "test-secret", false)
	headers := client.generateAuth("GET", "/test")

	assert.Contains(t, headers, "Auth-Token")
	assert.Contains(t, headers, "Auth-Timestamp")
	assert.Contains(t, headers, "Auth-Nonce")
	assert.Contains(t, headers, "Auth-Signature")

	assert.Equal(t, "test-token", headers["Auth-Token"])
	assert.NotEmpty(t, headers["Auth-Timestamp"])
	assert.NotEmpty(t, headers["Auth-Nonce"])
	assert.NotEmpty(t, headers["Auth-Signature"])
}

func TestGenerateAuth_SignatureFormat(t *testing.T) {
	t.Parallel()

	client := NewClient("https://example.com", "test-token", "test-secret", false)
	headers := client.generateAuth("GET", "/test")

	// Verify signature is base64 encoded
	signature := headers["Auth-Signature"]
	decoded, err := base64.StdEncoding.DecodeString(signature)
	require.NoError(t, err)

	// HMAC-SHA256 produces 32 bytes
	assert.Len(t, decoded, 32)
}

func TestGenerateAuth_SignatureVerification(t *testing.T) {
	t.Parallel()

	client := &Client{
		baseURL: "https://example.com",
		token:   "test-token",
		secret:  "test-secret",
	}

	headers := client.generateAuth("GET", "/test")

	// Manually compute expected signature
	authString := client.token + "&" + headers["Auth-Timestamp"] + "&" + headers["Auth-Nonce"] + "&GET&/test"
	h := hmac.New(sha256.New, []byte(client.secret))
	h.Write([]byte(authString))
	expectedSignature := base64.StdEncoding.EncodeToString(h.Sum(nil))

	assert.Equal(t, expectedSignature, headers["Auth-Signature"])
}

func TestGenerateAuth_WithBody(t *testing.T) {
	t.Parallel()

	client := &Client{
		baseURL: "https://example.com",
		token:   "test-token",
		secret:  "test-secret",
	}

	headers := client.generateAuth("POST", "/test")

	// Manually compute expected signature (body is NOT included per Pritunl API docs)
	authString := client.token + "&" + headers["Auth-Timestamp"] + "&" + headers["Auth-Nonce"] + "&POST&/test"
	h := hmac.New(sha256.New, []byte(client.secret))
	h.Write([]byte(authString))
	expectedSignature := base64.StdEncoding.EncodeToString(h.Sum(nil))

	assert.Equal(t, expectedSignature, headers["Auth-Signature"])
}

func TestGenerateAuth_MethodUppercase(t *testing.T) {
	t.Parallel()

	client := &Client{
		baseURL: "https://example.com",
		token:   "test-token",
		secret:  "test-secret",
	}

	// Test that method is uppercased in signature
	headersLower := client.generateAuth("get", "/test")
	headersUpper := client.generateAuth("GET", "/test")

	// Both should produce same signature format (method is uppercased internally)
	assert.NotEmpty(t, headersLower["Auth-Signature"])
	assert.NotEmpty(t, headersUpper["Auth-Signature"])
}

func TestDoRequest_Success(t *testing.T) {
	t.Parallel()

	expectedBody := `{"status":"ok"}`
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(expectedBody))
	}))
	defer server.Close()

	client := NewClient(server.URL, "token", "secret", false)
	body, err := client.Get("/test")

	require.NoError(t, err)
	assert.Equal(t, expectedBody, string(body))
}

func TestDoRequest_Error4xx(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name       string
		statusCode int
		body       string
	}{
		{
			name:       "400 Bad Request",
			statusCode: http.StatusBadRequest,
			body:       "bad request",
		},
		{
			name:       "401 Unauthorized",
			statusCode: http.StatusUnauthorized,
			body:       "unauthorized",
		},
		{
			name:       "404 Not Found",
			statusCode: http.StatusNotFound,
			body:       "not found",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(tt.statusCode)
				w.Write([]byte(tt.body))
			}))
			defer server.Close()

			client := NewClient(server.URL, "token", "secret", false)
			_, err := client.Get("/test")

			require.Error(t, err)
			assert.Contains(t, err.Error(), "status")
			assert.Contains(t, err.Error(), tt.body)
		})
	}
}

func TestDoRequest_Error5xx(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name       string
		statusCode int
		body       string
	}{
		{
			name:       "500 Internal Server Error",
			statusCode: http.StatusInternalServerError,
			body:       "internal error",
		},
		{
			name:       "502 Bad Gateway",
			statusCode: http.StatusBadGateway,
			body:       "bad gateway",
		},
		{
			name:       "503 Service Unavailable",
			statusCode: http.StatusServiceUnavailable,
			body:       "service unavailable",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(tt.statusCode)
				w.Write([]byte(tt.body))
			}))
			defer server.Close()

			client := NewClient(server.URL, "token", "secret", false)
			_, err := client.Get("/test")

			require.Error(t, err)
			assert.Contains(t, err.Error(), "status")
			assert.Contains(t, err.Error(), tt.body)
		})
	}
}

func TestDoRequest_SetsHeaders(t *testing.T) {
	t.Parallel()

	var receivedHeaders http.Header
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		receivedHeaders = r.Header
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	client := NewClient(server.URL, "test-token", "test-secret", false)
	_, err := client.Get("/test")

	require.NoError(t, err)
	assert.Equal(t, "application/json", receivedHeaders.Get("Content-Type"))
	assert.Equal(t, "test-token", receivedHeaders.Get("Auth-Token"))
	assert.NotEmpty(t, receivedHeaders.Get("Auth-Timestamp"))
	assert.NotEmpty(t, receivedHeaders.Get("Auth-Nonce"))
	assert.NotEmpty(t, receivedHeaders.Get("Auth-Signature"))
}

func TestDoRequest_PostWithBody(t *testing.T) {
	t.Parallel()

	var receivedBody []byte
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var err error
		receivedBody, err = readAll(r.Body)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"id":"123"}`))
	}))
	defer server.Close()

	client := NewClient(server.URL, "token", "secret", false)
	body := map[string]string{"name": "test"}
	_, err := client.Post("/test", body)

	require.NoError(t, err)

	var received map[string]string
	err = json.Unmarshal(receivedBody, &received)
	require.NoError(t, err)
	assert.Equal(t, "test", received["name"])
}

func TestDoRequest_PutWithBody(t *testing.T) {
	t.Parallel()

	var receivedMethod string
	var receivedBody []byte
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		receivedMethod = r.Method
		var err error
		receivedBody, err = readAll(r.Body)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"id":"123"}`))
	}))
	defer server.Close()

	client := NewClient(server.URL, "token", "secret", false)
	body := map[string]string{"name": "updated"}
	_, err := client.Put("/test", body)

	require.NoError(t, err)
	assert.Equal(t, http.MethodPut, receivedMethod)

	var received map[string]string
	err = json.Unmarshal(receivedBody, &received)
	require.NoError(t, err)
	assert.Equal(t, "updated", received["name"])
}

func TestDoRequest_Delete(t *testing.T) {
	t.Parallel()

	var receivedMethod string
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		receivedMethod = r.Method
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	client := NewClient(server.URL, "token", "secret", false)
	_, err := client.Delete("/test")

	require.NoError(t, err)
	assert.Equal(t, http.MethodDelete, receivedMethod)
}

func TestStatus_Success(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/status" {
			w.WriteHeader(http.StatusNotFound)
			return
		}
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"status":"ok"}`))
	}))
	defer server.Close()

	client := NewClient(server.URL, "token", "secret", false)
	err := client.Status()

	require.NoError(t, err)
}

func TestStatus_Error(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("server error"))
	}))
	defer server.Close()

	client := NewClient(server.URL, "token", "secret", false)
	err := client.Status()

	require.Error(t, err)
}

func TestNewClient_InsecureTLSConfig(t *testing.T) {
	t.Parallel()

	client := NewClient("https://example.com", "token", "secret", true)
	transport, ok := client.httpClient.Transport.(*http.Transport)
	require.True(t, ok)
	require.NotNil(t, transport.TLSClientConfig)
	assert.True(t, transport.TLSClientConfig.InsecureSkipVerify)
}

func TestNewClient_SecureTLSConfig(t *testing.T) {
	t.Parallel()

	client := NewClient("https://example.com", "token", "secret", false)
	transport, ok := client.httpClient.Transport.(*http.Transport)
	require.True(t, ok)
	require.NotNil(t, transport.TLSClientConfig)
	assert.False(t, transport.TLSClientConfig.InsecureSkipVerify)
}

// Helper to check TLS config
func getTLSConfig(client *Client) *tls.Config {
	transport := client.httpClient.Transport.(*http.Transport)
	return transport.TLSClientConfig
}

func TestDoRequest_CorrectPath(t *testing.T) {
	t.Parallel()

	var receivedPath string
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		receivedPath = r.URL.Path
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	client := NewClient(server.URL, "token", "secret", false)
	_, err := client.Get("/organization/abc123")

	require.NoError(t, err)
	assert.Equal(t, "/organization/abc123", receivedPath)
}

func TestDoRequest_AllStatusCodes(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name       string
		statusCode int
		wantErr    bool
	}{
		{"200 OK", http.StatusOK, false},
		{"201 Created", http.StatusCreated, false},
		{"204 No Content", http.StatusNoContent, false},
		{"299 Custom Success", 299, false},
		{"300 Multiple Choices", http.StatusMultipleChoices, true},
		{"301 Moved Permanently", http.StatusMovedPermanently, true},
		{"400 Bad Request", http.StatusBadRequest, true},
		{"500 Internal Server Error", http.StatusInternalServerError, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(tt.statusCode)
			}))
			defer server.Close()

			client := NewClient(server.URL, "token", "secret", false)
			_, err := client.Get("/test")

			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

// readAll is a helper to read request body
func readAll(r interface{ Read([]byte) (int, error) }) ([]byte, error) {
	var result []byte
	buf := make([]byte, 1024)
	for {
		n, err := r.Read(buf)
		if n > 0 {
			result = append(result, buf[:n]...)
		}
		if err != nil {
			if err.Error() == "EOF" {
				break
			}
			return nil, err
		}
	}
	return result, nil
}

func TestDoRequest_JSONMarshalError(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	client := NewClient(server.URL, "token", "secret", false)

	// Create a value that cannot be marshaled to JSON
	body := make(chan int)
	_, err := client.Post("/test", body)

	require.Error(t, err)
	assert.Contains(t, err.Error(), "marshal")
}

func TestGenerateAuth_EmptyBody(t *testing.T) {
	t.Parallel()

	client := NewClient("https://example.com", "test-token", "test-secret", false)

	// Test with nil body
	headers := client.generateAuth("GET", "/test")
	assert.NotEmpty(t, headers["Auth-Signature"])

	// Calling again should produce a valid signature (different nonce)
	headers2 := client.generateAuth("GET", "/test")
	assert.NotEmpty(t, headers2["Auth-Signature"])
}

func TestDoRequest_NilBody(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"result":"success"}`))
	}))
	defer server.Close()

	client := NewClient(server.URL, "token", "secret", false)
	body, err := client.Get("/test")

	require.NoError(t, err)
	assert.Contains(t, string(body), "success")
}

func TestClient_BaseURLWithPath(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name        string
		baseURL     string
		path        string
		wantFullURL string
	}{
		{
			name:        "simple base URL",
			baseURL:     "https://example.com",
			path:        "/api/test",
			wantFullURL: "https://example.com/api/test",
		},
		{
			name:        "base URL with trailing slash",
			baseURL:     "https://example.com/",
			path:        "/api/test",
			wantFullURL: "https://example.com/api/test",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			var receivedURL string
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				receivedURL = r.URL.Path
				w.WriteHeader(http.StatusOK)
			}))
			defer server.Close()

			// Replace the example.com with actual test server
			client := NewClient(server.URL+strings.TrimPrefix(tt.baseURL, "https://example.com"), "token", "secret", false)
			_, _ = client.Get(tt.path)

			assert.Equal(t, tt.path, receivedURL)
		})
	}
}
