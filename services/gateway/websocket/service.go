package websocket

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/ammysap/plivo-pub-sub/logging"
	"github.com/ammysap/plivo-pub-sub/pubsub"
	"github.com/gorilla/websocket"
)

// Service interface for WebSocket operations
type Service interface {
	HandleWebSocketConnection(conn *websocket.Conn, ctx context.Context)
}

// WebSocketHandler handles WebSocket connections for pub/sub
type WebSocketHandler struct {
	pubsubService pubsub.Service
	clients       map[string]*Client // client_id -> client
	clientsMu     sync.RWMutex
	shutdown      chan struct{}
}

// Client represents a WebSocket client connection
type Client struct {
	ID            string
	Conn          *websocket.Conn
	Subscriptions map[string]*pubsub.Subscriber // topic -> subscriber
	mu            sync.RWMutex
	done          chan struct{}
}

// service implements the Service interface
type service struct {
	handler *WebSocketHandler
}

// NewService creates a new WebSocket service
func NewService() Service {
	handler := &WebSocketHandler{
		pubsubService: pubsub.GetService(),
		clients:       make(map[string]*Client),
		shutdown:      make(chan struct{}),
	}

	return &service{
		handler: handler,
	}
}

// HandleWebSocketConnection handles WebSocket connections
func (s *service) HandleWebSocketConnection(conn *websocket.Conn, ctx context.Context) {
	s.handler.HandleWebSocketConnection(conn, ctx)
}

// HandleWebSocketConnection handles WebSocket connections
func (h *WebSocketHandler) HandleWebSocketConnection(conn *websocket.Conn, ctx context.Context) {
	defer conn.Close()

	// Get authenticated user ID from context
	userID, ok := ctx.Value(ctxKeyUserID).(string)
	if !ok || userID == "" {
		logging.WithContext(ctx).Errorw("No authenticated user ID in context")
		return
	}

	// Use user ID as client ID for authenticated connections
	clientID := userID

	client := &Client{
		ID:            clientID,
		Conn:          conn,
		Subscriptions: make(map[string]*pubsub.Subscriber),
		done:          make(chan struct{}),
	}

	// Register client
	h.clientsMu.Lock()
	h.clients[clientID] = client
	h.clientsMu.Unlock()

	// Cleanup on disconnect
	defer func() {
		h.clientsMu.Lock()
		delete(h.clients, clientID)
		h.clientsMu.Unlock()

		// Unsubscribe from all topics
		client.mu.RLock()
		for topicName := range client.Subscriptions {
			h.pubsubService.Unsubscribe(ctx, topicName, clientID)
		}
		client.mu.RUnlock()

		close(client.done)
	}()

	// Start message sender goroutine
	go h.messageSender(client)

	// Handle incoming messages
	for {
		select {
		case <-h.shutdown:
			return
		case <-client.done:
			return
		default:
			var req WSRequest
			err := conn.ReadJSON(&req)
			if err != nil {
				if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
					logging.WithContext(ctx).Errorw("WebSocket read error", "error", err, "client_id", clientID)
				}
				return
			}

			h.handleMessage(ctx, client, &req)
		}
	}
}

// handleMessage processes incoming WebSocket messages
func (h *WebSocketHandler) handleMessage(ctx context.Context, client *Client, req *WSRequest) {
	log := logging.WithContext(ctx)

	response := &WSResponse{
		RequestID: req.RequestID,
		Timestamp: time.Now(),
	}

	switch req.Type {
	case WSMessageTypeSubscribe:
		h.handleSubscribe(ctx, client, req, response)
	case WSMessageTypeUnsubscribe:
		h.handleUnsubscribe(ctx, client, req, response)
	case WSMessageTypePublish:
		h.handlePublish(ctx, client, req, response)
	case WSMessageTypePing:
		h.handlePing(ctx, client, req, response)
	default:
		response.Type = WSResponseTypeError
		response.Error = &WSError{
			Code:    ErrorCodeBadRequest,
			Message: fmt.Sprintf("Unknown message type: %s", req.Type),
		}
	}

	// Send response
	if err := client.Conn.WriteJSON(response); err != nil {
		log.Errorw("Failed to send WebSocket response", "error", err, "client_id", client.ID)
	}
}

// handleSubscribe handles subscribe requests
func (h *WebSocketHandler) handleSubscribe(ctx context.Context, client *Client, req *WSRequest, response *WSResponse) {
	log := logging.WithContext(ctx)

	if req.Topic == "" {
		response.Type = WSResponseTypeError
		response.Error = &WSError{
			Code:    ErrorCodeBadRequest,
			Message: "topic is required for subscribe",
		}
		return
	}

	// Use authenticated user ID as client ID
	clientID := client.ID

	subscriber, err := h.pubsubService.Subscribe(ctx, req.Topic, clientID, req.LastN)
	if err != nil {
		response.Type = WSResponseTypeError
		if err.Error() == fmt.Sprintf("topic %s not found", req.Topic) {
			response.Error = &WSError{
				Code:    ErrorCodeTopicNotFound,
				Message: err.Error(),
			}
		} else {
			response.Error = &WSError{
				Code:    ErrorCodeInternal,
				Message: err.Error(),
			}
		}
		return
	}

	// Store subscription
	client.mu.Lock()
	client.Subscriptions[req.Topic] = subscriber
	client.mu.Unlock()

	response.Type = WSResponseTypeAck
	response.Topic = req.Topic
	response.Status = "ok"

	log.Info("Client subscribed to topic", "client_id", clientID, "topic", req.Topic, "last_n", req.LastN)
}

