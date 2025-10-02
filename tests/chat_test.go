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

	"github.com/gorilla/websocket"
	"github.com/stretchr/testify/assert"
)

// TestWebSocketBroadcast valida el flujo completo de broadcast entre dos clientes.
// 1. Crea los usuarios, equipo y canal necesarios.
// 2. Conecta ambos clientes al WebSocket.
// 3. Un cliente envía un mensaje.
// 4. Verifica que el otro cliente lo recibe.
func TestWebSocketBroadcast(t *testing.T) {
	// --- 1. SETUP: Servidor y usuarios ---
	server, _ := setupTestServer(t)
	client := &http.Client{}
	wsURL := "ws" + strings.TrimPrefix(server.URL, "http")

	// Crear Usuario 1 (Fran, el remitente)
	franEmail := fmt.Sprintf("fran_%d@test.com", time.Now().UnixNano())
	franID, franToken := registerAndLogin(t, server.URL, "fran", franEmail, "password123")

	// Crear Usuario 2 (Maria, la receptora)
	mariaEmail := fmt.Sprintf("maria_%d@test.com", time.Now().UnixNano())
	mariaID, mariaToken := registerAndLogin(t, server.URL, "maria", mariaEmail, "password123")

	// --- 2. SETUP: Equipo y Canal ---

	// Fran crea un equipo
	teamData := map[string]string{"name": "Equipo de Broadcast"}
	teamBody, _ := json.Marshal(teamData)
	req, _ := http.NewRequest("POST", server.URL+"/teams", bytes.NewBuffer(teamBody))
	req.Header.Set("Authorization", "Bearer "+franToken)
	req.Header.Set("Content-Type", "application/json")
	resp, err := client.Do(req)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusCreated, resp.StatusCode)
	var teamResp map[string]map[string]interface{}
	json.NewDecoder(resp.Body).Decode(&teamResp)
	teamID := int(teamResp["team"]["id"].(float64))
	resp.Body.Close()

	// Fran crea un canal
	channelData := map[string]string{"name": "Canal de Broadcast"}
	channelBody, _ := json.Marshal(channelData)
	channelURL := fmt.Sprintf(server.URL+"/api/v1/teams/%d/channels", teamID)
	req, _ = http.NewRequest("POST", channelURL, bytes.NewBuffer(channelBody))
	req.Header.Set("Authorization", "Bearer "+franToken)
	req.Header.Set("Content-Type", "application/json")
	resp, err = client.Do(req)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusCreated, resp.StatusCode)
	var channelResp map[string]interface{}
	json.NewDecoder(resp.Body).Decode(&channelResp)
	channelID := int(channelResp["id"].(float64))
	resp.Body.Close()

	// Fran añade a Maria al equipo y al canal
	addMemberToTeamURL := fmt.Sprintf("%s/teams/%d/members", server.URL, teamID)
	addMemberData, _ := json.Marshal(map[string]int{"user_id": mariaID})
	req, _ = http.NewRequest("POST", addMemberToTeamURL, bytes.NewBuffer(addMemberData))
	req.Header.Set("Authorization", "Bearer "+franToken)
	req.Header.Set("Content-Type", "application/json")
	resp, err = client.Do(req)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
	resp.Body.Close()

	addMemberToChannelURL := fmt.Sprintf(server.URL+"/api/v1/channels/%d/members", channelID)
	addMemberToChannelData, _ := json.Marshal(map[string]interface{}{"user_id": mariaID, "role": "user"})
	req, _ = http.NewRequest("POST", addMemberToChannelURL, bytes.NewBuffer(addMemberToChannelData))
	req.Header.Set("Authorization", "Bearer "+franToken)
	req.Header.Set("Content-Type", "application/json")
	resp, err = client.Do(req)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
	resp.Body.Close()

	// --- 3. Lógica del Test de WebSocket ---

	// Conectar a Maria (receptora)
	mariaConnURL := fmt.Sprintf("%s/ws/channel/%d?token=%s", wsURL, channelID, mariaToken)
	mariaConn, _, err := websocket.DefaultDialer.Dial(mariaConnURL, nil)
	assert.NoError(t, err)
	defer mariaConn.Close()

	// Goroutine para escuchar los mensajes de Maria
	msgChan := make(chan []byte, 1)
	go func() {
		defer close(msgChan)
		// El historial de mensajes está vacío, así que el primer mensaje debe ser el de broadcast
		_, msg, err := mariaConn.ReadMessage()
		if err != nil {
			return // El test principal detectará el timeout
		}
		msgChan <- msg
	}()

	// Conectar a Fran (remitente)
	franConnURL := fmt.Sprintf("%s/ws/channel/%d?token=%s", wsURL, channelID, franToken)
	franConn, _, err := websocket.DefaultDialer.Dial(franConnURL, nil)
	assert.NoError(t, err)
	defer franConn.Close()

	// Pequeña pausa para asegurar que ambos clientes están registrados en el Hub
	time.Sleep(200 * time.Millisecond)

	// Fran envía el mensaje
	messageContent := "Hola Maria, esto es un test auto-contenido!"
	err = franConn.WriteJSON(chat.IncomingMessage{Type: "message", Content: messageContent})
	assert.NoError(t, err)

	// --- 4. VERIFICACIÓN ---
	select {
	case receivedMsgBytes := <-msgChan:
		var receivedMsg chat.OutgoingMessage
		err := json.Unmarshal(receivedMsgBytes, &receivedMsg)
		assert.NoError(t, err, "El mensaje recibido debería ser un JSON válido")

		// Verificar contenido y remitente
		assert.Equal(t, messageContent, receivedMsg.Content)
		assert.Equal(t, franID, int(receivedMsg.UserID))

	case <-time.After(2 * time.Second):
		t.Fatal("Timeout: Maria no recibió el mensaje de broadcast a tiempo.")
	}
}
