package client

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestListOrganizations_Success(t *testing.T) {
	t.Parallel()

	orgs := []Organization{
		{ID: "org1", Name: "Engineering", UserCount: 10},
		{ID: "org2", Name: "Marketing", UserCount: 5},
	}

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/organization", r.URL.Path)
		assert.Equal(t, http.MethodGet, r.Method)

		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(orgs)
	}))
	defer server.Close()

	client := NewClient(server.URL, "token", "secret", false)
	result, err := client.ListOrganizations()

	require.NoError(t, err)
	require.Len(t, result, 2)
	assert.Equal(t, "org1", result[0].ID)
	assert.Equal(t, "Engineering", result[0].Name)
	assert.Equal(t, 10, result[0].UserCount)
	assert.Equal(t, "org2", result[1].ID)
	assert.Equal(t, "Marketing", result[1].Name)
}

func TestListOrganizations_Empty(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("[]"))
	}))
	defer server.Close()

	client := NewClient(server.URL, "token", "secret", false)
	result, err := client.ListOrganizations()

	require.NoError(t, err)
	assert.Empty(t, result)
}

func TestListOrganizations_Error(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("server error"))
	}))
	defer server.Close()

	client := NewClient(server.URL, "token", "secret", false)
	_, err := client.ListOrganizations()

	require.Error(t, err)
}

func TestListOrganizations_InvalidJSON(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("not json"))
	}))
	defer server.Close()

	client := NewClient(server.URL, "token", "secret", false)
	_, err := client.ListOrganizations()

	require.Error(t, err)
	assert.Contains(t, err.Error(), "unmarshal")
}

func TestGetOrganization_Success(t *testing.T) {
	t.Parallel()

	org := Organization{
		ID:        "org123",
		Name:      "Engineering",
		AuthAPI:   true,
		UserCount: 15,
	}

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/organization/org123", r.URL.Path)
		assert.Equal(t, http.MethodGet, r.Method)

		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(org)
	}))
	defer server.Close()

	client := NewClient(server.URL, "token", "secret", false)
	result, err := client.GetOrganization("org123")

	require.NoError(t, err)
	assert.Equal(t, "org123", result.ID)
	assert.Equal(t, "Engineering", result.Name)
	assert.True(t, result.AuthAPI)
	assert.Equal(t, 15, result.UserCount)
}

func TestGetOrganization_NotFound(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
		w.Write([]byte("organization not found"))
	}))
	defer server.Close()

	client := NewClient(server.URL, "token", "secret", false)
	_, err := client.GetOrganization("nonexistent")

	require.Error(t, err)
	assert.Contains(t, err.Error(), "404")
}

func TestGetOrganization_InvalidJSON(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("invalid json"))
	}))
	defer server.Close()

	client := NewClient(server.URL, "token", "secret", false)
	_, err := client.GetOrganization("org123")

	require.Error(t, err)
	assert.Contains(t, err.Error(), "unmarshal")
}

func TestGetOrganizationByName_Found(t *testing.T) {
	t.Parallel()

	orgs := []Organization{
		{ID: "org1", Name: "Engineering"},
		{ID: "org2", Name: "Marketing"},
		{ID: "org3", Name: "Sales"},
	}

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/organization", r.URL.Path)
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(orgs)
	}))
	defer server.Close()

	client := NewClient(server.URL, "token", "secret", false)
	result, err := client.GetOrganizationByName("Marketing")

	require.NoError(t, err)
	assert.Equal(t, "org2", result.ID)
	assert.Equal(t, "Marketing", result.Name)
}

func TestGetOrganizationByName_NotFound(t *testing.T) {
	t.Parallel()

	orgs := []Organization{
		{ID: "org1", Name: "Engineering"},
		{ID: "org2", Name: "Marketing"},
	}

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(orgs)
	}))
	defer server.Close()

	client := NewClient(server.URL, "token", "secret", false)
	_, err := client.GetOrganizationByName("NonExistent")

	require.Error(t, err)
	assert.Contains(t, err.Error(), "not found")
	assert.Contains(t, err.Error(), "NonExistent")
}

func TestGetOrganizationByName_ListError(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("server error"))
	}))
	defer server.Close()

	client := NewClient(server.URL, "token", "secret", false)
	_, err := client.GetOrganizationByName("Engineering")

	require.Error(t, err)
}

