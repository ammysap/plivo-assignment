package pubsub

import (
	"sync"
	"time"
)

// Configuration constants
const (
	DefaultRingBufferSize    = 100
	DefaultChannelBufferSize = 100
	GracefulShutdownTimeout  = 30 * time.Second
)

// Config holds configurable parameters
type Config struct {
	RingBufferSize    int
	ChannelBufferSize int
}

// DefaultConfig returns default configuration
func DefaultConfig() *Config {
	return &Config{
		RingBufferSize:    DefaultRingBufferSize,
		ChannelBufferSize: DefaultChannelBufferSize,
	}
}

// Topic represents a pub/sub topic
type Topic struct {
	Name        string                 `json:"name"`
	Subscribers map[string]*Subscriber `json:"-"` // client_id -> subscriber
	Messages    *RingBuffer            `json:"-"` // Ring buffer for message replay
	CreatedAt   time.Time              `json:"created_at"`
	mu          sync.RWMutex           `json:"-"`
}

// Subscriber represents a WebSocket connection subscribed to a topic
type Subscriber struct {
	ClientID    string        `json:"client_id"`
	TopicName   string        `json:"topic_name"`
	MessageChan chan *Message `json:"-"` // Channel for sending messages
	LastSeen    time.Time     `json:"last_seen"`
}

// Message represents a published message
type Message struct {
	ID        string      `json:"id"`
	Payload   interface{} `json:"payload"`
	Topic     string      `json:"topic"`
	Timestamp time.Time   `json:"timestamp"`
}

// TopicInfo represents topic information for external APIs
type TopicInfo struct {
	Name        string `json:"name"`
	Subscribers int    `json:"subscribers"`
}

// HealthResponse represents health information
type HealthResponse struct {
	UptimeSec   int64 `json:"uptime_sec"`
	Topics      int   `json:"topics"`
	Subscribers int   `json:"subscribers"`
}

// TopicStats represents statistics for a topic
type TopicStats struct {
	Messages    int `json:"messages"`
	Subscribers int `json:"subscribers"`
}

// StatsResponse represents overall statistics
type StatsResponse struct {
	Topics map[string]TopicStats `json:"topics"`
}

// RingBuffer for message replay with drop-oldest backpressure policy
type RingBuffer struct {
	buffer []*Message
	size   int
	head   int
	tail   int
	count  int
	mu     sync.RWMutex
}

// NewRingBuffer creates a new ring buffer with specified size
func NewRingBuffer(size int) *RingBuffer {
	return &RingBuffer{
		buffer: make([]*Message, size),
		size:   size,
	}
}

// Add adds a message to the ring buffer (drop-oldest policy)
func (rb *RingBuffer) Add(msg *Message) {
	rb.mu.Lock()
	defer rb.mu.Unlock()

	rb.buffer[rb.tail] = msg
	rb.tail = (rb.tail + 1) % rb.size

	if rb.count < rb.size {
		rb.count++
	} else {
		// Drop oldest message (advance head)
		rb.head = (rb.head + 1) % rb.size
	}
}

// GetLastN returns the last n messages in chronological order
func (rb *RingBuffer) GetLastN(n int) []*Message {
	rb.mu.RLock()
	defer rb.mu.RUnlock()

	if n <= 0 || rb.count == 0 {
		return []*Message{}
	}

	if n > rb.count {
		n = rb.count
	}

	messages := make([]*Message, 0, n)

	// Start from the most recent message
	start := (rb.tail - 1 + rb.size) % rb.size

	for i := 0; i < n; i++ {
		idx := (start - i + rb.size) % rb.size
		if rb.buffer[idx] != nil {
			messages = append(messages, rb.buffer[idx])
		}
	}

	// Reverse to get chronological order (oldest first)
	for i, j := 0, len(messages)-1; i < j; i, j = i+1, j-1 {
		messages[i], messages[j] = messages[j], messages[i]
	}

	return messages
}

// Count returns the number of messages in the buffer
func (rb *RingBuffer) Count() int {
	rb.mu.RLock()
	defer rb.mu.RUnlock()
	return rb.count
}

// GetMessages returns all messages in the buffer (for stats)
func (rb *RingBuffer) GetMessages() []*Message {
	rb.mu.RLock()
	defer rb.mu.RUnlock()

	if rb.count == 0 {
		return []*Message{}
	}

	messages := make([]*Message, 0, rb.count)

	// Get messages in chronological order
	for i := 0; i < rb.count; i++ {
		idx := (rb.head + i) % rb.size
		if rb.buffer[idx] != nil {
			messages = append(messages, rb.buffer[idx])
		}
	}

	return messages
}