// handleUnsubscribe handles unsubscribe requests
func (h *WebSocketHandler) handleUnsubscribe(ctx context.Context, client *Client, req *WSRequest, response *WSResponse) {
	log := logging.WithContext(ctx)

	if req.Topic == "" {
		response.Type = WSResponseTypeError
		response.Error = &WSError{
			Code:    ErrorCodeBadRequest,
			Message: "topic is required for unsubscribe",
		}
		return
	}

	// Use authenticated user ID as client ID
	clientID := client.ID

	err := h.pubsubService.Unsubscribe(ctx, req.Topic, clientID)
	if err != nil {
		response.Type = WSResponseTypeError
		if err.Error() == fmt.Sprintf("topic %s not found", req.Topic) {
			response.Error = &WSError{
				Code:    ErrorCodeTopicNotFound,
				Message: err.Error(),
			}
		} else {
			response.Error = &WSError{
				Code:    ErrorCodeInternal,
				Message: err.Error(),
			}
		}
		return
	}

	// Remove subscription
	client.mu.Lock()
	delete(client.Subscriptions, req.Topic)
	client.mu.Unlock()

	response.Type = WSResponseTypeAck
	response.Topic = req.Topic
	response.Status = "ok"

	log.Info("Client unsubscribed from topic", "client_id", clientID, "topic", req.Topic)
}

// handlePublish handles publish requests
func (h *WebSocketHandler) handlePublish(ctx context.Context, client *Client, req *WSRequest, response *WSResponse) {
	log := logging.WithContext(ctx)

	if req.Topic == "" || req.Message == nil {
		response.Type = WSResponseTypeError
		response.Error = &WSError{
			Code:    ErrorCodeBadRequest,
			Message: "topic and message are required for publish",
		}
		return
	}

	// Validate message ID
	if req.Message.ID == "" {
		response.Type = WSResponseTypeError
		response.Error = &WSError{
			Code:    ErrorCodeBadRequest,
			Message: "message.id is required",
		}
		return
	}

	err := h.pubsubService.Publish(ctx, req.Topic, req.Message)
	if err != nil {
		response.Type = WSResponseTypeError
		if err.Error() == fmt.Sprintf("topic %s not found", req.Topic) {
			response.Error = &WSError{
				Code:    ErrorCodeTopicNotFound,
				Message: err.Error(),
			}
		} else {
			response.Error = &WSError{
				Code:    ErrorCodeInternal,
				Message: err.Error(),
			}
		}
		return
	}

	response.Type = WSResponseTypeAck
	response.Topic = req.Topic
	response.Status = "ok"

	log.Info("Message published", "topic", req.Topic, "message_id", req.Message.ID)
}

// handlePing handles ping requests
func (h *WebSocketHandler) handlePing(ctx context.Context, client *Client, req *WSRequest, response *WSResponse) {
	response.Type = WSResponseTypePong
	logging.WithContext(ctx).Debug("Received ping from client", "client_id", client.ID)
}

// messageSender sends messages from subscriber channels to WebSocket
func (h *WebSocketHandler) messageSender(client *Client) {
	for {
		select {
		case <-h.shutdown:
			return
		case <-client.done:
			return
		default:
			// Check all subscriptions for new messages
			client.mu.RLock()
			subscriptions := make([]*pubsub.Subscriber, 0, len(client.Subscriptions))
			for _, subscriber := range client.Subscriptions {
				subscriptions = append(subscriptions, subscriber)
			}
			client.mu.RUnlock()

			// Use select with default to avoid blocking
			messageSent := false
			for _, subscriber := range subscriptions {
				select {
				case message := <-subscriber.MessageChan:
					response := &WSResponse{
						Type:      WSResponseTypeEvent,
						Topic:     message.Topic,
						Message:   message,
						Timestamp: time.Now(),
					}

					if err := client.Conn.WriteJSON(response); err != nil {
						logging.WithContext(context.Background()).Errorw("Failed to send event message",
							"error", err, "client_id", client.ID, "topic", message.Topic)
						return
					}
					messageSent = true
				default:
					// No message available, continue
				}
			}

			// If no messages were sent, sleep briefly to avoid busy waiting
			if !messageSent {
				time.Sleep(10 * time.Millisecond)
			}
		}
	}
}

// Shutdown gracefully shuts down the WebSocket handler
func (h *WebSocketHandler) Shutdown() {
	close(h.shutdown)

	// Close all client connections
	h.clientsMu.RLock()
	for _, client := range h.clients {
		client.Conn.Close()
		close(client.done)
	}
	h.clientsMu.RUnlock()
}
