package model

import "time"

type DLQMessage struct {
	TaskID         string              `json:"task_id"`
	OriginalTask   ImageProcessingTask `json:"original_task"`
	Error          string              `json:"error"`
	PartialResults []string            `json:"partial_results,omitempty"`
	Timestamp      time.Time           `json:"timestamp"`
	RetryCount     int                 `json:"retry_count"`
}
