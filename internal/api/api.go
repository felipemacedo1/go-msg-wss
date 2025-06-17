package api

import (
	"context"
	"encoding/json"
	"log/slog"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/felipemacedo1/go-msg-wss/internal/store/pgstore"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/cors"
	"github.com/google/uuid"
	"github.com/gorilla/websocket"
)

type apiHandler struct {
	q           *pgstore.Queries
	r           *chi.Mux
	upgrader    websocket.Upgrader
	subscribers map[string]map[*websocket.Conn]context.CancelFunc
	mu          *sync.Mutex
}

func (h apiHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	h.r.ServeHTTP(w, r)
}

func NewHandler(q *pgstore.Queries) http.Handler {
	a := apiHandler{
		q:           q,
		upgrader:    websocket.Upgrader{CheckOrigin: func(r *http.Request) bool { return true }},
		subscribers: make(map[string]map[*websocket.Conn]context.CancelFunc),
		mu:          &sync.Mutex{},
	}

	r := chi.NewRouter()
	r.Use(middleware.RequestID, middleware.Recoverer, middleware.Logger)

	r.Use(cors.Handler(cors.Options{
		AllowedOrigins:   []string{"https://*", "http://*"},
		AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "OPTIONS", "PATCH"},
		AllowedHeaders:   []string{"Accept", "Authorization", "Content-Type", "X-CSRF-Token"},
		ExposedHeaders:   []string{"Link"},
		AllowCredentials: false,
		MaxAge:           300,
	}))

	r.Get("/subscribe/{room_id}", a.handleSubscribe)

	r.Route("/api", func(r chi.Router) {
		r.Route("/rooms", func(r chi.Router) {
			r.Post("/", a.handleCreateRoom)
			r.Get("/", a.handleGetRooms)

			r.Route("/{room_id}", func(r chi.Router) {
				r.Get("/", a.handleGetRoom)

				r.Route("/messages", func(r chi.Router) {
					r.Post("/", a.handleCreateRoomMessage)
					r.Get("/", a.handleGetRoomMessages)

					r.Route("/{message_id}", func(r chi.Router) {
						r.Get("/", a.handleGetRoomMessage)
						r.Patch("/react", a.handleReactToMessage)
						r.Delete("/react", a.handleRemoveReactFromMessage)
						r.Patch("/answer", a.handleMarkMessageAsAnswered)
					})
				})
			})
		})
	})

	a.r = r
	return a
}

const (
	MessageKindMessageCreated          = "message_created"
	MessageKindMessageRactionIncreased = "message_reaction_increased"
	MessageKindMessageRactionDecreased = "message_reaction_decreased"
	MessageKindMessageAnswered         = "message_answered"
)

type MessageMessageReactionIncreased struct {
	ID    string `json:"id"`
	Count int64  `json:"count"`
}

type MessageMessageReactionDecreased struct {
	ID    string `json:"id"`
	Count int64  `json:"count"`
}

type MessageMessageAnswered struct {
	ID string `json:"id"`
}

type MessageMessageCreated struct {
	ID      string `json:"id"`
	Message string `json:"message"`
	Author  string `json:"author"`
}

type Message struct {
	Kind   string `json:"kind"`
	Value  any    `json:"value"`
	RoomID string `json:"room_id"`
}

func (h apiHandler) notifyClients(msg Message) {
	h.mu.Lock()
	defer h.mu.Unlock()

	slog.Info("notifyClients called", "room_id", msg.RoomID, "kind", msg.Kind)

	subscribers, ok := h.subscribers[msg.RoomID]
	if !ok || len(subscribers) == 0 {
		slog.Warn("notifyClients: no subscribers found", "room_id", msg.RoomID)
		return
	}

	slog.Info("notifyClients: sending to subscribers", "room_id", msg.RoomID, "subscriber_count", len(subscribers))
	var failedConns []*websocket.Conn
	for conn, cancel := range subscribers {
		if err := conn.WriteJSON(msg); err != nil {
			failedConns = append(failedConns, conn)
			cancel()
		}
	}
	for _, conn := range failedConns {
		delete(subscribers, conn)
	}
}

