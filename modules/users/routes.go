package users

import (
	"database/sql"

	"toller-server/modules/auth"

	"github.com/gorilla/mux"
)

func RegisterUserRoutes(router *mux.Router, db *sql.DB) {
	h := NewUserHandler(db)

	s := router.PathPrefix("/api/v1/").Subrouter()
	s.Use(auth.JWTMiddleware)

	s.HandleFunc("/users", h.GetAllUsersHandler).Methods("GET")
	s.HandleFunc("/users/{id:[0-9]+}", h.GetUserByIDHandler).Methods("GET")
	s.HandleFunc("/users/search", h.SearchUsersHandler).Methods("GET")
	s.HandleFunc("/users/me/{id:[0-9]+}", h.GetUserMeHandler).Methods("GET")
}
