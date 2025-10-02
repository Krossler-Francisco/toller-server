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

	token, user, err := h.Service.Login(req.Email, req.Password)
	if err != nil {
		http.Error(w, err.Error(), http.StatusUnauthorized)
		return
	}

	resp := map[string]interface{}{
		"token": token,
		"user": map[string]interface{}{
			"id":       user.ID,
			"username": user.Username,
			"email":    user.Email,
		},
	}
	json.NewEncoder(w).Encode(resp)
}

// Registrar rutas
func RegisterRoutes(r *mux.Router, handler *AuthHandler) {

	api := r.PathPrefix("/api/v1/auth").Subrouter()
	api.HandleFunc("/register", handler.RegisterHandler).Methods("POST")
	api.HandleFunc("/login", handler.LoginHandler).Methods("POST")
}
