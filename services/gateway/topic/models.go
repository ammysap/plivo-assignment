package topic

// REST API Models
type CreateTopicRequest struct {
	Name string `json:"name" binding:"required"`
}

type CreateTopicResponse struct {
	Status string `json:"status"`
	Topic  string `json:"topic"`
}

type DeleteTopicResponse struct {
	Status string `json:"status"`
	Topic  string `json:"topic"`
}

type TopicInfo struct {
	Name        string `json:"name"`
	Subscribers int    `json:"subscribers"`
}

type ListTopicsResponse struct {
	Topics []TopicInfo `json:"topics"`
}

type HealthResponse struct {
	UptimeSec   int64 `json:"uptime_sec"`
	Topics      int   `json:"topics"`
	Subscribers int   `json:"subscribers"`
}

type TopicStats struct {
	Messages    int `json:"messages"`
	Subscribers int `json:"subscribers"`
}

type StatsResponse struct {
	Topics map[string]TopicStats `json:"topics"`
}
