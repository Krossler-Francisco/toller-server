package users

import (
	"database/sql"

	"github.com/gorilla/mux"
	"toller-server/modules/auth"
)

func RegisterUserRoutes(router *mux.Router, db *sql.DB) {
	h := NewUserHandler(db)

	s := router.PathPrefix("/api/v1/users/").Subrouter()
	s.Use(auth.JWTMiddleware)

	s.HandleFunc("", h.GetAllUsersHandler).Methods("GET")
	s.HandleFunc("/{id:[0-9]+}", h.GetUserByIDHandler).Methods("GET")
	s.HandleFunc("/search", h.SearchUsersHandler).Methods("GET")
}
