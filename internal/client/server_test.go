package client

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestListServers_Success(t *testing.T) {
	t.Parallel()

	servers := []Server{
		{ID: "srv1", Name: "Production", Network: "10.10.0.0/24", Port: 1194, Protocol: "udp"},
		{ID: "srv2", Name: "Development", Network: "10.20.0.0/24", Port: 1195, Protocol: "tcp"},
	}

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/server", r.URL.Path)
		assert.Equal(t, http.MethodGet, r.Method)

		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(servers)
	}))
	defer server.Close()

	client := NewClient(server.URL, "token", "secret", false)
	result, err := client.ListServers()

	require.NoError(t, err)
	require.Len(t, result, 2)
	assert.Equal(t, "srv1", result[0].ID)
	assert.Equal(t, "Production", result[0].Name)
	assert.Equal(t, "10.10.0.0/24", result[0].Network)
	assert.Equal(t, 1194, result[0].Port)
	assert.Equal(t, "udp", result[0].Protocol)
}

func TestListServers_Empty(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("[]"))
	}))
	defer server.Close()

	client := NewClient(server.URL, "token", "secret", false)
	result, err := client.ListServers()

	require.NoError(t, err)
	assert.Empty(t, result)
}

func TestListServers_Error(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("server error"))
	}))
	defer server.Close()

	client := NewClient(server.URL, "token", "secret", false)
	_, err := client.ListServers()

	require.Error(t, err)
}

func TestListServers_InvalidJSON(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("not json"))
	}))
	defer server.Close()

	client := NewClient(server.URL, "token", "secret", false)
	_, err := client.ListServers()

	require.Error(t, err)
	assert.Contains(t, err.Error(), "unmarshal")
}

func TestGetServer_Success(t *testing.T) {
	t.Parallel()

	srv := Server{
		ID:           "srv123",
		Name:         "Production",
		Status:       "online",
		Network:      "10.10.0.0/24",
		Port:         1194,
		Protocol:     "udp",
		Cipher:       "aes256",
		Hash:         "sha256",
		InterClient:  true,
		PingInterval: 10,
		PingTimeout:  60,
	}

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/server/srv123", r.URL.Path)
		assert.Equal(t, http.MethodGet, r.Method)

		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(srv)
	}))
	defer server.Close()

	client := NewClient(server.URL, "token", "secret", false)
	result, err := client.GetServer("srv123")

	require.NoError(t, err)
	assert.Equal(t, "srv123", result.ID)
	assert.Equal(t, "Production", result.Name)
	assert.Equal(t, "online", result.Status)
	assert.Equal(t, "10.10.0.0/24", result.Network)
	assert.True(t, result.InterClient)
}

func TestGetServer_NotFound(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
		w.Write([]byte("server not found"))
	}))
	defer server.Close()

	client := NewClient(server.URL, "token", "secret", false)
	_, err := client.GetServer("nonexistent")

	require.Error(t, err)
	assert.Contains(t, err.Error(), "404")
}

func TestGetServer_InvalidJSON(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("invalid json"))
	}))
	defer server.Close()

	client := NewClient(server.URL, "token", "secret", false)
	_, err := client.GetServer("srv123")

	require.Error(t, err)
	assert.Contains(t, err.Error(), "unmarshal")
}

func TestGetServerByName_Found(t *testing.T) {
	t.Parallel()

	servers := []Server{
		{ID: "srv1", Name: "Production"},
		{ID: "srv2", Name: "Development"},
		{ID: "srv3", Name: "Staging"},
	}

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/server", r.URL.Path)
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(servers)
	}))
	defer server.Close()

	client := NewClient(server.URL, "token", "secret", false)
	result, err := client.GetServerByName("Development")

	require.NoError(t, err)
	assert.Equal(t, "srv2", result.ID)
	assert.Equal(t, "Development", result.Name)
}

func TestGetServerByName_NotFound(t *testing.T) {
	t.Parallel()

	servers := []Server{
		{ID: "srv1", Name: "Production"},
		{ID: "srv2", Name: "Development"},
	}

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(servers)
	}))
	defer server.Close()

	client := NewClient(server.URL, "token", "secret", false)
	_, err := client.GetServerByName("NonExistent")

	require.Error(t, err)
	assert.Contains(t, err.Error(), "not found")
	assert.Contains(t, err.Error(), "NonExistent")
}

func TestGetServerByName_ListError(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("server error"))
	}))
	defer server.Close()

	client := NewClient(server.URL, "token", "secret", false)
	_, err := client.GetServerByName("Production")

	require.Error(t, err)
}

