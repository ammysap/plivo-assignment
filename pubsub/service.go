package pubsub

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/ammysap/plivo-pub-sub/logging"
	"github.com/google/uuid"
)

// Service interface for external access
type Service interface {
	CreateTopic(ctx context.Context, name string) error
	DeleteTopic(ctx context.Context, name string) error
	GetTopic(ctx context.Context, name string) (*Topic, error)
	ListTopics(ctx context.Context) ([]TopicInfo, error)
	Subscribe(ctx context.Context, topicName, clientID string, lastN int) (*Subscriber, error)
	Unsubscribe(ctx context.Context, topicName, clientID string) error
	Publish(ctx context.Context, topicName string, message *Message) error
	GetStats(ctx context.Context) (*StatsResponse, error)
	GetHealth(ctx context.Context) (*HealthResponse, error)
	Start(ctx context.Context) error
	Stop(ctx context.Context) error
}

// Singleton instance
var (
	instance *service
	once     sync.Once
)

// service implements the PubSub service with singleton pattern
type service struct {
	topics    map[string]*Topic
	config    *Config
	startTime time.Time
	mu        sync.RWMutex
	shutdown  chan struct{}
	wg        sync.WaitGroup
}

// InitService initializes the singleton PubSub service
func InitService(config *Config) *service {
	once.Do(func() {
		if config == nil {
			config = DefaultConfig()
		}

		instance = &service{
			topics:   make(map[string]*Topic),
			config:   config,
			shutdown: make(chan struct{}),
		}
	})
	return instance
}

// GetService returns the singleton instance
func GetService() *service {
	if instance == nil {
		panic("PubSub service not initialized. Call InitService() first.")
	}
	return instance
}

// Start initializes the service
func (s *service) Start(ctx context.Context) error {
	s.startTime = time.Now()
	log := logging.WithContext(ctx)
	log.Info("PubSub service started")
	return nil
}

// Stop gracefully shuts down the service
func (s *service) Stop(ctx context.Context) error {
	log := logging.WithContext(ctx)
	log.Info("Stopping PubSub service...")

	// Signal shutdown
	close(s.shutdown)

	// Wait for graceful shutdown with timeout
	done := make(chan struct{})
	go func() {
		s.wg.Wait()
		close(done)
	}()

	select {
	case <-done:
		log.Info("PubSub service stopped gracefully")
	case <-time.After(GracefulShutdownTimeout):
		log.Warn("PubSub service shutdown timeout exceeded")
	}

	return nil
}

// CreateTopic creates a new topic
func (s *service) CreateTopic(ctx context.Context, name string) error {
	log := logging.WithContext(ctx)

	s.mu.Lock()
	defer s.mu.Unlock()

	if _, exists := s.topics[name]; exists {
		return fmt.Errorf("topic %s already exists", name)
	}

	topic := &Topic{
		Name:        name,
		Subscribers: make(map[string]*Subscriber),
		Messages:    NewRingBuffer(s.config.RingBufferSize),
		CreatedAt:   time.Now(),
	}

	s.topics[name] = topic
	log.Info("Created topic", "topic", name)

	return nil
}

// DeleteTopic deletes a topic and disconnects all subscribers
func (s *service) DeleteTopic(ctx context.Context, name string) error {
	log := logging.WithContext(ctx)

	s.mu.Lock()
	defer s.mu.Unlock()

	topic, exists := s.topics[name]
	if !exists {
		return fmt.Errorf("topic %s not found", name)
	}

	// Disconnect all subscribers
	topic.mu.Lock()
	for clientID, subscriber := range topic.Subscribers {
		close(subscriber.MessageChan)
		log.Info("Disconnected subscriber", "topic", name, "client_id", clientID)
	}
	topic.mu.Unlock()

	delete(s.topics, name)
	log.Info("Deleted topic", "topic", name)

	return nil
}

// GetTopic retrieves a topic by name
func (s *service) GetTopic(ctx context.Context, name string) (*Topic, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	topic, exists := s.topics[name]
	if !exists {
		return nil, fmt.Errorf("topic %s not found", name)
	}

	return topic, nil
}

// ListTopics returns all topics with subscriber counts
func (s *service) ListTopics(ctx context.Context) ([]TopicInfo, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	topics := make([]TopicInfo, 0, len(s.topics))
	for name, topic := range s.topics {
		topic.mu.RLock()
		subscriberCount := len(topic.Subscribers)
		topic.mu.RUnlock()

		topics = append(topics, TopicInfo{
			Name:        name,
			Subscribers: subscriberCount,
		})
	}

	return topics, nil
}

