package service

import (
	"crypto/rand"
	"encoding/hex"
	"errors"
	"time"

	"task-scheduler/internal/model"
	"task-scheduler/internal/repository"
)

type JobService struct {
	repository *repository.JobRepository
}

func NewJobService(repository *repository.JobRepository) *JobService {
	return &JobService{
		repository: repository,
	}
}

type CreateJobRequest struct {
	Type       string                 `json:"type"`
	Payload    map[string]interface{} `json:"payload"`
	Priority   int                    `json:"priority"`
	MaxRetries int                    `json:"max_retries"`
}

func generateID() (string, error) {
	bytes := make([]byte, 8)

	_, err := rand.Read(bytes)
	if err != nil {
		return "", err
	}

	return hex.EncodeToString(bytes), nil
}

func (s *JobService) CreateJob(request CreateJobRequest) (*model.Job, error) {

	if request.Type == "" {
		return nil, errors.New("job type is required")
	}

	if request.Priority < 0 {
		return nil, errors.New("priority cannot be negative")
	}

	if request.MaxRetries <= 0 {
		request.MaxRetries = 3
	}

	id, err := generateID()
	if err != nil {
		return nil, err
	}

	now := time.Now()

	job := &model.Job{
		ID:         id,
		Type:       request.Type,
		Payload:    request.Payload,
		Status:     model.JobQueued,
		Priority:   request.Priority,
		RetryCount: 0,
		MaxRetries: request.MaxRetries,
		CreatedAt:  now,
		UpdatedAt:  now,
	}

	err = s.repository.Create(job)
	if err != nil {
		return nil, err
	}

	return job, nil
}

func (s *JobService) GetJob(id string) (*model.Job, error) {
	return s.repository.GetByID(id)
}

func (s *JobService) GetAllJobs() ([]*model.Job, error) {
	return s.repository.GetAll()
}
