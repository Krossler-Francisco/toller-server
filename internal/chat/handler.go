package chat

import (
	"net/http"

	"github.com/gorilla/mux"
)

func RegisterRoutes(r *mux.Router) {
	r.HandleFunc("/chat", CreateChat).Methods("POST")
	r.HandleFunc("/chat/{id}", GetChat).Methods("GET")
	r.HandleFunc("/chat/{id}/message", SendMessage).Methods("POST")
	r.HandleFunc("/chat/{id}/messages", GetMessages).Methods("GET")
}

func CreateChat(w http.ResponseWriter, r *http.Request)  {}
func GetChat(w http.ResponseWriter, r *http.Request)     {}
func SendMessage(w http.ResponseWriter, r *http.Request) {}
func GetMessages(w http.ResponseWriter, r *http.Request) {}
