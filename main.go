package main

import (
	"database/sql"
	"log"
	"net/http"
	"os"
	"path/filepath"

	"github.com/gorilla/mux"
	"github.com/joho/godotenv"
	_ "github.com/lib/pq"

	"toller-server/modules/auth"
	"toller-server/modules/channels"
	"toller-server/modules/chat"
	"toller-server/modules/dms"
	"toller-server/modules/teams"
	"toller-server/modules/users"
)

func main() {
	// Cargar .env
	envPath := filepath.Join("..", "..", ".env")
	if err := godotenv.Load(envPath); err != nil {
		if err := godotenv.Load(".env"); err != nil {
			log.Println("No se encontró archivo .env, usando variables de entorno del sistema")
		}
	}

	dbURL := os.Getenv("DB_URL")
	if dbURL == "" {
		log.Fatal("DB_URL no está configurada.")
	}

	jwtSecret := os.Getenv("JWT_SECRET")
	if jwtSecret == "" {
		log.Fatal("JWT_SECRET no está configurada.")
	}

	log.Println("Variables de entorno cargadas correctamente")

	db, err := sql.Open("postgres", dbURL)
	if err != nil {
		log.Fatal("Error al conectar a la DB:", err)
	}
	defer db.Close()

	if err = db.Ping(); err != nil {
		log.Fatal("Error al hacer ping a la DB:", err)
	}
	log.Println("Conectado a la DB exitosamente")

	// Inicializar Auth
	authRepo := &auth.UserRepository{DB: db}
	authService := &auth.AuthService{Repo: authRepo}
	authHandler := &auth.AuthHandler{Service: authService}

	// Inicializar Teams
	teamsRepo := &teams.TeamRepository{DB: db}
	teamsService := &teams.TeamService{Repo: teamsRepo}
	teamsHandler := &teams.TeamHandler{Service: teamsService}

	// Inicializar Channels
	channelsRepo := &channels.ChannelRepository{DB: db}
	channelsService := &channels.ChannelService{Repo: channelsRepo}
	channelsHandler := &channels.ChannelHandler{Service: channelsService}


	// Inicializar Chat
	hub := chat.NewHub()
	chatHandler := chat.NewHandler(db, jwtSecret, hub)

	// Router
	r := mux.NewRouter()

	// -----------------------------------------
	// WebSocket
	// -----------------------------------------
	r.HandleFunc("/ws/channel/{channel_id}", chatHandler.ServeWS)

	// -----------------------------------------
	// Rutas REST
	// -----------------------------------------
	// Públicas
	auth.RegisterRoutes(r, authHandler)

	// Protegidas con JWT
	teams.RegisterRoutes(r, teamsHandler, auth.JWTMiddleware)
	channels.RegisterRoutes(r, channelsHandler, auth.JWTMiddleware)
	dms.RegisterDMSRoutes(r, db)
	users.RegisterUserRoutes(r, db)

	// Puerto
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	log.Println("===========================================")
	log.Printf("🚀 Servidor corriendo en puerto %s\n", port)
	log.Println("Página de inicio disponible en http://localhost:" + port)
	log.Println("===========================================")

	if err := http.ListenAndServe(":"+port, r); err != nil {
		log.Fatal("Error al iniciar el servidor:", err)
	}
}