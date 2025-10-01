package tests

import (
	"encoding/json"
	"net/http"
	"strconv"
	"testing"

	"github.com/stretchr/testify/assert"
	"toller-server/modules/users"
)

func TestUsersFlow(t *testing.T) {
	server, _ := setupTestServer(t)
	client := &http.Client{}

	// --- 1. Create users ---
	userA_ID, userA_Token := registerAndLogin(t, server.URL, "userA_users", "usera_users@test.com", "password123")
	_, _ = registerAndLogin(t, server.URL, "userB_users", "userb_users@test.com", "password123")

	// --- 2. Get all users ---
	req, _ := http.NewRequest("GET", server.URL+"/api/v1/users", nil)
	req.Header.Set("Authorization", "Bearer "+userA_Token)
	resp, err := client.Do(req)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
	var allUsers []users.User
	json.NewDecoder(resp.Body).Decode(&allUsers)
	assert.GreaterOrEqual(t, len(allUsers), 2, "Should have at least 2 users")
	resp.Body.Close()

	// --- 3. Get user by ID ---
	req, _ = http.NewRequest("GET", server.URL+"/api/v1/users/"+strconv.Itoa(userA_ID), nil)
	req.Header.Set("Authorization", "Bearer "+userA_Token)
	resp, err = client.Do(req)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
	var userA users.User
	json.NewDecoder(resp.Body).Decode(&userA)
	assert.Equal(t, userA_ID, userA.ID)
	assert.Equal(t, "userA_users", userA.Username)
	resp.Body.Close()

	// --- 4. Search users ---
	req, _ = http.NewRequest("GET", server.URL+"/api/v1/users/search?q=userB", nil)
	req.Header.Set("Authorization", "Bearer "+userA_Token)
	resp, err = client.Do(req)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
	var searchResult []users.User
	json.NewDecoder(resp.Body).Decode(&searchResult)
	assert.Len(t, searchResult, 1, "Should find 1 user")
	assert.Equal(t, "userB_users", searchResult[0].Username)
	resp.Body.Close()
}
