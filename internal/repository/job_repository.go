package repository

import (
	"context"
	"encoding/json"
	"errors"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"task-scheduler/internal/model"
)

type JobRepository struct {
	db *pgxpool.Pool
}

func NewJobRepository(db *pgxpool.Pool) *JobRepository {
	return &JobRepository{
		db: db,
	}
}

func (r *JobRepository) Create(job *model.Job) error {

	payloadJSON, err := json.Marshal(job.Payload)
	if err != nil {
		return err
	}

	query := `
		INSERT INTO jobs (
			id,
			job_type,
			payload,
			status,
			priority,
			retry_count,
			max_retries,
			created_at,
			updated_at
		)
		VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9)
	`

	_, err = r.db.Exec(
		context.Background(),
		query,
		job.ID,
		job.Type,
		payloadJSON,
		job.Status,
		job.Priority,
		job.RetryCount,
		job.MaxRetries,
		job.CreatedAt,
		job.UpdatedAt,
	)

	return err
}

func (r *JobRepository) GetByID(id string) (*model.Job, error) {

	query := `
		SELECT
			id,
			job_type,
			payload,
			status,
			priority,
			retry_count,
			max_retries,
			created_at,
			updated_at
		FROM jobs
		WHERE id = $1
	`

	var job model.Job
	var payloadBytes []byte

	err := r.db.QueryRow(
		context.Background(),
		query,
		id,
	).Scan(
		&job.ID,
		&job.Type,
		&payloadBytes,
		&job.Status,
		&job.Priority,
		&job.RetryCount,
		&job.MaxRetries,
		&job.CreatedAt,
		&job.UpdatedAt,
	)

	if errors.Is(err, pgx.ErrNoRows) {
		return nil, errors.New("job not found")
	}

	if err != nil {
		return nil, err
	}

	err = json.Unmarshal(payloadBytes, &job.Payload)
	if err != nil {
		return nil, err
	}

	return &job, nil
}

func (r *JobRepository) GetAll() ([]*model.Job, error) {

	query := `
		SELECT
			id,
			job_type,
			payload,
			status,
			priority,
			retry_count,
			max_retries,
			created_at,
			updated_at
		FROM jobs
		ORDER BY created_at DESC
	`

	rows, err := r.db.Query(
		context.Background(),
		query,
	)

	if err != nil {
		return nil, err
	}

	defer rows.Close()

	var jobs []*model.Job

	for rows.Next() {

		var job model.Job
		var payloadBytes []byte

		err := rows.Scan(
			&job.ID,
			&job.Type,
			&payloadBytes,
			&job.Status,
			&job.Priority,
			&job.RetryCount,
			&job.MaxRetries,
			&job.CreatedAt,
			&job.UpdatedAt,
		)

		if err != nil {
			return nil, err
		}

		err = json.Unmarshal(payloadBytes, &job.Payload)
		if err != nil {
			return nil, err
		}

		jobs = append(jobs, &job)
	}

	return jobs, nil
}
