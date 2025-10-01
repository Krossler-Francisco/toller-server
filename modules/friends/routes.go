package friends

import (
	"database/sql"

	"github.com/gorilla/mux"
	"toller-server/modules/auth"
)

func RegisterFriendRoutes(router *mux.Router, db *sql.DB) {
	h := NewFriendHandler(db)

	s := router.PathPrefix("/api/v1/friends").Subrouter()
	s.Use(auth.JWTMiddleware)

	s.HandleFunc("/requests", h.SendFriendRequestHandler).Methods("POST")
	s.HandleFunc("/requests/{friendID:[0-9]+}", h.UpdateFriendRequestHandler).Methods("PUT")
	s.HandleFunc("", h.ListFriendsHandler).Methods("GET")
	s.HandleFunc("/requests/pending", h.ListPendingRequestsHandler).Methods("GET")
}
