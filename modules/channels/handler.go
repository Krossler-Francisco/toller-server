// modules/channels/handler.go
package channels

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/gorilla/mux"
)

type ChannelHandler struct {
	Service *ChannelService
}

type CreateChannelRequest struct {
	Name string `json:"name"`
}

type AddMemberRequest struct {
	UserID int    `json:"user_id"`
	Role   string `json:"role"`
}

type UpdateChannelRequest struct {
	Name string `json:"name"`
}

// CreateChannel cria um novo canal
func (h *ChannelHandler) CreateChannel(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	teamID, err := strconv.Atoi(vars["team_id"])
	if err != nil {
		http.Error(w, "ID do time inválido", http.StatusBadRequest)
		return
	}

	userID, ok := r.Context().Value("user_id").(int)
	if !ok {
		http.Error(w, "Usuário não autenticado", http.StatusUnauthorized)
		return
	}

	var req CreateChannelRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Dados inválidos", http.StatusBadRequest)
		return
	}

	channel, err := h.Service.CreateChannel(req.Name, teamID, userID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(channel)
}

// GetChannelsByTeam lista todos os canais de um time
func (h *ChannelHandler) GetChannelsByTeam(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	teamID, err := strconv.Atoi(vars["team_id"])
	if err != nil {
		http.Error(w, "ID do time inválido", http.StatusBadRequest)
		return
	}

	userID, ok := r.Context().Value("user_id").(int)
	if !ok {
		http.Error(w, "Usuário não autenticado", http.StatusUnauthorized)
		return
	}

	channels, err := h.Service.GetChannelsByTeam(teamID, userID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(channels)
}

// GetChannelByID retorna um canal específico
func (h *ChannelHandler) GetChannelByID(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	channelID, err := strconv.Atoi(vars["channel_id"])
	if err != nil {
		http.Error(w, "ID do canal inválido", http.StatusBadRequest)
		return
	}

	userID, ok := r.Context().Value("user_id").(int)
	if !ok {
		http.Error(w, "Usuário não autenticado", http.StatusUnauthorized)
		return
	}

	channel, err := h.Service.GetChannelByID(channelID, userID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusForbidden)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(channel)
}

// AddMember adiciona um membro ao canal
func (h *ChannelHandler) AddMember(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	channelID, err := strconv.Atoi(vars["channel_id"])
	if err != nil {
		http.Error(w, "ID do canal inválido", http.StatusBadRequest)
		return
	}

	requestingUserID, ok := r.Context().Value("user_id").(int)
	if !ok {
		http.Error(w, "Usuário não autenticado", http.StatusUnauthorized)
		return
	}

	var req AddMemberRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Dados inválidos", http.StatusBadRequest)
		return
	}

	err = h.Service.AddMemberToChannel(channelID, req.UserID, requestingUserID, req.Role)
	if err != nil {
		http.Error(w, err.Error(), http.StatusForbidden)
		return
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{"message": "Membro adicionado com sucesso"})
}

// RemoveMember remove um membro do canal
func (h *ChannelHandler) RemoveMember(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	channelID, err := strconv.Atoi(vars["channel_id"])
	if err != nil {
		http.Error(w, "ID do canal inválido", http.StatusBadRequest)
		return
	}

	userID, err := strconv.Atoi(vars["user_id"])
	if err != nil {
		http.Error(w, "ID do usuário inválido", http.StatusBadRequest)
		return
	}

	requestingUserID, ok := r.Context().Value("user_id").(int)
	if !ok {
		http.Error(w, "Usuário não autenticado", http.StatusUnauthorized)
		return
	}

	err = h.Service.RemoveMemberFromChannel(channelID, userID, requestingUserID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusForbidden)
		return
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{"message": "Membro removido com sucesso"})
}

// GetChannelMembers lista todos os membros de um canal
func (h *ChannelHandler) GetChannelMembers(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	channelID, err := strconv.Atoi(vars["channel_id"])
	if err != nil {
		http.Error(w, "ID do canal inválido", http.StatusBadRequest)
		return
	}

	requestingUserID, ok := r.Context().Value("user_id").(int)
	if !ok {
		http.Error(w, "Usuário não autenticado", http.StatusUnauthorized)
		return
	}

	members, err := h.Service.GetChannelMembers(channelID, requestingUserID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusForbidden)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(members)
}

// DeleteChannel deleta um canal
func (h *ChannelHandler) DeleteChannel(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	channelID, err := strconv.Atoi(vars["channel_id"])
	if err != nil {
		http.Error(w, "ID do canal inválido", http.StatusBadRequest)
		return
	}

	userID, ok := r.Context().Value("user_id").(int)
	if !ok {
		http.Error(w, "Usuário não autenticado", http.StatusUnauthorized)
		return
	}

	err = h.Service.DeleteChannel(channelID, userID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusForbidden)
		return
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{"message": "Canal deletado com sucesso"})
}

// UpdateChannel atualiza um canal
func (h *ChannelHandler) UpdateChannel(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	channelID, err := strconv.Atoi(vars["channel_id"])
	if err != nil {
		http.Error(w, "ID do canal inválido", http.StatusBadRequest)
		return
	}

	userID, ok := r.Context().Value("user_id").(int)
	if !ok {
		http.Error(w, "Usuário não autenticado", http.StatusUnauthorized)
		return
	}

	var req UpdateChannelRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Dados inválidos", http.StatusBadRequest)
		return
	}

	err = h.Service.UpdateChannelName(channelID, req.Name, userID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusForbidden)
		return
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{"message": "Canal atualizado com sucesso"})
}
