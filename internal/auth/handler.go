package auth

import (
	"encoding/json"
	"net/http"

	"github.com/gorilla/mux"
)

type AuthHandler struct {
	Service *AuthService
}

func (h *AuthHandler) RegisterHandler(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Username string `json:"username"`
		Email    string `json:"email"`
		Password string `json:"password"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "request inválido", http.StatusBadRequest)
		return
	}

	user, err := h.Service.Register(req.Username, req.Email, req.Password)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(user)
}

func (h *AuthHandler) LoginHandler(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Email    string `json:"email"`
		Password string `json:"password"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "request inválido", http.StatusBadRequest)
		return
	}

	token, err := h.Service.Login(req.Email, req.Password)
	if err != nil {
		http.Error(w, err.Error(), http.StatusUnauthorized)
		return
	}

	json.NewEncoder(w).Encode(map[string]string{"token": token})
}

// Registrar rutas
func RegisterRoutes(r *mux.Router, handler *AuthHandler) {
	r.HandleFunc("/register", handler.RegisterHandler).Methods("POST")
	r.HandleFunc("/login", handler.LoginHandler).Methods("POST")
}