func TestCreateServer_Success(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/server", r.URL.Path)
		assert.Equal(t, http.MethodPost, r.Method)

		var input Server
		err := json.NewDecoder(r.Body).Decode(&input)
		require.NoError(t, err)
		assert.Equal(t, "NewServer", input.Name)
		assert.Equal(t, "10.30.0.0/24", input.Network)

		created := Server{
			ID:       "new-srv-id",
			Name:     input.Name,
			Network:  input.Network,
			Port:     input.Port,
			Protocol: input.Protocol,
		}
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(created)
	}))
	defer server.Close()

	client := NewClient(server.URL, "token", "secret", false)
	srv := &Server{
		Name:     "NewServer",
		Network:  "10.30.0.0/24",
		Port:     1194,
		Protocol: "udp",
	}
	result, err := client.CreateServer(srv)

	require.NoError(t, err)
	assert.Equal(t, "new-srv-id", result.ID)
	assert.Equal(t, "NewServer", result.Name)
}

func TestCreateServer_Error(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte("invalid server"))
	}))
	defer server.Close()

	client := NewClient(server.URL, "token", "secret", false)
	srv := &Server{Name: ""}
	_, err := client.CreateServer(srv)

	require.Error(t, err)
}

func TestCreateServer_InvalidResponseJSON(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("not json"))
	}))
	defer server.Close()

	client := NewClient(server.URL, "token", "secret", false)
	srv := &Server{Name: "NewServer"}
	_, err := client.CreateServer(srv)

	require.Error(t, err)
	assert.Contains(t, err.Error(), "unmarshal")
}

func TestUpdateServer_Success(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/server/srv123", r.URL.Path)
		assert.Equal(t, http.MethodPut, r.Method)

		var input Server
		err := json.NewDecoder(r.Body).Decode(&input)
		require.NoError(t, err)
		assert.Equal(t, "UpdatedServer", input.Name)

		updated := Server{
			ID:   "srv123",
			Name: input.Name,
		}
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(updated)
	}))
	defer server.Close()

	client := NewClient(server.URL, "token", "secret", false)
	srv := &Server{Name: "UpdatedServer"}
	result, err := client.UpdateServer("srv123", srv)

	require.NoError(t, err)
	assert.Equal(t, "srv123", result.ID)
	assert.Equal(t, "UpdatedServer", result.Name)
}

func TestUpdateServer_NotFound(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
		w.Write([]byte("server not found"))
	}))
	defer server.Close()

	client := NewClient(server.URL, "token", "secret", false)
	srv := &Server{Name: "UpdatedServer"}
	_, err := client.UpdateServer("nonexistent", srv)

	require.Error(t, err)
	assert.Contains(t, err.Error(), "404")
}

func TestUpdateServer_InvalidResponseJSON(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("not json"))
	}))
	defer server.Close()

	client := NewClient(server.URL, "token", "secret", false)
	srv := &Server{Name: "UpdatedServer"}
	_, err := client.UpdateServer("srv123", srv)

	require.Error(t, err)
	assert.Contains(t, err.Error(), "unmarshal")
}

func TestDeleteServer_Success(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/server/srv123", r.URL.Path)
		assert.Equal(t, http.MethodDelete, r.Method)

		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	client := NewClient(server.URL, "token", "secret", false)
	err := client.DeleteServer("srv123")

	require.NoError(t, err)
}

func TestDeleteServer_NotFound(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
		w.Write([]byte("server not found"))
	}))
	defer server.Close()

	client := NewClient(server.URL, "token", "secret", false)
	err := client.DeleteServer("nonexistent")

	require.Error(t, err)
	assert.Contains(t, err.Error(), "404")
}

func TestGetServerOrganizations_Success(t *testing.T) {
	t.Parallel()

	orgs := []ServerOrganization{
		{ID: "org1", Name: "Engineering", Server: "srv123"},
		{ID: "org2", Name: "Marketing", Server: "srv123"},
	}

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/server/srv123/organization", r.URL.Path)
		assert.Equal(t, http.MethodGet, r.Method)

		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(orgs)
	}))
	defer server.Close()

	client := NewClient(server.URL, "token", "secret", false)
	result, err := client.GetServerOrganizations("srv123")

	require.NoError(t, err)
	require.Len(t, result, 2)
	assert.Equal(t, "org1", result[0].ID)
	assert.Equal(t, "Engineering", result[0].Name)
}

func TestGetServerOrganizations_Empty(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("[]"))
	}))
	defer server.Close()

	client := NewClient(server.URL, "token", "secret", false)
	result, err := client.GetServerOrganizations("srv123")

	require.NoError(t, err)
	assert.Empty(t, result)
}

