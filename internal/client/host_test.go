package client

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestListHosts_Success(t *testing.T) {
	t.Parallel()

	hosts := []Host{
		{ID: "host1", Name: "vpn-server-1", Hostname: "vpn1.example.com", Status: "online"},
		{ID: "host2", Name: "vpn-server-2", Hostname: "vpn2.example.com", Status: "offline"},
	}

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/host", r.URL.Path)
		assert.Equal(t, http.MethodGet, r.Method)

		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(hosts)
	}))
	defer server.Close()

	client := NewClient(server.URL, "token", "secret", false)
	result, err := client.ListHosts()

	require.NoError(t, err)
	require.Len(t, result, 2)
	assert.Equal(t, "host1", result[0].ID)
	assert.Equal(t, "vpn-server-1", result[0].Name)
	assert.Equal(t, "vpn1.example.com", result[0].Hostname)
	assert.Equal(t, "online", result[0].Status)
}

func TestListHosts_Empty(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("[]"))
	}))
	defer server.Close()

	client := NewClient(server.URL, "token", "secret", false)
	result, err := client.ListHosts()

	require.NoError(t, err)
	assert.Empty(t, result)
}

func TestListHosts_Error(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("server error"))
	}))
	defer server.Close()

	client := NewClient(server.URL, "token", "secret", false)
	_, err := client.ListHosts()

	require.Error(t, err)
}

func TestListHosts_InvalidJSON(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("not json"))
	}))
	defer server.Close()

	client := NewClient(server.URL, "token", "secret", false)
	_, err := client.ListHosts()

	require.Error(t, err)
	assert.Contains(t, err.Error(), "unmarshal")
}

func TestGetHost_Success(t *testing.T) {
	t.Parallel()

	host := Host{
		ID:           "host123",
		Name:         "vpn-server",
		Hostname:     "vpn.example.com",
		Status:       "online",
		Uptime:       86400,
		UsersOnline:  5,
		PublicAddr:   "1.2.3.4",
		PublicAddr6:  "2001:db8::1",
		LocalAddr:    "10.0.0.1",
		Availability: "high",
		CPUUsage:     25.5,
		MemUsage:     60.2,
		Version:      "1.32.3805.95",
	}

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/host/host123", r.URL.Path)
		assert.Equal(t, http.MethodGet, r.Method)

		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(host)
	}))
	defer server.Close()

	client := NewClient(server.URL, "token", "secret", false)
	result, err := client.GetHost("host123")

	require.NoError(t, err)
	assert.Equal(t, "host123", result.ID)
	assert.Equal(t, "vpn-server", result.Name)
	assert.Equal(t, "vpn.example.com", result.Hostname)
	assert.Equal(t, "online", result.Status)
	assert.Equal(t, int64(86400), result.Uptime)
	assert.Equal(t, 5, result.UsersOnline)
	assert.Equal(t, "1.2.3.4", result.PublicAddr)
	assert.InDelta(t, 25.5, result.CPUUsage, 0.001)
	assert.InDelta(t, 60.2, result.MemUsage, 0.001)
}

func TestGetHost_NotFound(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
		w.Write([]byte("host not found"))
	}))
	defer server.Close()

	client := NewClient(server.URL, "token", "secret", false)
	_, err := client.GetHost("nonexistent")

	require.Error(t, err)
	assert.Contains(t, err.Error(), "404")
}

func TestGetHost_InvalidJSON(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("invalid json"))
	}))
	defer server.Close()

	client := NewClient(server.URL, "token", "secret", false)
	_, err := client.GetHost("host123")

	require.Error(t, err)
	assert.Contains(t, err.Error(), "unmarshal")
}

func TestGetHostByName_Found(t *testing.T) {
	t.Parallel()

	hosts := []Host{
		{ID: "host1", Name: "vpn-server-1"},
		{ID: "host2", Name: "vpn-server-2"},
		{ID: "host3", Name: "vpn-server-3"},
	}

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/host", r.URL.Path)
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(hosts)
	}))
	defer server.Close()

	client := NewClient(server.URL, "token", "secret", false)
	result, err := client.GetHostByName("vpn-server-2")

	require.NoError(t, err)
	assert.Equal(t, "host2", result.ID)
	assert.Equal(t, "vpn-server-2", result.Name)
}

func TestGetHostByName_NotFound(t *testing.T) {
	t.Parallel()

	hosts := []Host{
		{ID: "host1", Name: "vpn-server-1"},
		{ID: "host2", Name: "vpn-server-2"},
	}

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(hosts)
	}))
	defer server.Close()

	client := NewClient(server.URL, "token", "secret", false)
	_, err := client.GetHostByName("nonexistent")

	require.Error(t, err)
	assert.Contains(t, err.Error(), "not found")
	assert.Contains(t, err.Error(), "nonexistent")
}

func TestGetHostByName_ListError(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("server error"))
	}))
	defer server.Close()

	client := NewClient(server.URL, "token", "secret", false)
	_, err := client.GetHostByName("vpn-server")

	require.Error(t, err)
}

