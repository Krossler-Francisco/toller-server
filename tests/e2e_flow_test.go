package tests

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"testing"
	"time"
	"toller-server/modules/channels"
	"toller-server/modules/chat"

	"github.com/gorilla/websocket"
	"github.com/stretchr/testify/assert"
)

func TestE2EFullFlow(t *testing.T) {
	server, db := setupTestServer(t)

	// --- 1. Registrar un nuevo usuario ---
	uniqueEmail := fmt.Sprintf("testuser_%d@example.com", time.Now().UnixNano())
	registerData := map[string]string{
		"username": "testuser",
		"email":    uniqueEmail,
		"password": "password123",
	}
	registerBody, _ := json.Marshal(registerData)

	resp, err := http.Post(server.URL+"/api/v1/auth/register", "application/json", bytes.NewBuffer(registerBody))
	assert.NoError(t, err)
	assert.Equal(t, http.StatusCreated, resp.StatusCode)

	var registerResp map[string]interface{}
	json.NewDecoder(resp.Body).Decode(&registerResp)
	userID := int((registerResp["id"]).(float64))
	assert.NotZero(t, userID)
	resp.Body.Close()

	// --- 2. Iniciar sesión ---
	loginData := map[string]string{
		"email":    uniqueEmail,
		"password": "password123",
	}
	loginBody, _ := json.Marshal(loginData)

	resp, err = http.Post(server.URL+"/api/v1/auth/login", "application/json", bytes.NewBuffer(loginBody))
	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)

	var loginResp map[string]string
	json.NewDecoder(resp.Body).Decode(&loginResp)
	token := loginResp["token"]
	assert.NotEmpty(t, token)
	resp.Body.Close()

	// --- 3. Crear un equipo ---
	teamData := map[string]string{
		"name":        "Mi Equipo de Prueba",
		"description": "Un equipo para el test E2E",
	}
	teamBody, _ := json.Marshal(teamData)

	req, _ := http.NewRequest("POST", server.URL+"/teams", bytes.NewBuffer(teamBody))
	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	resp, err = client.Do(req)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusCreated, resp.StatusCode)

	var teamResp map[string]map[string]interface{}
	json.NewDecoder(resp.Body).Decode(&teamResp)
	teamID := int(teamResp["team"]["id"].(float64))
	assert.NotZero(t, teamID)
	resp.Body.Close()

	// --- 4. Crear un canal ---
	channelData := map[string]string{"name": "Mi Canal de Prueba"}
	channelBody, _ := json.Marshal(channelData)

	channelURL := fmt.Sprintf(server.URL+"/api/v1/teams/%d/channels", teamID)
	req, _ = http.NewRequest("POST", channelURL, bytes.NewBuffer(channelBody))
	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Content-Type", "application/json")

	resp, err = client.Do(req)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusCreated, resp.StatusCode)

	var channelResp channels.Channel
	json.NewDecoder(resp.Body).Decode(&channelResp)
	channelID := channelResp.ID
	assert.NotZero(t, channelID)
	resp.Body.Close()

	// --- 5. Enviar un mensaje vía WebSocket ---
	wsURL := "ws" + strings.TrimPrefix(server.URL, "http")
	wsConnURL := fmt.Sprintf("%s/ws/channel/%d?token=%s", wsURL, channelID, token)

	ws, _, err := websocket.DefaultDialer.Dial(wsConnURL, nil)
	assert.NoError(t, err)
	defer ws.Close()

	// Enviar mensaje
	messageContent := "Hola, este es un mensaje de prueba E2E!"
	msgToSend := chat.IncomingMessage{Type: "message", Content: messageContent}
	err = ws.WriteJSON(msgToSend)
	assert.NoError(t, err)

	// Dar un pequeño margen para que el servidor procese y guarde el mensaje
	time.Sleep(200 * time.Millisecond)

	// --- 6. Verificar que el mensaje fue guardado en la DB ---
	var savedContent string
	query := "SELECT content FROM messages WHERE channel_id = $1 AND user_id = $2"
	err = db.QueryRow(query, channelID, userID).Scan(&savedContent)
	assert.NoError(t, err, "El mensaje no fue encontrado en la base de datos")
	assert.Equal(t, messageContent, savedContent)
}