func (h apiHandler) handleSubscribe(w http.ResponseWriter, r *http.Request) {
	slog.Info("handleSubscribe called", "url", r.URL.Path)

	_, rawRoomID, _, ok := h.readRoom(w, r)
	if !ok {
		slog.Warn("handleSubscribe: invalid room", "room_id", chi.URLParam(r, "room_id"))
		return
	}

	slog.Info("handleSubscribe: upgrading to websocket", "room_id", rawRoomID)

	c, err := h.upgrader.Upgrade(w, r, nil)
	if err != nil {
		slog.Warn("failed to upgrade connection", "error", err)
		http.Error(w, "failed to upgrade to ws connection", http.StatusBadRequest)
		return
	}

	defer func() {
		slog.Info("closing websocket connection", "room_id", rawRoomID, "client_ip", r.RemoteAddr)
		c.Close()
	}()

	ctx, cancel := context.WithCancel(r.Context())
	defer cancel()

	// Add client to subscribers
	h.mu.Lock()
	if _, ok := h.subscribers[rawRoomID]; !ok {
		h.subscribers[rawRoomID] = make(map[*websocket.Conn]context.CancelFunc)
		slog.Info("created subscriber map for room", "room_id", rawRoomID)
	}
	slog.Info("new client connected", "room_id", rawRoomID, "client_ip", r.RemoteAddr)
	h.subscribers[rawRoomID][c] = cancel
	slog.Info("subscriber added", "room_id", rawRoomID, "total_subscribers", len(h.subscribers[rawRoomID]))
	h.mu.Unlock()

	// Cleanup when function exits
	defer func() {
		h.mu.Lock()
		delete(h.subscribers[rawRoomID], c)
		slog.Info("client disconnected", "room_id", rawRoomID, "client_ip", r.RemoteAddr, "remaining_subscribers", len(h.subscribers[rawRoomID]))

		// Clean up empty room maps
		if len(h.subscribers[rawRoomID]) == 0 {
			delete(h.subscribers, rawRoomID)
			slog.Info("removed empty room from subscribers", "room_id", rawRoomID)
		}
		h.mu.Unlock()
	}()

	// Set connection timeouts
	c.SetReadDeadline(time.Now().Add(60 * time.Second))
	c.SetPongHandler(func(string) error {
		c.SetReadDeadline(time.Now().Add(60 * time.Second))
		return nil
	})

	// Start ping ticker
	ticker := time.NewTicker(54 * time.Second) // Slightly less than read deadline
	defer ticker.Stop()

	// Ping routine
	go func() {

		for {
			select {
			case <-ticker.C:
				if err := c.WriteMessage(websocket.PingMessage, nil); err != nil {
					slog.Warn("ping failed", "room_id", rawRoomID, "error", err)
					return
				}
			case <-ctx.Done():
				return
			}
			_, msgBytes, err := c.ReadMessage()
			if err != nil {
				if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
					slog.Warn("websocket error", "room_id", rawRoomID, "error", err)
				}
				return
			}
			// Optional: tratar mensagens tipo 'ping'
			var raw map[string]interface{}
			if err := json.Unmarshal(msgBytes, &raw); err == nil {
				if kind, ok := raw["kind"].(string); ok && kind == "client_ping" {
					slog.Info("received client ping", "room_id", rawRoomID, "ip", r.RemoteAddr)
					_ = c.WriteMessage(websocket.TextMessage, []byte(`{"kind":"server_pong"}`))
					continue
				}
			}
		}
	}()

	// Message reading loop (keeps connection alive)
	for {
		select {
		case <-ctx.Done():
			slog.Info("context cancelled, closing connection", "room_id", rawRoomID)
			return
		default:
			// Read messages (even if we don't process them, this keeps connection alive)
			_, _, err := c.ReadMessage()
			if err != nil {
				if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
					slog.Warn("websocket error", "room_id", rawRoomID, "error", err)
				}
				return
			}
			// You can add message processing logic here if needed
		}
	}
}

