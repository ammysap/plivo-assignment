package topic

import (
	"net/http"

	"github.com/ammysap/plivo-pub-sub/services/gateway/logger"
	"github.com/gin-gonic/gin"
)

// endpoint implements the Endpoint interface
type Endpoint interface {
	CreateTopic(c *gin.Context)
	DeleteTopic(c *gin.Context)
	ListTopics(c *gin.Context)
	GetHealth(c *gin.Context)
	GetStats(c *gin.Context)
}
type endpoint struct {
	service Service
}

// NewEndpoint creates a new endpoint
func NewEndpoint(service Service) Endpoint {
	return &endpoint{
		service: service,
	}
}

// CreateTopic handles POST /topics
func (e *endpoint) CreateTopic(c *gin.Context) {
	_, log, err := logger.GetLoggerFromGinContext(c)
	if err != nil {
		log.Errorw("Error getting logger from gin context", "error", err.Error())
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	var req CreateTopicRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		log.Errorw("Invalid request body", "error", err.Error())
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body"})
		return
	}

	if req.Name == "" {
		log.Errorw("Topic name is required")
		c.JSON(http.StatusBadRequest, gin.H{"error": "Topic name is required"})
		return
	}

	err = e.service.CreateTopic(req.Name)
	if err != nil {
		if err.Error() == "topic "+req.Name+" already exists" {
			log.Errorw("Topic already exists", "topic", req.Name)
			c.JSON(http.StatusConflict, gin.H{"error": "Topic already exists"})
			return
		}
		log.Errorw("Error creating topic", "error", err.Error(), "topic", req.Name)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create topic"})
		return
	}

	response := CreateTopicResponse{
		Status: "created",
		Topic:  req.Name,
	}

	log.Infow("Topic created successfully", "topic", req.Name)
	c.JSON(http.StatusCreated, response)
}

// DeleteTopic handles DELETE /topics/{name}
func (e *endpoint) DeleteTopic(c *gin.Context) {
	_, log, err := logger.GetLoggerFromGinContext(c)
	if err != nil {
		log.Errorw("Error getting logger from gin context", "error", err.Error())
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	topicName := c.Param("name")
	if topicName == "" {
		log.Errorw("Topic name is required")
		c.JSON(http.StatusBadRequest, gin.H{"error": "Topic name is required"})
		return
	}

	err = e.service.DeleteTopic(topicName)
	if err != nil {
		if err.Error() == "topic "+topicName+" not found" {
			log.Warnw("Topic not found", "topic", topicName)
			c.JSON(http.StatusNotFound, gin.H{"error": "Topic not found"})
			return
		}
		log.Errorw("Error deleting topic", "error", err.Error(), "topic", topicName)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete topic"})
		return
	}

	response := DeleteTopicResponse{
		Status: "deleted",
		Topic:  topicName,
	}

	log.Infow("Topic deleted successfully", "topic", topicName)
	c.JSON(http.StatusOK, response)
}

// ListTopics handles GET /topics
func (e *endpoint) ListTopics(c *gin.Context) {
	_, log, err := logger.GetLoggerFromGinContext(c)
	if err != nil {
		log.Errorw("Error getting logger from gin context", "error", err.Error())
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	topics, err := e.service.ListTopics()
	if err != nil {
		log.Errorw("Error listing topics", "error", err.Error())
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to list topics"})
		return
	}

	response := ListTopicsResponse{
		Topics: topics,
	}

	log.Infow("Topics listed successfully", "count", len(topics))
	c.JSON(http.StatusOK, response)
}

// GetHealth handles GET /health
func (e *endpoint) GetHealth(c *gin.Context) {
	_, log, err := logger.GetLoggerFromGinContext(c)
	if err != nil {
		log.Errorw("Error getting logger from gin context", "error", err.Error())
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	health, err := e.service.GetHealth()
	if err != nil {
		log.Errorw("Error getting health status", "error", err.Error())
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get health status"})
		return
	}

	log.Debugw("Health check requested", "uptime", health.UptimeSec, "topics", health.Topics, "subscribers", health.Subscribers)
	c.JSON(http.StatusOK, health)
}

// GetStats handles GET /stats
func (e *endpoint) GetStats(c *gin.Context) {
	_, log, err := logger.GetLoggerFromGinContext(c)
	if err != nil {
		log.Errorw("Error getting logger from gin context", "error", err.Error())
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	stats, err := e.service.GetStats()
	if err != nil {
		log.Errorw("Error getting stats", "error", err.Error())
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get stats"})
		return
	}

	log.Debugw("Stats requested", "topics_count", len(stats.Topics))
	c.JSON(http.StatusOK, stats)
}
