// modules/channels/routes.go
package channels

import (
	"net/http"

	"github.com/gorilla/mux"
)

// RegisterRoutes registra todas as rotas do módulo de canais
func RegisterRoutes(r *mux.Router, handler *ChannelHandler, authMiddleware func(http.Handler) http.Handler) {
	// Subrouter protegido com autenticação
	api := r.PathPrefix("/api/v1").Subrouter()
	api.Use(authMiddleware)

	// Rotas de canais por time
	api.HandleFunc("/teams/{team_id}/channels", handler.CreateChannel).Methods("POST")
	api.HandleFunc("/teams/{team_id}/channels", handler.GetChannelsByTeam).Methods("GET")

	// Rotas de canal específico
	api.HandleFunc("/channels/{channel_id}", handler.GetChannelByID).Methods("GET")
	api.HandleFunc("/channels/{channel_id}", handler.UpdateChannel).Methods("PUT", "PATCH")
	api.HandleFunc("/channels/{channel_id}", handler.DeleteChannel).Methods("DELETE")

	// Rotas de membros do canal
	api.HandleFunc("/channels/{channel_id}/members", handler.GetChannelMembers).Methods("GET")
	api.HandleFunc("/channels/{channel_id}/members", handler.AddMember).Methods("POST")
	api.HandleFunc("/channels/{channel_id}/members/{user_id}", handler.RemoveMember).Methods("DELETE")
}