func (h apiHandler) handleCreateRoom(w http.ResponseWriter, r *http.Request) {
	slog.Info("handleCreateRoom called", "url", r.URL.Path)

	type _body struct {
		Theme string `json:"theme"`
	}
	var body _body
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		slog.Warn("handleCreateRoom: invalid json", "error", err)
		http.Error(w, "invalid json", http.StatusBadRequest)
		return
	}

	slog.Info("handleCreateRoom: received", "theme", body.Theme)

	if strings.TrimSpace(body.Theme) == "" {
		slog.Warn("handleCreateRoom: empty theme")
		http.Error(w, "theme cannot be empty", http.StatusBadRequest)
		return
	}

	roomID, err := h.q.InsertRoom(r.Context(), body.Theme)
	if err != nil {
		slog.Error("failed to insert room", "error", err)
		http.Error(w, "something went wrong", http.StatusInternalServerError)
		return
	}

	slog.Info("handleCreateRoom: room created", "room_id", roomID.String())

	type response struct {
		ID string `json:"id"`
	}

	sendJSON(w, response{ID: roomID.String()})
}

func (h apiHandler) handleGetRooms(w http.ResponseWriter, r *http.Request) {
	rooms, err := h.q.GetRooms(r.Context())
	if err != nil {
		http.Error(w, "something went wrong", http.StatusInternalServerError)
		slog.Error("failed to get rooms", "error", err)
		return
	}

	if rooms == nil {
		rooms = []pgstore.Room{}
	}

	sendJSON(w, rooms)
}

func (h apiHandler) handleGetRoom(w http.ResponseWriter, r *http.Request) {
	room, _, _, ok := h.readRoom(w, r)
	if !ok {
		return
	}

	sendJSON(w, room)
}

func (h apiHandler) handleCreateRoomMessage(w http.ResponseWriter, r *http.Request) {
	_, rawRoomID, roomID, ok := h.readRoom(w, r)
	if !ok {
		return
	}

	slog.Info("handleCreateRoomMessage called", "room_id", rawRoomID)

	type _body struct {
		Message string `json:"message"`
	}
	var body _body
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		http.Error(w, "invalid json", http.StatusBadRequest)
		slog.Warn("handleCreateRoomMessage: invalid json", "error", err)
		return
	}

	slog.Info("handleCreateRoomMessage: received message", "message", body.Message, "room_id", rawRoomID)

	claims := extractClaimsFromJWT(r)
	authorID := "guest"
	authorName := "Guest"

	if claims != nil {
		if sub, ok := claims["sub"].(string); ok {
			authorID = sub
		}
		if name, ok := claims["name"].(string); ok {
			authorName = name
		}
	}

	messageID, err := h.q.InsertMessage(r.Context(), pgstore.InsertMessageParams{
		RoomID:     roomID,
		Message:    body.Message,
		AuthorID:   authorID,
		AuthorName: authorName,
	})
	if err != nil {
		slog.Error("failed to insert message", "error", err, "room_id", rawRoomID)
		http.Error(w, "something went wrong", http.StatusInternalServerError)
		return
	}

	// Buscar a mensagem completa após inserção
	fullMessage, err := h.q.GetMessage(r.Context(), messageID.ID)
	if err != nil {
		slog.Error("failed to retrieve full message", "error", err, "message_id", messageID)
		http.Error(w, "something went wrong", http.StatusInternalServerError)
		return
	}

	slog.Info("handleCreateRoomMessage: message created", "message_id", messageID, "room_id", rawRoomID)

	type response struct {
		ID string `json:"id"`
	}

	sendJSON(w, response{ID: messageID.ID.String()})

	go func() {
		slog.Info("handleCreateRoomMessage: notifying clients", "room_id", rawRoomID)
		h.notifyClients(Message{
			Kind:   MessageKindMessageCreated,
			RoomID: rawRoomID,
			Value:  fullMessage, // Enviar a mensagem completa do banco
		})
	}()
}

