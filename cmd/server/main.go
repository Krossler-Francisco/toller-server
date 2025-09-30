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

	"toller-server/internal/auth"
	"toller-server/internal/teams"
)

func main() {
	// Cargar .env solo si existe (en local)
	// En producción (Render), las variables ya estarán configuradas
	envPath := filepath.Join("..", "..", ".env")
	if err := godotenv.Load(envPath); err != nil {
		// Intenta cargar desde la raíz también
		if err := godotenv.Load(".env"); err != nil {
			log.Println("No se encontró archivo .env, usando variables de entorno del sistema")
		}
	}

	// Verificar que DB_URL existe
	dbURL := os.Getenv("DB_URL")
	if dbURL == "" {
		log.Fatal("DB_URL no está configurada. Configura las variables de entorno.")
	}

	jwtSecret := os.Getenv("JWT_SECRET")
	if jwtSecret == "" {
		log.Fatal("JWT_SECRET no está configurada. Configura las variables de entorno.")
	}

	log.Println("Variables de entorno cargadas correctamente")

	db, err := sql.Open("postgres", dbURL)
	if err != nil {
		log.Fatal("Error al conectar a la DB:", err)
	}
	defer db.Close()

	// Verificar la conexión
	if err = db.Ping(); err != nil {
		log.Fatal("Error al hacer ping a la DB:", err)
	}

	// Inicializar Auth
	authRepo := &auth.UserRepository{DB: db}
	authService := &auth.AuthService{Repo: authRepo}
	authHandler := &auth.AuthHandler{Service: authService}

	// Inicializar Teams
	teamsRepo := &teams.TeamRepository{DB: db}
	teamsService := &teams.TeamService{Repo: teamsRepo}
	teamsHandler := &teams.TeamHandler{Service: teamsService}

	// Router
	r := mux.NewRouter()

	// Rutas públicas (sin autenticación)
	auth.RegisterRoutes(r, authHandler)

	// Rutas protegidas (con autenticación JWT)
	teams.RegisterRoutes(r, teamsHandler, auth.JWTMiddleware)

	// Usar el puerto de Render si está disponible, si no usar 8080
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	log.Println("Conectado a la DB exitosamente")
	log.Println("===========================================")
	log.Println("📋 Rutas disponibles:")
	log.Println("   POST   /register")
	log.Println("   POST   /login")
	log.Println("   POST   /teams (requiere auth)")
	log.Println("   GET    /teams (requiere auth)")
	log.Println("   GET    /teams/{id} (requiere auth)")
	log.Println("   PUT    /teams/{id} (requiere auth)")
	log.Println("   GET    /teams/{id}/members (requiere auth)")
	log.Println("   POST   /teams/{id}/members (requiere auth)")
	log.Println("   DELETE /teams/{id}/members/{user_id} (requiere auth)")
	log.Println("   POST   /teams/{id}/leave (requiere auth)")
	log.Println("===========================================")
	log.Printf("🚀 Servidor corriendo en puerto %s\n", port)

	if err := http.ListenAndServe(":"+port, r); err != nil {
		log.Fatal("Error al iniciar el servidor:", err)
	}
}
