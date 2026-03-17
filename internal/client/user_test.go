package client

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestListUsers_Success(t *testing.T) {
	t.Parallel()

	users := []User{
		{ID: "user1", Name: "john.doe", Email: "john@example.com"},
		{ID: "user2", Name: "jane.doe", Email: "jane@example.com"},
	}

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/user/org123", r.URL.Path)
		assert.Equal(t, http.MethodGet, r.Method)

		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(users)
	}))
	defer server.Close()

	client := NewClient(server.URL, "token", "secret", false)
	result, err := client.ListUsers("org123")

	require.NoError(t, err)
	require.Len(t, result, 2)
	assert.Equal(t, "user1", result[0].ID)
	assert.Equal(t, "john.doe", result[0].Name)
	assert.Equal(t, "john@example.com", result[0].Email)
}

func TestListUsers_Empty(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("[]"))
	}))
	defer server.Close()

	client := NewClient(server.URL, "token", "secret", false)
	result, err := client.ListUsers("org123")

	require.NoError(t, err)
	assert.Empty(t, result)
}

func TestListUsers_Error(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("server error"))
	}))
	defer server.Close()

	client := NewClient(server.URL, "token", "secret", false)
	_, err := client.ListUsers("org123")

	require.Error(t, err)
}

func TestListUsers_InvalidJSON(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("not json"))
	}))
	defer server.Close()

	client := NewClient(server.URL, "token", "secret", false)
	_, err := client.ListUsers("org123")

	require.Error(t, err)
	assert.Contains(t, err.Error(), "unmarshal")
}

func TestGetUser_Success(t *testing.T) {
	t.Parallel()

	user := User{
		ID:           "user123",
		Name:         "john.doe",
		Email:        "john@example.com",
		Type:         "client",
		AuthType:     "local",
		Disabled:     false,
		Groups:       []string{"admin", "users"},
		Organization: "org123",
	}

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/user/org123/user123", r.URL.Path)
		assert.Equal(t, http.MethodGet, r.Method)

		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(user)
	}))
	defer server.Close()

	client := NewClient(server.URL, "token", "secret", false)
	result, err := client.GetUser("org123", "user123")

	require.NoError(t, err)
	assert.Equal(t, "user123", result.ID)
	assert.Equal(t, "john.doe", result.Name)
	assert.Equal(t, "john@example.com", result.Email)
	assert.Equal(t, []string{"admin", "users"}, result.Groups)
}

func TestGetUser_NotFound(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
		w.Write([]byte("user not found"))
	}))
	defer server.Close()

	client := NewClient(server.URL, "token", "secret", false)
	_, err := client.GetUser("org123", "nonexistent")

	require.Error(t, err)
	assert.Contains(t, err.Error(), "404")
}

func TestGetUser_InvalidJSON(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("invalid json"))
	}))
	defer server.Close()

	client := NewClient(server.URL, "token", "secret", false)
	_, err := client.GetUser("org123", "user123")

	require.Error(t, err)
	assert.Contains(t, err.Error(), "unmarshal")
}

func TestGetUserByName_Found(t *testing.T) {
	t.Parallel()

	users := []User{
		{ID: "user1", Name: "john.doe"},
		{ID: "user2", Name: "jane.doe"},
		{ID: "user3", Name: "bob.smith"},
	}

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/user/org123", r.URL.Path)
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(users)
	}))
	defer server.Close()

	client := NewClient(server.URL, "token", "secret", false)
	result, err := client.GetUserByName("org123", "jane.doe")

	require.NoError(t, err)
	assert.Equal(t, "user2", result.ID)
	assert.Equal(t, "jane.doe", result.Name)
}

func TestGetUserByName_NotFound(t *testing.T) {
	t.Parallel()

	users := []User{
		{ID: "user1", Name: "john.doe"},
		{ID: "user2", Name: "jane.doe"},
	}

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(users)
	}))
	defer server.Close()

	client := NewClient(server.URL, "token", "secret", false)
	_, err := client.GetUserByName("org123", "nonexistent")

	require.Error(t, err)
	assert.Contains(t, err.Error(), "not found")
	assert.Contains(t, err.Error(), "nonexistent")
	assert.Contains(t, err.Error(), "org123")
}

func TestGetUserByName_ListError(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("server error"))
	}))
	defer server.Close()

	client := NewClient(server.URL, "token", "secret", false)
	_, err := client.GetUserByName("org123", "john.doe")

	require.Error(t, err)
}

