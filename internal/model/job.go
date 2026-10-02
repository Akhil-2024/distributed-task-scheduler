package model

import "time"

type JobStatus string

const (
	JobQueued     JobStatus = "QUEUED"
	JobProcessing JobStatus = "PROCESSING"
	JobSuccess    JobStatus = "SUCCESS"
	JobFailed     JobStatus = "FAILED"
)

type Job struct {
	ID         string                 `json:"id"`
	Type       string                 `json:"type"`
	Payload    map[string]interface{} `json:"payload"`
	Status     JobStatus              `json:"status"`
	Priority   int                    `json:"priority"`
	RetryCount int                    `json:"retry_count"`
	MaxRetries int                    `json:"max_retries"`
	CreatedAt  time.Time              `json:"created_at"`
	UpdatedAt  time.Time              `json:"updated_at"`
}
