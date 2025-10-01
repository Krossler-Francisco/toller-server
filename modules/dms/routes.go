package dms

import (
	"database/sql"

	"github.com/gorilla/mux"
	"toller-server/modules/auth"
)

func RegisterDMSRoutes(router *mux.Router, db *sql.DB) {
	h := NewDMHandler(db)

	s := router.PathPrefix("/api/v1").Subrouter()
	s.Use(auth.JWTMiddleware)

	s.HandleFunc("/dms", h.CreateDMHandler).Methods("POST")
	s.HandleFunc("/dms", h.ListDMsHandler).Methods("GET")
	s.HandleFunc("/dms/{channelID:[0-9]+}/messages", h.GetMessagesHandler).Methods("GET")
	s.HandleFunc("/dms/{channelID:[0-9]+}/read", h.MarkAsReadHandler).Methods("POST")
}
