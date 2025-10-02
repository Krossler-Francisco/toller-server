package friends

import (
	"database/sql"

	"toller-server/modules/auth"

	"github.com/gorilla/mux"
)

func RegisterFriendRoutes(router *mux.Router, db *sql.DB) {
	h := NewFriendHandler(db)

	s := router.PathPrefix("/api/v1").Subrouter()
	s.Use(auth.JWTMiddleware)

	s.HandleFunc("/friends/requests", h.SendFriendRequestHandler).Methods("POST")
	s.HandleFunc("/friends/requests/{friendID:[0-9]+}", h.UpdateFriendRequestHandler).Methods("PUT")
	s.HandleFunc("/friends", h.ListFriendsHandler).Methods("GET")
	s.HandleFunc("/friends/requests/pending", h.ListPendingRequestsHandler).Methods("GET")
}