func (h apiHandler) handleGetRoomMessages(w http.ResponseWriter, r *http.Request) {
	_, _, roomID, ok := h.readRoom(w, r)
	if !ok {
		return
	}

	messages, err := h.q.GetRoomMessages(r.Context(), roomID)
	if err != nil {
		http.Error(w, "something went wrong", http.StatusInternalServerError)
		slog.Error("failed to get room messages", "error", err)
		return
	}

	if messages == nil {
		messages = []pgstore.Message{}
	}

	sendJSON(w, messages)
}

func (h apiHandler) handleGetRoomMessage(w http.ResponseWriter, r *http.Request) {
	_, _, _, ok := h.readRoom(w, r)
	if !ok {
		return
	}

	messageID := chi.URLParam(r, "message_id")
	parsedMessageID, err := uuid.Parse(messageID)
	if err != nil {
		http.Error(w, "invalid message id", http.StatusBadRequest)
		return
	}

	msg, err := h.q.GetMessage(r.Context(), parsedMessageID)
	if err != nil {
		slog.Error("failed to get message", "message_id", parsedMessageID, "error", err)
		http.Error(w, "failed to get message", http.StatusInternalServerError)
		return
	}

	sendJSON(w, msg)
}

func (h apiHandler) handleReactToMessage(w http.ResponseWriter, r *http.Request) {
	_, rawRoomID, _, ok := h.readRoom(w, r)
	if !ok {
		return
	}

	rawID := chi.URLParam(r, "message_id")
	id, err := uuid.Parse(rawID)
	if err != nil {
		http.Error(w, "invalid message id", http.StatusBadRequest)
		return
	}

	count, err := h.q.ReactToMessage(r.Context(), id)
	if err != nil {
		http.Error(w, "something went wrong", http.StatusInternalServerError)
		slog.Error("failed to react to message", "error", err)
		return
	}

	type response struct {
		Count int64 `json:"count"`
	}

	sendJSON(w, response{Count: count})

	go h.notifyClients(Message{
		Kind:   MessageKindMessageRactionIncreased,
		RoomID: rawRoomID,
		Value: MessageMessageReactionIncreased{
			ID:    rawID,
			Count: count,
		},
	})
}

func (h apiHandler) handleRemoveReactFromMessage(w http.ResponseWriter, r *http.Request) {
	_, rawRoomID, _, ok := h.readRoom(w, r)
	if !ok {
		return
	}

	rawID := chi.URLParam(r, "message_id")
	id, err := uuid.Parse(rawID)
	if err != nil {
		http.Error(w, "invalid message id", http.StatusBadRequest)
		return
	}

	count, err := h.q.RemoveReactionFromMessage(r.Context(), id)
	if err != nil {
		http.Error(w, "something went wrong", http.StatusInternalServerError)
		slog.Error("failed to react to message", "error", err)
		return
	}

	type response struct {
		Count int64 `json:"count"`
	}

	sendJSON(w, response{Count: count})

	go h.notifyClients(Message{
		Kind:   MessageKindMessageRactionDecreased,
		RoomID: rawRoomID,
		Value: MessageMessageReactionDecreased{
			ID:    rawID,
			Count: count,
		},
	})
}

func (h apiHandler) handleMarkMessageAsAnswered(w http.ResponseWriter, r *http.Request) {
	_, rawRoomID, _, ok := h.readRoom(w, r)
	if !ok {
		return
	}

	rawID := chi.URLParam(r, "message_id")
	id, err := uuid.Parse(rawID)
	if err != nil {
		http.Error(w, "invalid message id", http.StatusBadRequest)
		return
	}

	err = h.q.MarkMessageAsAnswered(r.Context(), id)
	if err != nil {
		http.Error(w, "something went wrong", http.StatusInternalServerError)
		slog.Error("failed to react to message", "error", err)
		return
	}

	w.WriteHeader(http.StatusOK)

	go h.notifyClients(Message{
		Kind:   MessageKindMessageAnswered,
		RoomID: rawRoomID,
		Value: MessageMessageAnswered{
			ID: rawID,
		},
	})
}
