package client

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestListRoutes_Success(t *testing.T) {
	t.Parallel()

	routes := []Route{
		{ID: "route1", Network: "10.0.0.0/8", Comment: "Private network", Nat: true},
		{ID: "route2", Network: "192.168.0.0/16", Comment: "Local network", Nat: false},
	}

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/server/srv123/route", r.URL.Path)
		assert.Equal(t, http.MethodGet, r.Method)

		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(routes)
	}))
	defer server.Close()

	client := NewClient(server.URL, "token", "secret", false)
	result, err := client.ListRoutes("srv123")

	require.NoError(t, err)
	require.Len(t, result, 2)
	assert.Equal(t, "route1", result[0].ID)
	assert.Equal(t, "10.0.0.0/8", result[0].Network)
	assert.Equal(t, "Private network", result[0].Comment)
	assert.True(t, result[0].Nat)
}

func TestListRoutes_Empty(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("[]"))
	}))
	defer server.Close()

	client := NewClient(server.URL, "token", "secret", false)
	result, err := client.ListRoutes("srv123")

	require.NoError(t, err)
	assert.Empty(t, result)
}

func TestListRoutes_Error(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("server error"))
	}))
	defer server.Close()

	client := NewClient(server.URL, "token", "secret", false)
	_, err := client.ListRoutes("srv123")

	require.Error(t, err)
}

func TestListRoutes_InvalidJSON(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("not json"))
	}))
	defer server.Close()

	client := NewClient(server.URL, "token", "secret", false)
	_, err := client.ListRoutes("srv123")

	require.Error(t, err)
	assert.Contains(t, err.Error(), "unmarshal")
}

func TestGetRoute_Success(t *testing.T) {
	t.Parallel()

	routes := []Route{
		{ID: "route1", Network: "10.0.0.0/8", Comment: "Route 1"},
		{ID: "route2", Network: "192.168.0.0/16", Comment: "Route 2"},
		{ID: "route3", Network: "172.16.0.0/12", Comment: "Route 3"},
	}

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/server/srv123/route", r.URL.Path)
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(routes)
	}))
	defer server.Close()

	client := NewClient(server.URL, "token", "secret", false)
	result, err := client.GetRoute("srv123", "route2")

	require.NoError(t, err)
	assert.Equal(t, "route2", result.ID)
	assert.Equal(t, "192.168.0.0/16", result.Network)
	assert.Equal(t, "Route 2", result.Comment)
}

func TestGetRoute_NotFound(t *testing.T) {
	t.Parallel()

	routes := []Route{
		{ID: "route1", Network: "10.0.0.0/8"},
		{ID: "route2", Network: "192.168.0.0/16"},
	}

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(routes)
	}))
	defer server.Close()

	client := NewClient(server.URL, "token", "secret", false)
	_, err := client.GetRoute("srv123", "nonexistent")

	require.Error(t, err)
	assert.Contains(t, err.Error(), "not found")
	assert.Contains(t, err.Error(), "nonexistent")
	assert.Contains(t, err.Error(), "srv123")
}

func TestGetRoute_ListError(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("server error"))
	}))
	defer server.Close()

	client := NewClient(server.URL, "token", "secret", false)
	_, err := client.GetRoute("srv123", "route1")

	require.Error(t, err)
}

func TestCreateRoute_Success(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/server/srv123/route", r.URL.Path)
		assert.Equal(t, http.MethodPost, r.Method)

		var input Route
		err := json.NewDecoder(r.Body).Decode(&input)
		require.NoError(t, err)
		assert.Equal(t, "10.100.0.0/16", input.Network)
		assert.Equal(t, "New route", input.Comment)
		assert.True(t, input.Nat)

		created := Route{
			ID:      "new-route-id",
			Network: input.Network,
			Comment: input.Comment,
			Nat:     input.Nat,
			Server:  "srv123",
		}
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(created)
	}))
	defer server.Close()

	client := NewClient(server.URL, "token", "secret", false)
	route := &Route{
		Network: "10.100.0.0/16",
		Comment: "New route",
		Nat:     true,
	}
	result, err := client.CreateRoute("srv123", route)

	require.NoError(t, err)
	assert.Equal(t, "new-route-id", result.ID)
	assert.Equal(t, "10.100.0.0/16", result.Network)
	assert.Equal(t, "srv123", result.Server)
}

func TestCreateRoute_Error(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte("invalid route"))
	}))
	defer server.Close()

	client := NewClient(server.URL, "token", "secret", false)
	route := &Route{Network: "invalid"}
	_, err := client.CreateRoute("srv123", route)

	require.Error(t, err)
}

