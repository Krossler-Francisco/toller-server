package tests

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"testing"
	"time"

	"toller-server/modules/chat"
	"toller-server/modules/dms"

	"github.com/gorilla/websocket"
	"github.com/stretchr/testify/assert"
)

func TestDirectMessageFlow(t *testing.T) {
	// --- 1. SETUP ---
	server, _ := setupTestServer(t)
	client := &http.Client{}
	wsURL := "ws" + strings.TrimPrefix(server.URL, "http")

	// --- 2. Crear usuarios ---
	userA_ID, userA_Token := registerAndLogin(t, server.URL, "userA_dm", "usera_dm@test.com", "password123")
	userB_ID, userB_Token := registerAndLogin(t, server.URL, "userB_dm", "userb_dm@test.com", "password123")
	userC_ID, _ := registerAndLogin(t, server.URL, "userC_dm", "userc_dm@test.com", "password123")

	// --- 3. Iniciar DMs ---
	// User A inicia un DM con User B
	startDMDataB := map[string]int{"recipient_id": userB_ID}
	startDMBodyB, _ := json.Marshal(startDMDataB)
	req, _ := http.NewRequest("POST", server.URL+"/api/v1/dms", bytes.NewBuffer(startDMBodyB))
	req.Header.Set("Authorization", "Bearer "+userA_Token)
	req.Header.Set("Content-Type", "application/json")
	resp, err := client.Do(req)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
	var dmChannelRespB map[string]int
	json.NewDecoder(resp.Body).Decode(&dmChannelRespB)
	channelID_AB := dmChannelRespB["channel_id"]
	resp.Body.Close()

	// User A inicia un DM con User C
	startDMDataC := map[string]int{"recipient_id": userC_ID}
	startDMBodyC, _ := json.Marshal(startDMDataC)
	req, _ = http.NewRequest("POST", server.URL+"/api/v1/dms", bytes.NewBuffer(startDMBodyC))
	req.Header.Set("Authorization", "Bearer "+userA_Token)
	req.Header.Set("Content-Type", "application/json")
	resp, err = client.Do(req)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
	resp.Body.Close()

	// --- 4. Listar DMs para Usuario A ---
	req, _ = http.NewRequest("GET", server.URL+"/api/v1/dms", nil)
	req.Header.Set("Authorization", "Bearer "+userA_Token)
	resp, err = client.Do(req)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
	var dmsForA []dms.DMChannelInfo
	json.NewDecoder(resp.Body).Decode(&dmsForA)
	assert.Len(t, dmsForA, 2, "Usuario A debería tener 2 DMs")
	// Verificar que los DMs son con B y C
	otherUserIDs := []int{dmsForA[0].OtherUserID, dmsForA[1].OtherUserID}
	assert.Contains(t, otherUserIDs, userB_ID)
	assert.Contains(t, otherUserIDs, userC_ID)
	resp.Body.Close()

	// --- 5. Listar DMs para Usuario B ---
	req, _ = http.NewRequest("GET", server.URL+"/api/v1/dms", nil)
	req.Header.Set("Authorization", "Bearer "+userB_Token)
	resp, err = client.Do(req)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
	var dmsForB []dms.DMChannelInfo
	json.NewDecoder(resp.Body).Decode(&dmsForB)
	assert.Len(t, dmsForB, 1, "Usuario B debería tener 1 DM")
	assert.Equal(t, userA_ID, dmsForB[0].OtherUserID, "El DM del Usuario B debería ser con el Usuario A")
	resp.Body.Close()

	// --- 6. Test de comunicación por WebSocket en el canal de DM (A -> B) ---
	userB_ConnURL := fmt.Sprintf("%s/ws/channel/%d?token=%s", wsURL, channelID_AB, userB_Token)
	userB_Conn, _, err := websocket.DefaultDialer.Dial(userB_ConnURL, nil)
	assert.NoError(t, err)
	defer userB_Conn.Close()

	msgChan := make(chan []byte, 1)
	go func() {
		defer close(msgChan)
		_, msg, err := userB_Conn.ReadMessage()
		if err != nil {
			return
		}
		msgChan <- msg
	}()

	userA_ConnURL := fmt.Sprintf("%s/ws/channel/%d?token=%s", wsURL, channelID_AB, userA_Token)
	userA_Conn, _, err := websocket.DefaultDialer.Dial(userA_ConnURL, nil)
	assert.NoError(t, err)
	defer userA_Conn.Close()

	time.Sleep(200 * time.Millisecond)

	message := "Hola, este es nuestro primer DM!"
	err = userA_Conn.WriteJSON(chat.IncomingMessage{Type: "message", Content: message})
	assert.NoError(t, err)

	select {
	case receivedMsgBytes := <-msgChan:
		var receivedMsg chat.OutgoingMessage
		err := json.Unmarshal(receivedMsgBytes, &receivedMsg)
		assert.NoError(t, err)
		assert.Equal(t, message, receivedMsg.Content)
		assert.Equal(t, userA_ID, int(receivedMsg.UserID))
	case <-time.After(2 * time.Second):
		t.Fatal("Timeout: El Usuario B no recibió el mensaje del DM.")
	}
}