func TestHost_AllFields(t *testing.T) {
	t.Parallel()

	host := Host{
		ID:              "host123",
		Name:            "vpn-server",
		Hostname:        "vpn.example.com",
		Status:          "online",
		Uptime:          172800,
		UsersOnline:     10,
		PublicAddr:      "1.2.3.4",
		PublicAddr6:     "2001:db8::1",
		RoutedSubnet6:   "2001:db8:1::/48",
		RoutedSubnet6WG: "2001:db8:2::/48",
		LocalAddr:       "10.0.0.1",
		LocalAddr6:      "fd00::1",
		LinkAddr:        "10.0.0.1",
		SyncAddress:     "10.0.0.1:8080",
		Availability:    "high",
		AvailabilityG:   "default",
		CPUUsage:        45.5,
		MemUsage:        75.2,
		Version:         "1.32.3805.95",
	}

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(host)
	}))
	defer server.Close()

	client := NewClient(server.URL, "token", "secret", false)
	result, err := client.GetHost("host123")

	require.NoError(t, err)
	assert.Equal(t, host.ID, result.ID)
	assert.Equal(t, host.Name, result.Name)
	assert.Equal(t, host.Hostname, result.Hostname)
	assert.Equal(t, host.Status, result.Status)
	assert.Equal(t, host.Uptime, result.Uptime)
	assert.Equal(t, host.UsersOnline, result.UsersOnline)
	assert.Equal(t, host.PublicAddr, result.PublicAddr)
	assert.Equal(t, host.PublicAddr6, result.PublicAddr6)
	assert.Equal(t, host.RoutedSubnet6, result.RoutedSubnet6)
	assert.Equal(t, host.RoutedSubnet6WG, result.RoutedSubnet6WG)
	assert.Equal(t, host.LocalAddr, result.LocalAddr)
	assert.Equal(t, host.LocalAddr6, result.LocalAddr6)
	assert.Equal(t, host.LinkAddr, result.LinkAddr)
	assert.Equal(t, host.SyncAddress, result.SyncAddress)
	assert.Equal(t, host.Availability, result.Availability)
	assert.Equal(t, host.AvailabilityG, result.AvailabilityG)
	assert.InDelta(t, host.CPUUsage, result.CPUUsage, 0.001)
	assert.InDelta(t, host.MemUsage, result.MemUsage, 0.001)
	assert.Equal(t, host.Version, result.Version)
}

func TestGetHostByName_FirstMatch(t *testing.T) {
	t.Parallel()

	// Test that GetHostByName returns the first matching host
	hosts := []Host{
		{ID: "host1", Name: "server-a"},
		{ID: "host2", Name: "target-host"},
		{ID: "host3", Name: "server-b"},
	}

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(hosts)
	}))
	defer server.Close()

	client := NewClient(server.URL, "token", "secret", false)
	result, err := client.GetHostByName("target-host")

	require.NoError(t, err)
	assert.Equal(t, "host2", result.ID)
	assert.Equal(t, "target-host", result.Name)
}

func TestListHosts_MultipleHosts(t *testing.T) {
	t.Parallel()

	hosts := []Host{
		{ID: "host1", Name: "primary", Status: "online", CPUUsage: 10.0, MemUsage: 50.0},
		{ID: "host2", Name: "secondary", Status: "online", CPUUsage: 15.0, MemUsage: 55.0},
		{ID: "host3", Name: "backup", Status: "offline", CPUUsage: 0.0, MemUsage: 0.0},
	}

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(hosts)
	}))
	defer server.Close()

	client := NewClient(server.URL, "token", "secret", false)
	result, err := client.ListHosts()

	require.NoError(t, err)
	require.Len(t, result, 3)

	// Verify order is preserved
	assert.Equal(t, "host1", result[0].ID)
	assert.Equal(t, "host2", result[1].ID)
	assert.Equal(t, "host3", result[2].ID)

	// Check status
	assert.Equal(t, "online", result[0].Status)
	assert.Equal(t, "online", result[1].Status)
	assert.Equal(t, "offline", result[2].Status)
}

func TestGetHost_WithZeroValues(t *testing.T) {
	t.Parallel()

	// Test host with zero/empty values
	host := Host{
		ID:       "host123",
		Name:     "minimal-host",
		Status:   "offline",
		CPUUsage: 0.0,
		MemUsage: 0.0,
	}

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(host)
	}))
	defer server.Close()

	client := NewClient(server.URL, "token", "secret", false)
	result, err := client.GetHost("host123")

	require.NoError(t, err)
	assert.Equal(t, "host123", result.ID)
	assert.Equal(t, "minimal-host", result.Name)
	assert.Equal(t, "offline", result.Status)
	assert.InDelta(t, 0.0, result.CPUUsage, 0.001)
	assert.InDelta(t, 0.0, result.MemUsage, 0.001)
	assert.Empty(t, result.PublicAddr)
	assert.Empty(t, result.Version)
}
