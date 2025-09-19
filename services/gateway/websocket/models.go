package websocket

import (
	"time"

	"github.com/ammysap/plivo-pub-sub/pubsub"
)

// WebSocket Message Types
type WSMessageType string

const (
	WSMessageTypeSubscribe   WSMessageType = "subscribe"
	WSMessageTypeUnsubscribe WSMessageType = "unsubscribe"
	WSMessageTypePublish     WSMessageType = "publish"
	WSMessageTypePing        WSMessageType = "ping"
)

type WSResponseType string

const (
	WSResponseTypeAck   WSResponseType = "ack"
	WSResponseTypeEvent WSResponseType = "event"
	WSResponseTypeError WSResponseType = "error"
	WSResponseTypePong  WSResponseType = "pong"
	WSResponseTypeInfo  WSResponseType = "info"
)

// WebSocket Request Message
type WSRequest struct {
	Type      WSMessageType   `json:"type"`
	Topic     string          `json:"topic,omitempty"`
	Message   *pubsub.Message `json:"message,omitempty"`
	ClientID  string          `json:"client_id,omitempty"`
	LastN     int             `json:"last_n,omitempty"`
	RequestID string          `json:"request_id,omitempty"`
}

// WebSocket Response Message
type WSResponse struct {
	Type      WSResponseType  `json:"type"`
	RequestID string          `json:"request_id,omitempty"`
	Topic     string          `json:"topic,omitempty"`
	Message   *pubsub.Message `json:"message,omitempty"`
	Error     *WSError        `json:"error,omitempty"`
	Status    string          `json:"status,omitempty"`
	Msg       string          `json:"msg,omitempty"`
	Timestamp time.Time       `json:"ts"`
}

// WebSocket Error
type WSError struct {
	Code    string `json:"code"`
	Message string `json:"message"`
}

// Error Codes
const (
	ErrorCodeBadRequest    = "BAD_REQUEST"
	ErrorCodeTopicNotFound = "TOPIC_NOT_FOUND"
	ErrorCodeSlowConsumer  = "SLOW_CONSUMER"
	ErrorCodeUnauthorized  = "UNAUTHORIZED"
	ErrorCodeInternal      = "INTERNAL"
)