func TestCreateOrganization_Success(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/organization", r.URL.Path)
		assert.Equal(t, http.MethodPost, r.Method)

		var input Organization
		err := json.NewDecoder(r.Body).Decode(&input)
		require.NoError(t, err)
		assert.Equal(t, "NewOrg", input.Name)

		created := Organization{
			ID:   "new-org-id",
			Name: input.Name,
		}
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(created)
	}))
	defer server.Close()

	client := NewClient(server.URL, "token", "secret", false)
	org := &Organization{Name: "NewOrg"}
	result, err := client.CreateOrganization(org)

	require.NoError(t, err)
	assert.Equal(t, "new-org-id", result.ID)
	assert.Equal(t, "NewOrg", result.Name)
}

func TestCreateOrganization_Error(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte("invalid organization"))
	}))
	defer server.Close()

	client := NewClient(server.URL, "token", "secret", false)
	org := &Organization{Name: ""}
	_, err := client.CreateOrganization(org)

	require.Error(t, err)
}

func TestCreateOrganization_InvalidResponseJSON(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("not json"))
	}))
	defer server.Close()

	client := NewClient(server.URL, "token", "secret", false)
	org := &Organization{Name: "NewOrg"}
	_, err := client.CreateOrganization(org)

	require.Error(t, err)
	assert.Contains(t, err.Error(), "unmarshal")
}

func TestUpdateOrganization_Success(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/organization/org123", r.URL.Path)
		assert.Equal(t, http.MethodPut, r.Method)

		var input Organization
		err := json.NewDecoder(r.Body).Decode(&input)
		require.NoError(t, err)
		assert.Equal(t, "UpdatedName", input.Name)

		updated := Organization{
			ID:   "org123",
			Name: input.Name,
		}
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(updated)
	}))
	defer server.Close()

	client := NewClient(server.URL, "token", "secret", false)
	org := &Organization{Name: "UpdatedName"}
	result, err := client.UpdateOrganization("org123", org)

	require.NoError(t, err)
	assert.Equal(t, "org123", result.ID)
	assert.Equal(t, "UpdatedName", result.Name)
}

func TestUpdateOrganization_NotFound(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
		w.Write([]byte("organization not found"))
	}))
	defer server.Close()

	client := NewClient(server.URL, "token", "secret", false)
	org := &Organization{Name: "UpdatedName"}
	_, err := client.UpdateOrganization("nonexistent", org)

	require.Error(t, err)
	assert.Contains(t, err.Error(), "404")
}

func TestUpdateOrganization_InvalidResponseJSON(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("not json"))
	}))
	defer server.Close()

	client := NewClient(server.URL, "token", "secret", false)
	org := &Organization{Name: "UpdatedName"}
	_, err := client.UpdateOrganization("org123", org)

	require.Error(t, err)
	assert.Contains(t, err.Error(), "unmarshal")
}

func TestDeleteOrganization_Success(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/organization/org123", r.URL.Path)
		assert.Equal(t, http.MethodDelete, r.Method)

		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	client := NewClient(server.URL, "token", "secret", false)
	err := client.DeleteOrganization("org123")

	require.NoError(t, err)
}

func TestDeleteOrganization_NotFound(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
		w.Write([]byte("organization not found"))
	}))
	defer server.Close()

	client := NewClient(server.URL, "token", "secret", false)
	err := client.DeleteOrganization("nonexistent")

	require.Error(t, err)
	assert.Contains(t, err.Error(), "404")
}

func TestDeleteOrganization_ServerError(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("internal error"))
	}))
	defer server.Close()

	client := NewClient(server.URL, "token", "secret", false)
	err := client.DeleteOrganization("org123")

	require.Error(t, err)
}

func TestOrganization_AllFields(t *testing.T) {
	t.Parallel()

	org := Organization{
		ID:         "org123",
		Name:       "Test Org",
		AuthAPI:    true,
		AuthToken:  "token123",
		AuthSecret: "secret123",
		UserCount:  42,
	}

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(org)
	}))
	defer server.Close()

	client := NewClient(server.URL, "token", "secret", false)
	result, err := client.GetOrganization("org123")

	require.NoError(t, err)
	assert.Equal(t, org.ID, result.ID)
	assert.Equal(t, org.Name, result.Name)
	assert.Equal(t, org.AuthAPI, result.AuthAPI)
	assert.Equal(t, org.AuthToken, result.AuthToken)
	assert.Equal(t, org.AuthSecret, result.AuthSecret)
	assert.Equal(t, org.UserCount, result.UserCount)
}