// Subscribe adds a client to a topic
func (s *service) Subscribe(ctx context.Context, topicName, clientID string, lastN int) (*Subscriber, error) {
	log := logging.WithContext(ctx)

	s.mu.RLock()
	topic, exists := s.topics[topicName]
	s.mu.RUnlock()

	if !exists {
		return nil, fmt.Errorf("topic %s not found", topicName)
	}

	topic.mu.Lock()
	defer topic.mu.Unlock()

	// Check if already subscribed
	if _, exists := topic.Subscribers[clientID]; exists {
		return nil, fmt.Errorf("client %s already subscribed to topic %s", clientID, topicName)
	}

	// Create subscriber with buffered channel
	subscriber := &Subscriber{
		ClientID:    clientID,
		TopicName:   topicName,
		MessageChan: make(chan *Message, s.config.ChannelBufferSize),
		LastSeen:    time.Now(),
	}

	topic.Subscribers[clientID] = subscriber

	// Send historical messages if requested
	if lastN > 0 {
		historicalMessages := topic.Messages.GetLastN(lastN)
		go func() {
			for _, msg := range historicalMessages {
				select {
				case subscriber.MessageChan <- msg:
				case <-s.shutdown:
					return
				default:
					// Channel is full, drop message (backpressure)
					log.Warn("Dropped historical message due to full channel",
						"client_id", clientID, "topic", topicName)
				}
			}
		}()
	}

	log.Info("Subscribed client to topic", "client_id", clientID, "topic", topicName, "last_n", lastN)
	return subscriber, nil
}

// Unsubscribe removes a client from a topic
func (s *service) Unsubscribe(ctx context.Context, topicName, clientID string) error {
	log := logging.WithContext(ctx)

	s.mu.RLock()
	topic, exists := s.topics[topicName]
	s.mu.RUnlock()

	if !exists {
		return fmt.Errorf("topic %s not found", topicName)
	}

	topic.mu.Lock()
	defer topic.mu.Unlock()

	subscriber, exists := topic.Subscribers[clientID]
	if !exists {
		return fmt.Errorf("client %s not subscribed to topic %s", clientID, topicName)
	}

	// Close the message channel
	close(subscriber.MessageChan)
	delete(topic.Subscribers, clientID)

	log.Info("Unsubscribed client from topic", "client_id", clientID, "topic", topicName)
	return nil
}

// Publish sends a message to all subscribers of a topic
func (s *service) Publish(ctx context.Context, topicName string, message *Message) error {
	log := logging.WithContext(ctx)

	s.mu.RLock()
	topic, exists := s.topics[topicName]
	s.mu.RUnlock()

	if !exists {
		return fmt.Errorf("topic %s not found", topicName)
	}

	// Set message metadata
	message.Topic = topicName
	message.Timestamp = time.Now()
	if message.ID == "" {
		message.ID = uuid.New().String()
	}

	// Add to ring buffer for replay
	topic.Messages.Add(message)

	// Fan-out to all subscribers
	topic.mu.RLock()
	subscribers := make([]*Subscriber, 0, len(topic.Subscribers))
	for _, subscriber := range topic.Subscribers {
		subscribers = append(subscribers, subscriber)
	}
	topic.mu.RUnlock()

	// Send message to all subscribers concurrently
	for _, subscriber := range subscribers {
		go func(sub *Subscriber) {
			select {
			case sub.MessageChan <- message:
				// Message sent successfully
			case <-s.shutdown:
				// Service is shutting down
				return
			default:
				// Channel is full, drop message (backpressure policy)
				log.Warn("Dropped message due to full subscriber channel",
					"client_id", sub.ClientID, "topic", topicName)
			}
		}(subscriber)
	}

	log.Info("Published message to topic", "topic", topicName, "message_id", message.ID, "subscribers", len(subscribers))
	return nil
}

// GetStats returns detailed statistics
func (s *service) GetStats(ctx context.Context) (*StatsResponse, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	stats := &StatsResponse{
		Topics: make(map[string]TopicStats),
	}

	for name, topic := range s.topics {
		topic.mu.RLock()
		subscriberCount := len(topic.Subscribers)
		messageCount := topic.Messages.Count()
		topic.mu.RUnlock()

		stats.Topics[name] = TopicStats{
			Messages:    messageCount,
			Subscribers: subscriberCount,
		}
	}

	return stats, nil
}

// GetHealth returns service health information
func (s *service) GetHealth(ctx context.Context) (*HealthResponse, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	totalSubscribers := 0
	for _, topic := range s.topics {
		topic.mu.RLock()
		totalSubscribers += len(topic.Subscribers)
		topic.mu.RUnlock()
	}

	return &HealthResponse{
		UptimeSec:   int64(time.Since(s.startTime).Seconds()),
		Topics:      len(s.topics),
		Subscribers: totalSubscribers,
	}, nil
}
