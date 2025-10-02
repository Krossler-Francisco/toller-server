package tests

import (
	"bytes"
	"encoding/json"
	"net/http"
	"strconv"
	"testing"

	"toller-server/modules/friends"

	"github.com/stretchr/testify/assert"
)

func TestFriendsFlow(t *testing.T) {
	server, _ := setupTestServer(t)
	client := &http.Client{}

	// --- 1. Create users ---
	userA_ID, userA_Token := registerAndLogin(t, server.URL, "userA_friends", "usera_friends@test.com", "password123")
	userB_ID, userB_Token := registerAndLogin(t, server.URL, "userB_friends", "userb_friends@test.com", "password123")

	// --- 2. User A sends a friend request to User B ---
	friendReqData := map[string]int{"friend_id": userB_ID}
	friendReqBody, _ := json.Marshal(friendReqData)
	req, _ := http.NewRequest("POST", server.URL+"/api/v1/friends/requests", bytes.NewBuffer(friendReqBody))
	req.Header.Set("Authorization", "Bearer "+userA_Token)
	req.Header.Set("Content-Type", "application/json")
	resp, err := client.Do(req)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusCreated, resp.StatusCode)
	resp.Body.Close()

	// --- 3. User B lists pending friend requests ---
	req, _ = http.NewRequest("GET", server.URL+"/api/v1/friends/requests/pending", nil)
	req.Header.Set("Authorization", "Bearer "+userB_Token)
	resp, err = client.Do(req)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
	var pendingRequests []friends.Friend
	json.NewDecoder(resp.Body).Decode(&pendingRequests)
	assert.Len(t, pendingRequests, 1, "User B should have 1 pending request")
	assert.Equal(t, userA_ID, pendingRequests[0].UserID)
	resp.Body.Close()

	// --- 4. User B accepts the friend request from User A ---
	updateReqData := map[string]string{"status": "accepted"}
	updateReqBody, _ := json.Marshal(updateReqData)
	req, _ = http.NewRequest("PUT", server.URL+"/api/v1/friends/requests/"+strconv.Itoa(userA_ID), bytes.NewBuffer(updateReqBody))
	req.Header.Set("Authorization", "Bearer "+userB_Token)
	req.Header.Set("Content-Type", "application/json")
	resp, err = client.Do(req)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusNoContent, resp.StatusCode)
	resp.Body.Close()

	// --- 5. User A lists friends ---
	req, _ = http.NewRequest("GET", server.URL+"/api/v1/friends", nil)
	req.Header.Set("Authorization", "Bearer "+userA_Token)
	resp, err = client.Do(req)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
	var userAFriends []friends.Friend
	json.NewDecoder(resp.Body).Decode(&userAFriends)
	assert.Len(t, userAFriends, 1, "User A should have 1 friend")
	assert.Equal(t, userB_ID, userAFriends[0].FriendID)
	resp.Body.Close()

	// --- 6. User B lists friends ---
	req, _ = http.NewRequest("GET", server.URL+"/api/v1/friends", nil)
	req.Header.Set("Authorization", "Bearer "+userB_Token)
	resp, err = client.Do(req)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
	var userBFriends []friends.Friend
	json.NewDecoder(resp.Body).Decode(&userBFriends)
	assert.Len(t, userBFriends, 1, "User B should have 1 friend")
	assert.Equal(t, userA_ID, userBFriends[0].UserID)
	resp.Body.Close()
}
