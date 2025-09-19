package topic

import (
	"context"

	"github.com/ammysap/plivo-pub-sub/pubsub"
)

// service implements the Service interface
type Service interface {
	CreateTopic(name string) error
	DeleteTopic(name string) error
	ListTopics() ([]TopicInfo, error)
	GetHealth() (HealthResponse, error)
	GetStats() (StatsResponse, error)
}
type service struct {
	pubsubService pubsub.Service
}

// NewService creates a new topic service
func NewService() Service {
	return &service{
		pubsubService: pubsub.GetService(),
	}
}

// CreateTopic creates a new topic
func (s *service) CreateTopic(name string) error {
	ctx := context.Background()
	return s.pubsubService.CreateTopic(ctx, name)
}

// DeleteTopic deletes a topic
func (s *service) DeleteTopic(name string) error {
	ctx := context.Background()
	return s.pubsubService.DeleteTopic(ctx, name)
}

// ListTopics returns all topics
func (s *service) ListTopics() ([]TopicInfo, error) {
	ctx := context.Background()
	pubsubTopics, err := s.pubsubService.ListTopics(ctx)
	if err != nil {
		return nil, err
	}

	// Convert pubsub.TopicInfo to local TopicInfo
	topics := make([]TopicInfo, len(pubsubTopics))
	for i, topic := range pubsubTopics {
		topics[i] = TopicInfo{
			Name:        topic.Name,
			Subscribers: topic.Subscribers,
		}
	}

	return topics, nil
}

// GetHealth returns service health
func (s *service) GetHealth() (HealthResponse, error) {
	ctx := context.Background()
	pubsubHealth, err := s.pubsubService.GetHealth(ctx)
	if err != nil {
		return HealthResponse{}, err
	}

	return HealthResponse{
		UptimeSec:   pubsubHealth.UptimeSec,
		Topics:      pubsubHealth.Topics,
		Subscribers: pubsubHealth.Subscribers,
	}, nil
}

// GetStats returns service statistics
func (s *service) GetStats() (StatsResponse, error) {
	ctx := context.Background()
	pubsubStats, err := s.pubsubService.GetStats(ctx)
	if err != nil {
		return StatsResponse{}, err
	}

	// Convert pubsub.StatsResponse to local StatsResponse
	stats := StatsResponse{
		Topics: make(map[string]TopicStats),
	}

	for name, topicStats := range pubsubStats.Topics {
		stats.Topics[name] = TopicStats{
			Messages:    topicStats.Messages,
			Subscribers: topicStats.Subscribers,
		}
	}

	return stats, nil
}
