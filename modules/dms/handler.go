package dms

import (
	"database/sql"
	"encoding/json"
	"log"
	"net/http"
	"strconv"

	"github.com/gorilla/mux"
)

type DMHandler struct {
	Service *DMService
}

func NewDMHandler(db *sql.DB) *DMHandler {
	repo := NewDMRepository(db)
	service := NewDMService(repo)
	return &DMHandler{Service: service}
}

func (h *DMHandler) CreateDMHandler(w http.ResponseWriter, r *http.Request) {

	userID, ok := r.Context().Value("user_id").(int)
	if !ok {
		log.Println("[DM] User not authenticated")
		http.Error(w, "User not authenticated", http.StatusUnauthorized)
		return
	}

	var req struct {
		RecipientID int `json:"recipient_id"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		log.Println("[DM] Invalid request body:", err)
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	log.Printf("[DM] CreateDMHandler: userID=%d, recipientID=%d\n", userID, req.RecipientID)

	channelID, err := h.Service.CreateDM(userID, req.RecipientID)
	if err != nil {
		log.Printf("[DM] Error creating DM channel: %v\n", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	log.Printf("[DM] DM channel created: channelID=%d\n", channelID)
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]int{"channel_id": channelID})
}

// ListDMsHandler maneja la petición para listar los DMs de un usuario.
func (h *DMHandler) ListDMsHandler(w http.ResponseWriter, r *http.Request) {
	userID, ok := r.Context().Value("user_id").(int)
	if !ok {
		http.Error(w, "Usuario no autenticado", http.StatusUnauthorized)
		return
	}

	dms, err := h.Service.ListDMs(userID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Devolver un array vacío en lugar de null si no hay DMs
	if dms == nil {
		dms = []DMChannelInfo{}
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(dms)
}

func (h *DMHandler) GetMessagesHandler(w http.ResponseWriter, r *http.Request) {
	userID, ok := r.Context().Value("user_id").(int)
	if !ok {
		http.Error(w, "User not authenticated", http.StatusUnauthorized)
		return
	}

	vars := mux.Vars(r)
	channelID, err := strconv.Atoi(vars["channelID"])
	if err != nil {
		http.Error(w, "Invalid channel ID", http.StatusBadRequest)
		return
	}

	messages, err := h.Service.GetMessages(userID, channelID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(messages)
}

func (h *DMHandler) MarkAsReadHandler(w http.ResponseWriter, r *http.Request) {
	userID, ok := r.Context().Value("user_id").(int)
	if !ok {
		http.Error(w, "User not authenticated", http.StatusUnauthorized)
		return
	}

	vars := mux.Vars(r)
	channelID, err := strconv.Atoi(vars["channelID"])
	if err != nil {
		http.Error(w, "Invalid channel ID", http.StatusBadRequest)
		return
	}

	if err := h.Service.MarkAsRead(userID, channelID); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
