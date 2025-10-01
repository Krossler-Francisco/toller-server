package chat

import (
	"database/sql"
	"errors"
	"log"
	"net/http"
	"strconv"

	"github.com/golang-jwt/jwt/v5"
	"github.com/gorilla/mux"
	"github.com/gorilla/websocket"
)

type ChatHandler struct {
	Hub       *Hub
	Repo      *Repository
	JWTSecret []byte
	Upgrader  websocket.Upgrader
}

func NewHandler(db *sql.DB, jwtSecret string, hub *Hub) *ChatHandler {
	return &ChatHandler{
		Hub:       hub,
		Repo:      NewRepository(db),
		JWTSecret: []byte(jwtSecret),
		Upgrader: websocket.Upgrader{
			CheckOrigin: func(r *http.Request) bool { return true }, // ajustar em prod
		},
	}
}

// parse token (aceita token no query param "token" ou header Authorization: Bearer ...)
func (h *ChatHandler) parseTokenGetUserID(r *http.Request) (int64, error) {
	tokenStr := r.URL.Query().Get("token")
	if tokenStr == "" {
		auth := r.Header.Get("Authorization")
		if len(auth) > 7 && auth[:7] == "Bearer " {
			tokenStr = auth[7:]
		}
	}
	if tokenStr == "" {
		return 0, errors.New("token não fornecido")
	}

	token, err := jwt.Parse(tokenStr, func(t *jwt.Token) (interface{}, error) {
		// opcional: checar método de assinatura
		return h.JWTSecret, nil
	})
	if err != nil || !token.Valid {
		return 0, errors.New("token inválido")
	}
	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		return 0, errors.New("claims inválidos")
	}
	uidFloat, ok := claims["user_id"].(float64)
	if !ok {
		// tentar por "sub"
		if sub, ok2 := claims["sub"].(string); ok2 {
			if id, err := strconv.ParseInt(sub, 10, 64); err == nil {
				return id, nil
			}
		}
		return 0, errors.New("user_id não encontrado no token")
	}
	return int64(uidFloat), nil
}

func (h *ChatHandler) ServeWS(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	chStr := vars["channel_id"]
	if chStr == "" {
		http.Error(w, "channel_id requerido", http.StatusBadRequest)
		return
	}
	channelID, err := strconv.ParseInt(chStr, 10, 64)
	if err != nil {
		http.Error(w, "channel_id inválido", http.StatusBadRequest)
		return
	}

	userID, err := h.parseTokenGetUserID(r)
	if err != nil {
		http.Error(w, "autenticação falhou: "+err.Error(), http.StatusUnauthorized)
		return
	}

	conn, err := h.Upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Println("upgrade error:", err)
		return
	}

	client := &Client{
		conn:      conn,
		send:      make(chan OutgoingMessage, 256),
		userID:    userID,
		channelID: channelID,
		hub:       h.Hub,
		repo:      h.Repo,
	}

	// registrar
	h.Hub.Register(client, channelID)

	// enviar últimas mensagens (ex.: 50)
	if msgs, err := h.Repo.LoadLastMessages(channelID, 50); err == nil {
		for _, m := range msgs {
			client.send <- m
		}
	} else {
		// opcional: log
		log.Println("LoadLastMessages error:", err)
	}

	// iniciar pumps
	go client.writePump()
	go client.readPump()
}
