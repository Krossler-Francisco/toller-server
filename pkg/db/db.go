package db

import (
	"context"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

func ConnectDB() (*pgxpool.Pool, error) {
	// Leer variables de entorno
	dsn := os.Getenv("DATABASE_URL")
	if dsn == "" {
		// formato de conexión manual
		dsn = "postgres://postgres:1234@localhost:5432/postgres"
	}

	// Configurar pool
	config, err := pgxpool.ParseConfig(dsn)
	if err != nil {
		return nil, fmt.Errorf("error parseando config: %w", err)
	}

	// Intentar conexión
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	db, err := pgxpool.NewWithConfig(ctx, config)
	if err != nil {
		return nil, fmt.Errorf("no se pudo conectar a la DB: %w", err)
	}

	// Probar ping
	if err := db.Ping(ctx); err != nil {
		return nil, fmt.Errorf("DB no responde: %w", err)
	}

	log.Println("DB conectada")
	return db, nil
}
