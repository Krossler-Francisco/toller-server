package teams

import (
	"net/http"

	"github.com/gorilla/mux"
)

// Registra las rutas del módulo de equipos bajo /api/v1/teams
func RegisterRoutes(router *mux.Router, handler *TeamHandler, authMiddleware func(http.Handler) http.Handler) {
	s := router.PathPrefix("/api/v1").Subrouter()
	s.Use(authMiddleware)

	s.HandleFunc("/teams", handler.GetUserTeams).Methods("GET")
	s.HandleFunc("/teams", handler.CreateTeam).Methods("POST")
	s.HandleFunc("/teams/{id:[0-9]+}", handler.GetTeam).Methods("GET")
	s.HandleFunc("/teams/{id:[0-9]+}/members", handler.GetTeamMembers).Methods("GET")
	s.HandleFunc("/teams/{id:[0-9]+}", handler.UpdateTeam).Methods("PUT", "PATCH")
	s.HandleFunc("/teams/{team_id:[0-9]+}/members", handler.AddMember).Methods("POST")
	s.HandleFunc("/teams/{team_id:[0-9]+}/members/{user_id:[0-9]+}", handler.RemoveMember).Methods("DELETE")
}