func TestCreateUser_Success(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/user/org123", r.URL.Path)
		assert.Equal(t, http.MethodPost, r.Method)

		var input User
		err := json.NewDecoder(r.Body).Decode(&input)
		require.NoError(t, err)
		assert.Equal(t, "newuser", input.Name)
		assert.Equal(t, "new@example.com", input.Email)

		// API returns array with single user
		created := []User{{
			ID:    "new-user-id",
			Name:  input.Name,
			Email: input.Email,
		}}
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(created)
	}))
	defer server.Close()

	client := NewClient(server.URL, "token", "secret", false)
	user := &User{
		Name:  "newuser",
		Email: "new@example.com",
	}
	result, err := client.CreateUser("org123", user)

	require.NoError(t, err)
	assert.Equal(t, "new-user-id", result.ID)
	assert.Equal(t, "newuser", result.Name)
	assert.Equal(t, "new@example.com", result.Email)
}

func TestCreateUser_EmptyArrayResponse(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Return empty array
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("[]"))
	}))
	defer server.Close()

	client := NewClient(server.URL, "token", "secret", false)
	user := &User{Name: "newuser"}
	_, err := client.CreateUser("org123", user)

	require.Error(t, err)
	assert.Contains(t, err.Error(), "no user returned")
}

func TestCreateUser_Error(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte("invalid user"))
	}))
	defer server.Close()

	client := NewClient(server.URL, "token", "secret", false)
	user := &User{Name: ""}
	_, err := client.CreateUser("org123", user)

	require.Error(t, err)
}

func TestCreateUser_InvalidResponseJSON(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("not json"))
	}))
	defer server.Close()

	client := NewClient(server.URL, "token", "secret", false)
	user := &User{Name: "newuser"}
	_, err := client.CreateUser("org123", user)

	require.Error(t, err)
	assert.Contains(t, err.Error(), "unmarshal")
}

func TestUpdateUser_Success(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/user/org123/user456", r.URL.Path)
		assert.Equal(t, http.MethodPut, r.Method)

		var input User
		err := json.NewDecoder(r.Body).Decode(&input)
		require.NoError(t, err)
		assert.Equal(t, "updated@example.com", input.Email)

		updated := User{
			ID:    "user456",
			Name:  input.Name,
			Email: input.Email,
		}
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(updated)
	}))
	defer server.Close()

	client := NewClient(server.URL, "token", "secret", false)
	user := &User{
		Name:  "john.doe",
		Email: "updated@example.com",
	}
	result, err := client.UpdateUser("org123", "user456", user)

	require.NoError(t, err)
	assert.Equal(t, "user456", result.ID)
	assert.Equal(t, "updated@example.com", result.Email)
}

func TestUpdateUser_NotFound(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
		w.Write([]byte("user not found"))
	}))
	defer server.Close()

	client := NewClient(server.URL, "token", "secret", false)
	user := &User{Name: "john.doe"}
	_, err := client.UpdateUser("org123", "nonexistent", user)

	require.Error(t, err)
	assert.Contains(t, err.Error(), "404")
}

func TestUpdateUser_InvalidResponseJSON(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("not json"))
	}))
	defer server.Close()

	client := NewClient(server.URL, "token", "secret", false)
	user := &User{Name: "john.doe"}
	_, err := client.UpdateUser("org123", "user456", user)

	require.Error(t, err)
	assert.Contains(t, err.Error(), "unmarshal")
}

func TestDeleteUser_Success(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/user/org123/user456", r.URL.Path)
		assert.Equal(t, http.MethodDelete, r.Method)

		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	client := NewClient(server.URL, "token", "secret", false)
	err := client.DeleteUser("org123", "user456")

	require.NoError(t, err)
}

func TestDeleteUser_NotFound(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
		w.Write([]byte("user not found"))
	}))
	defer server.Close()

	client := NewClient(server.URL, "token", "secret", false)
	err := client.DeleteUser("org123", "nonexistent")

	require.Error(t, err)
	assert.Contains(t, err.Error(), "404")
}

func TestDeleteUser_ServerError(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("internal error"))
	}))
	defer server.Close()

	client := NewClient(server.URL, "token", "secret", false)
	err := client.DeleteUser("org123", "user456")

	require.Error(t, err)
}