func TestGetServerOrganizations_InvalidJSON(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("not json"))
	}))
	defer server.Close()

	client := NewClient(server.URL, "token", "secret", false)
	_, err := client.GetServerOrganizations("srv123")

	require.Error(t, err)
	assert.Contains(t, err.Error(), "unmarshal")
}

func TestAttachOrganization_Success(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/server/srv123/organization/org456", r.URL.Path)
		assert.Equal(t, http.MethodPut, r.Method)

		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	client := NewClient(server.URL, "token", "secret", false)
	err := client.AttachOrganization("srv123", "org456")

	require.NoError(t, err)
}

func TestAttachOrganization_Error(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte("cannot attach"))
	}))
	defer server.Close()

	client := NewClient(server.URL, "token", "secret", false)
	err := client.AttachOrganization("srv123", "org456")

	require.Error(t, err)
}

func TestDetachOrganization_Success(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/server/srv123/organization/org456", r.URL.Path)
		assert.Equal(t, http.MethodDelete, r.Method)

		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	client := NewClient(server.URL, "token", "secret", false)
	err := client.DetachOrganization("srv123", "org456")

	require.NoError(t, err)
}

func TestDetachOrganization_Error(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte("cannot detach"))
	}))
	defer server.Close()

	client := NewClient(server.URL, "token", "secret", false)
	err := client.DetachOrganization("srv123", "org456")

	require.Error(t, err)
}

func TestStartServer_Success(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/server/srv123/operation/start", r.URL.Path)
		assert.Equal(t, http.MethodPut, r.Method)

		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	client := NewClient(server.URL, "token", "secret", false)
	err := client.StartServer("srv123")

	require.NoError(t, err)
}

func TestStartServer_Error(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte("cannot start"))
	}))
	defer server.Close()

	client := NewClient(server.URL, "token", "secret", false)
	err := client.StartServer("srv123")

	require.Error(t, err)
}

func TestStopServer_Success(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/server/srv123/operation/stop", r.URL.Path)
		assert.Equal(t, http.MethodPut, r.Method)

		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	client := NewClient(server.URL, "token", "secret", false)
	err := client.StopServer("srv123")

	require.NoError(t, err)
}

func TestStopServer_Error(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte("cannot stop"))
	}))
	defer server.Close()

	client := NewClient(server.URL, "token", "secret", false)
	err := client.StopServer("srv123")

	require.Error(t, err)
}

func TestServer_AllFields(t *testing.T) {
	t.Parallel()

	srv := Server{
		ID:               "srv123",
		Name:             "Test Server",
		Status:           "online",
		Uptime:           86400,
		UsersOnline:      5,
		DevicesOnline:    10,
		UserCount:        20,
		Network:          "10.10.0.0/24",
		NetworkWG:        "10.11.0.0/24",
		NetworkMode:      "tunnel",
		NetworkStart:     "10.10.0.10",
		NetworkEnd:       "10.10.0.250",
		RestrictRoutes:   true,
		IPv6:             false,
		IPv6Firewall:     false,
		BindAddress:      "0.0.0.0",
		Protocol:         "udp",
		Port:             1194,
		PortWG:           51820,
		DNSServers:       []string{"8.8.8.8", "8.8.4.4"},
		SearchDomain:     "example.com",
		InterClient:      true,
		PingInterval:     10,
		PingTimeout:      60,
		LinkPingInterval: 1,
		LinkPingTimeout:  5,
		InactiveTimeout:  0,
		SessionTimeout:   0,
		AllowedDevices:   "mobile",
		MaxClients:       100,
		MaxDevices:       3,
		ReplicaCount:     1,
		VxLan:            false,
		DNSMapping:       true,
		Debug:            false,
		OtpAuth:          false,
		LzoCompression:   false,
		Cipher:           "aes256",
		Hash:             "sha256",
		BlockOutsideDNS:  false,
		JumboFrames:      false,
		PreConnectMsg:    "Welcome",
		MSFixTimeout:     1400,
		Multihome:        false,
		Groups:           []string{"admin", "users"},
		Organizations:    []string{"org1", "org2"},
		WG:               false,
	}

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(srv)
	}))
	defer server.Close()

	client := NewClient(server.URL, "token", "secret", false)
	result, err := client.GetServer("srv123")

	require.NoError(t, err)
	assert.Equal(t, srv.ID, result.ID)
	assert.Equal(t, srv.Name, result.Name)
	assert.Equal(t, srv.Status, result.Status)
	assert.Equal(t, srv.Network, result.Network)
	assert.Equal(t, srv.DNSServers, result.DNSServers)
	assert.Equal(t, srv.Groups, result.Groups)
	assert.Equal(t, srv.Organizations, result.Organizations)
}
