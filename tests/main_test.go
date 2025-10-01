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
	"testing"

	"github.com/gorilla/mux"
	"github.com/joho/godotenv"
	"github.com/stretchr/testify/assert"
	"toller-server/modules/auth"
	"toller-server/modules/channels"
	"toller-server/modules/chat"
	"toller-server/modules/dms"
	"toller-server/modules/friends"
	"toller-server/modules/teams"
	"toller-server/modules/users"
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

	r := mux.NewRouter()

	// Rutas
	auth.RegisterRoutes(r, &auth.AuthHandler{Service: &auth.AuthService{Repo: &auth.UserRepository{DB: db}}})
	teams.RegisterRoutes(r, &teams.TeamHandler{Service: &teams.TeamService{Repo: &teams.TeamRepository{DB: db}}}, auth.JWTMiddleware)
	channels.RegisterRoutes(r, &channels.ChannelHandler{Service: &channels.ChannelService{Repo: &channels.ChannelRepository{DB: db}}}, auth.JWTMiddleware)
	dms.RegisterDMSRoutes(r, db)
	users.RegisterUserRoutes(r, db)
	friends.RegisterFriendRoutes(r, db)

	hub := chat.NewHub()
	chatHandler := chat.NewHandler(db, jwtSecret, hub)
	r.HandleFunc("/ws/channel/{channel_id}", chatHandler.ServeWS)

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
		DELETE FROM friends;
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

func registerAndLogin(t *testing.T, serverURL, username, email, password string) (int, string) {
	// Register
	registerData := map[string]string{"username": username, "email": email, "password": password}
	registerBody, _ := json.Marshal(registerData)
	resp, err := http.Post(serverURL+"/register", "application/json", bytes.NewBuffer(registerBody))
	assert.NoError(t, err)
	assert.Equal(t, http.StatusCreated, resp.StatusCode)

	var registerResp map[string]interface{}
	json.NewDecoder(resp.Body).Decode(&registerResp)
	userID := int(registerResp["id"].(float64))
	resp.Body.Close()

	// Login
	loginData := map[string]string{"email": email, "password": password}
	loginBody, _ := json.Marshal(loginData)
	resp, err = http.Post(serverURL+"/login", "application/json", bytes.NewBuffer(loginBody))
	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)

	var loginResp map[string]string
	json.NewDecoder(resp.Body).Decode(&loginResp)
	token := loginResp["token"]
	resp.Body.Close()

	return userID, token
}