func TestGetUserKey_Success(t *testing.T) {
	t.Parallel()

	keyData := `[Interface]
PrivateKey = abc123
Address = 10.10.0.2/24

[Peer]
PublicKey = xyz789
AllowedIPs = 0.0.0.0/0
Endpoint = vpn.example.com:1194`

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/key/org123/user456", r.URL.Path)
		assert.Equal(t, http.MethodGet, r.Method)

		w.WriteHeader(http.StatusOK)
		w.Write([]byte(keyData))
	}))
	defer server.Close()

	client := NewClient(server.URL, "token", "secret", false)
	result, err := client.GetUserKey("org123", "user456")

	require.NoError(t, err)
	assert.Equal(t, keyData, string(result))
}

func TestGetUserKey_NotFound(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
		w.Write([]byte("user not found"))
	}))
	defer server.Close()

	client := NewClient(server.URL, "token", "secret", false)
	_, err := client.GetUserKey("org123", "nonexistent")

	require.Error(t, err)
	assert.Contains(t, err.Error(), "404")
}

func TestUser_AllFields(t *testing.T) {
	t.Parallel()

	user := User{
		ID:               "user123",
		Name:             "john.doe",
		Type:             "client",
		AuthType:         "local",
		DNSServers:       []string{"8.8.8.8", "8.8.4.4"},
		DNSSuffix:        "example.com",
		DnsMappingKind:   "internal",
		Disabled:         false,
		NetworkLinks:     []string{"link1", "link2"},
		PortForwarding:   []string{"8080:80"},
		Email:            "john@example.com",
		Status:           true,
		OtpSecret:        "secret123",
		ClientToClient:   true,
		MacAddresses:     []string{"00:11:22:33:44:55"},
		YubicoID:         "yubi123",
		SSO:              "google",
		BypassSecondary:  false,
		Groups:           []string{"admin", "users"},
		Audit:            true,
		Gravatar:         true,
		OtpAuth:          false,
		DeviceAuth:       true,
		Organization:     "org123",
		OrganizationName: "Engineering",
		Pin:              true,
		Servers:          []string{"srv1", "srv2"},
	}

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(user)
	}))
	defer server.Close()

	client := NewClient(server.URL, "token", "secret", false)
	result, err := client.GetUser("org123", "user123")

	require.NoError(t, err)
	assert.Equal(t, user.ID, result.ID)
	assert.Equal(t, user.Name, result.Name)
	assert.Equal(t, user.Type, result.Type)
	assert.Equal(t, user.AuthType, result.AuthType)
	assert.Equal(t, user.DNSServers, result.DNSServers)
	assert.Equal(t, user.DNSSuffix, result.DNSSuffix)
	assert.Equal(t, user.Disabled, result.Disabled)
	assert.Equal(t, user.NetworkLinks, result.NetworkLinks)
	assert.Equal(t, user.Email, result.Email)
	assert.Equal(t, user.Groups, result.Groups)
	assert.Equal(t, user.Organization, result.Organization)
	assert.Equal(t, user.OrganizationName, result.OrganizationName)
	assert.Equal(t, user.Servers, result.Servers)
}

func TestCreateUser_WithGroups(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var input User
		err := json.NewDecoder(r.Body).Decode(&input)
		require.NoError(t, err)
		assert.Equal(t, []string{"admin", "developers"}, input.Groups)

		created := []User{{
			ID:     "new-user-id",
			Name:   input.Name,
			Groups: input.Groups,
		}}
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(created)
	}))
	defer server.Close()

	client := NewClient(server.URL, "token", "secret", false)
	user := &User{
		Name:   "newuser",
		Groups: []string{"admin", "developers"},
	}
	result, err := client.CreateUser("org123", user)

	require.NoError(t, err)
	assert.Equal(t, []string{"admin", "developers"}, result.Groups)
}

func TestUpdateUser_DisableUser(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var input User
		err := json.NewDecoder(r.Body).Decode(&input)
		require.NoError(t, err)
		assert.True(t, input.Disabled)

		updated := User{
			ID:       "user456",
			Name:     input.Name,
			Disabled: input.Disabled,
		}
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(updated)
	}))
	defer server.Close()

	client := NewClient(server.URL, "token", "secret", false)
	user := &User{
		Name:     "john.doe",
		Disabled: true,
	}
	result, err := client.UpdateUser("org123", "user456", user)

	require.NoError(t, err)
	assert.True(t, result.Disabled)
}
