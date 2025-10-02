package tests

import (
	"bytes"
	"database/sql"
	"encoding/json"
	"log"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	_ "github.com/lib/pq"

	"toller-server/modules/auth"
	"toller-server/modules/channels"
	"toller-server/modules/chat"
	"toller-server/modules/dms"
	"toller-server/modules/friends"
	"toller-server/modules/teams"
	"toller-server/modules/users"

	"github.com/gorilla/mux"

	"github.com/joho/godotenv"
)

// setupTestServer inicializa el servidor y la base de datos para los tests
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

	cleanupTables(t, db)

	// Inicializar handlers
	authRepo := &auth.UserRepository{DB: db}
	authService := &auth.AuthService{Repo: authRepo}
	authHandler := &auth.AuthHandler{Service: authService}

	teamsRepo := &teams.TeamRepository{DB: db}
	teamsService := &teams.TeamService{Repo: teamsRepo}
	teamsHandler := &teams.TeamHandler{Service: teamsService}

	channelsRepo := &channels.ChannelRepository{DB: db}
	channelsService := &channels.ChannelService{Repo: channelsRepo}
	channelsHandler := &channels.ChannelHandler{Service: channelsService}

	hub := chat.NewHub()
	chatHandler := chat.NewHandler(db, jwtSecret, hub)

	r := mux.NewRouter()
	r.HandleFunc("/ws/channel/{channel_id}", chatHandler.ServeWS)
	auth.RegisterRoutes(r, authHandler)
	teams.RegisterRoutes(r, teamsHandler, auth.JWTMiddleware)
	channels.RegisterRoutes(r, channelsHandler, auth.JWTMiddleware)
	dms.RegisterDMSRoutes(r, db)
	users.RegisterUserRoutes(r, db)
	friends.RegisterFriendRoutes(r, db)

	server := httptest.NewServer(r)

	t.Cleanup(func() {
		server.Close()
		db.Close()
	})

	return server, db
}

// cleanupTables limpia las tablas relevantes para los tests
func cleanupTables(t *testing.T, db *sql.DB) {
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

// registerAndLogin registra y loguea un usuario, devolviendo su ID y token
func registerAndLogin(t *testing.T, serverURL, username, email, password string) (int, string) {
	client := &http.Client{}
	registerData := map[string]string{"username": username, "email": email, "password": password}
	registerBody, _ := json.Marshal(registerData)
	resp, err := client.Post(serverURL+"/api/v1/auth/register", "application/json", bytes.NewBuffer(registerBody))
	if err != nil {
		t.Fatalf("Error en registro: %v", err)
	}
	if resp.StatusCode != http.StatusCreated {
		t.Fatalf("Registro falló: %v", resp.Status)
	}
	var registerResp map[string]interface{}
	json.NewDecoder(resp.Body).Decode(&registerResp)
	userID := int(registerResp["id"].(float64))
	resp.Body.Close()

	loginData := map[string]string{"email": email, "password": password}
	loginBody, _ := json.Marshal(loginData)
	resp, err = client.Post(serverURL+"/api/v1/auth/login", "application/json", bytes.NewBuffer(loginBody))
	if err != nil {
		t.Fatalf("Error en login: %v", err)
	}
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("Login falló: %v", resp.Status)
	}
	var loginResp map[string]interface{}
	json.NewDecoder(resp.Body).Decode(&loginResp)
	token := loginResp["token"].(string)
	resp.Body.Close()

	return userID, token
}
