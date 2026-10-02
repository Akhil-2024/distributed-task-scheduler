package main

import (
	"context"
	"errors"
	"fmt"
	"log"
	"strings"
	"time"

	"task-scheduler/internal/database"
)

func processJob(jobType string) error {

	// Simulate actual work
	time.Sleep(3 * time.Second)

	// For testing failure/retry logic
	if strings.Contains(jobType, "fail") {
		return errors.New("simulated job failure")
	}

	return nil
}

func main() {

	dbPool, err := database.NewPostgresPool()
	if err != nil {
		log.Fatal("Failed to connect to PostgreSQL:", err)
	}
	defer dbPool.Close()

	fmt.Println("Worker connected to PostgreSQL")
	fmt.Println("Worker RETRY VERSION started...")

	for {

		ctx := context.Background()

		tx, err := dbPool.Begin(ctx)
		if err != nil {
			fmt.Println("Failed to start transaction:", err)
			time.Sleep(2 * time.Second)
			continue
		}

		var jobID string
		var jobType string
		var retryCount int
		var maxRetries int

		err = tx.QueryRow(
			ctx,
			`
			SELECT
				id,
				job_type,
				retry_count,
				max_retries
			FROM jobs
			WHERE status = 'QUEUED'
			  AND (
				  next_retry_at IS NULL
				  OR next_retry_at <= NOW()
			  )
			ORDER BY priority DESC, created_at ASC
			FOR UPDATE SKIP LOCKED
			LIMIT 1
			`,
		).Scan(
			&jobID,
			&jobType,
			&retryCount,
			&maxRetries,
		)

		if err != nil {
			tx.Rollback(ctx)
			time.Sleep(2 * time.Second)
			continue
		}

		_, err = tx.Exec(
			ctx,
			`
			UPDATE jobs
			SET status = 'PROCESSING',
			    updated_at = NOW()
			WHERE id = $1
			`,
			jobID,
		)

		if err != nil {
			tx.Rollback(ctx)

			fmt.Println("Failed to claim job:", err)

			time.Sleep(2 * time.Second)
			continue
		}

		err = tx.Commit(ctx)
		if err != nil {
			fmt.Println("Failed to commit transaction:", err)

			time.Sleep(2 * time.Second)
			continue
		}

		fmt.Println()
		fmt.Println("Picked job:", jobID)
		fmt.Println("Type:", jobType)
		fmt.Println("Retry:", retryCount, "/", maxRetries)

		err = processJob(jobType)

		// -----------------------------
		// SUCCESS CASE
		// -----------------------------
		if err == nil {

			_, dbErr := dbPool.Exec(
				ctx,
				`
				UPDATE jobs
				SET status = 'SUCCESS',
				    last_error = NULL,
				    next_retry_at = NULL,
				    updated_at = NOW()
				WHERE id = $1
				`,
				jobID,
			)

			if dbErr != nil {
				fmt.Println("Failed to mark success:", dbErr)
				continue
			}

			fmt.Println("Job completed successfully:", jobID)

			continue
		}

		// -----------------------------
		// FAILURE CASE
		// -----------------------------
		fmt.Println("Job failed:", err)

		newRetryCount := retryCount + 1

		// -----------------------------
		// MAX RETRIES REACHED
		// -----------------------------
		if newRetryCount >= maxRetries {

			_, dbErr := dbPool.Exec(
				ctx,
				`
				UPDATE jobs
				SET status = 'DEAD',
				    retry_count = $2,
				    last_error = $3,
				    next_retry_at = NULL,
				    updated_at = NOW()
				WHERE id = $1
				`,
				jobID,
				newRetryCount,
				err.Error(),
			)

			if dbErr != nil {
				fmt.Println("Failed to mark job DEAD:", dbErr)
				continue
			}

			fmt.Println("Max retries reached")
			fmt.Println("Job moved to DEAD:", jobID)

			continue
		}

		// -----------------------------
		// RETRY WITH EXPONENTIAL BACKOFF
		// -----------------------------
		delaySeconds := 1 << newRetryCount

		nextRetry := time.Now().Add(
			time.Duration(delaySeconds) * time.Second,
		)

		_, dbErr := dbPool.Exec(
			ctx,
			`
			UPDATE jobs
			SET status = 'QUEUED',
			    retry_count = $2,
			    last_error = $3,
			    next_retry_at = $4,
			    updated_at = NOW()
			WHERE id = $1
			`,
			jobID,
			newRetryCount,
			err.Error(),
			nextRetry,
		)

		if dbErr != nil {
			fmt.Println("Failed to schedule retry:", dbErr)
			continue
		}

		fmt.Println("Retry scheduled after", delaySeconds, "seconds")
	}
}
