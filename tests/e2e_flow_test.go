package tests

import (
	"bytes"
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/gorilla/mux"
	"github.com/gorilla/websocket"
	"github.com/joho/godotenv"
	_ "github.com/lib/pq"
	"github.com/stretchr/testify/assert"

	"toller-server/modules/auth"
	"toller-server/modules/channels"
	"toller-server/modules/chat"
	"toller-server/modules/dms"
	"toller-server/modules/teams"
)

// setupTestServer inicializa un servidor de prueba con una base de datos limpia.
func setupTestServer(t *testing.T) (*httptest.Server, *sql.DB) {
	if err := godotenv.Load("../.env"); err != nil {
		log.Println("No se encontró archivo .env, usando variables de entorno del sistema")
	}

	dbURL := os.Getenv("DB_URL")
	if dbURL == "" {
		log.Fatal("DB_URL no está configurada para el test.")
	}

	jwtSecret := os.Getenv("JWT_SECRET")
	if jwtSecret == "" {
		log.Fatal("JWT_SECRET no está configurada para el test.")
	}

	db, err := sql.Open("postgres", dbURL)
	if err != nil {
		log.Fatal("Error al conectar a la DB:", err)
	}

	// Limpiar tablas para un estado inicial consistente
	cleanupTables(t, db)

	authRepo := &auth.UserRepository{DB: db}
	authService := &auth.AuthService{Repo: authRepo}
	authHandler := &auth.AuthHandler{Service: authService}

	teamsRepo := &teams.TeamRepository{DB: db}
	teamsService := &teams.TeamService{Repo: teamsRepo}
	teamsHandler := &teams.TeamHandler{Service: teamsService}

	channelsRepo := &channels.ChannelRepository{DB: db}
	channelsService := &channels.ChannelService{Repo: channelsRepo}
	channelsHandler := &channels.ChannelHandler{Service: channelsService}

	// Inicializar DMs
	
	
	

	hub := chat.NewHub()
	chatHandler := chat.NewHandler(db, jwtSecret, hub)

	r := mux.NewRouter()

	// Rutas
	r.HandleFunc("/ws/channel/{channel_id}", chatHandler.ServeWS)
	auth.RegisterRoutes(r, authHandler)
	teams.RegisterRoutes(r, teamsHandler, auth.JWTMiddleware)
	channels.RegisterRoutes(r, channelsHandler, auth.JWTMiddleware)
	dms.RegisterDMSRoutes(r, db)

	server := httptest.NewServer(r)

	// Función de limpieza para cerrar la DB y el servidor
	t.Cleanup(func() {
		server.Close()
		db.Close()
	})

	return server, db
}

func cleanupTables(t *testing.T, db *sql.DB) {
	// El orden es importante por las foreign keys
	_, err := db.Exec(`
        DELETE FROM messages;
        DELETE FROM channel_users;
        DELETE FROM channels;
        DELETE FROM user_teams;
        DELETE FROM teams;
        DELETE FROM users;
    `)
	if err != nil {
		t.Fatalf("Failed to clean up tables: %v", err)
	}
}

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

	resp, err := http.Post(server.URL+"/register", "application/json", bytes.NewBuffer(registerBody))
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

	resp, err = http.Post(server.URL+"/login", "application/json", bytes.NewBuffer(loginBody))
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