func TestCreateRoute_InvalidResponseJSON(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("not json"))
	}))
	defer server.Close()

	client := NewClient(server.URL, "token", "secret", false)
	route := &Route{Network: "10.0.0.0/8"}
	_, err := client.CreateRoute("srv123", route)

	require.Error(t, err)
	assert.Contains(t, err.Error(), "unmarshal")
}

func TestUpdateRoute_Success(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/server/srv123/route/route456", r.URL.Path)
		assert.Equal(t, http.MethodPut, r.Method)

		var input Route
		err := json.NewDecoder(r.Body).Decode(&input)
		require.NoError(t, err)
		assert.Equal(t, "Updated comment", input.Comment)

		updated := Route{
			ID:      "route456",
			Network: input.Network,
			Comment: input.Comment,
		}
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(updated)
	}))
	defer server.Close()

	client := NewClient(server.URL, "token", "secret", false)
	route := &Route{
		Network: "10.0.0.0/8",
		Comment: "Updated comment",
	}
	result, err := client.UpdateRoute("srv123", "route456", route)

	require.NoError(t, err)
	assert.Equal(t, "route456", result.ID)
	assert.Equal(t, "Updated comment", result.Comment)
}

func TestUpdateRoute_NotFound(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
		w.Write([]byte("route not found"))
	}))
	defer server.Close()

	client := NewClient(server.URL, "token", "secret", false)
	route := &Route{Network: "10.0.0.0/8"}
	_, err := client.UpdateRoute("srv123", "nonexistent", route)

	require.Error(t, err)
	assert.Contains(t, err.Error(), "404")
}

func TestUpdateRoute_InvalidResponseJSON(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("not json"))
	}))
	defer server.Close()

	client := NewClient(server.URL, "token", "secret", false)
	route := &Route{Network: "10.0.0.0/8"}
	_, err := client.UpdateRoute("srv123", "route456", route)

	require.Error(t, err)
	assert.Contains(t, err.Error(), "unmarshal")
}

func TestDeleteRoute_Success(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/server/srv123/route/route456", r.URL.Path)
		assert.Equal(t, http.MethodDelete, r.Method)

		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	client := NewClient(server.URL, "token", "secret", false)
	err := client.DeleteRoute("srv123", "route456")

	require.NoError(t, err)
}

func TestDeleteRoute_NotFound(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
		w.Write([]byte("route not found"))
	}))
	defer server.Close()

	client := NewClient(server.URL, "token", "secret", false)
	err := client.DeleteRoute("srv123", "nonexistent")

	require.Error(t, err)
	assert.Contains(t, err.Error(), "404")
}

func TestDeleteRoute_ServerError(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("internal error"))
	}))
	defer server.Close()

	client := NewClient(server.URL, "token", "secret", false)
	err := client.DeleteRoute("srv123", "route456")

	require.Error(t, err)
}

func TestRoute_AllFields(t *testing.T) {
	t.Parallel()

	route := Route{
		ID:           "route123",
		Network:      "10.0.0.0/8",
		Comment:      "Test route",
		Metric:       100,
		Nat:          true,
		NatInterface: "eth0",
		NatNetmap:    "10.1.0.0/16",
		Advertise:    true,
		VpcRegion:    "us-east-1",
		VpcID:        "vpc-12345",
		NetGateway:   false,
		Server:       "srv123",
	}

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode([]Route{route})
	}))
	defer server.Close()

	client := NewClient(server.URL, "token", "secret", false)
	result, err := client.GetRoute("srv123", "route123")

	require.NoError(t, err)
	assert.Equal(t, route.ID, result.ID)
	assert.Equal(t, route.Network, result.Network)
	assert.Equal(t, route.Comment, result.Comment)
	assert.Equal(t, route.Metric, result.Metric)
	assert.Equal(t, route.Nat, result.Nat)
	assert.Equal(t, route.NatInterface, result.NatInterface)
	assert.Equal(t, route.NatNetmap, result.NatNetmap)
	assert.Equal(t, route.Advertise, result.Advertise)
	assert.Equal(t, route.VpcRegion, result.VpcRegion)
	assert.Equal(t, route.VpcID, result.VpcID)
	assert.Equal(t, route.NetGateway, result.NetGateway)
	assert.Equal(t, route.Server, result.Server)
}

func TestGetRoute_FirstMatch(t *testing.T) {
	t.Parallel()

	// Test that GetRoute returns the first matching route by ID
	routes := []Route{
		{ID: "route1", Network: "10.0.0.0/8"},
		{ID: "target", Network: "192.168.0.0/16"},
		{ID: "route3", Network: "172.16.0.0/12"},
	}

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(routes)
	}))
	defer server.Close()

	client := NewClient(server.URL, "token", "secret", false)
	result, err := client.GetRoute("srv123", "target")

	require.NoError(t, err)
	assert.Equal(t, "target", result.ID)
	assert.Equal(t, "192.168.0.0/16", result.Network)
}
