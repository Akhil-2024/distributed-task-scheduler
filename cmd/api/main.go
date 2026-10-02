package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"

	"task-scheduler/internal/database"
	"task-scheduler/internal/handler"
	"task-scheduler/internal/repository"
	"task-scheduler/internal/service"
)

func main() {

	// Connect to PostgreSQL
	dbPool, err := database.NewPostgresPool()

	if err != nil {
		log.Fatal("Failed to connect to PostgreSQL:", err)
	}

	defer dbPool.Close()

	fmt.Println("Connected to PostgreSQL successfully")

	// Repository now uses PostgreSQL
	jobRepository := repository.NewJobRepository(dbPool)

	// Service layer
	jobService := service.NewJobService(jobRepository)

	// Handler layer
	jobHandler := handler.NewJobHandler(jobService)

	// Routes
	http.HandleFunc("/jobs", jobHandler.HandleJobs)

	http.HandleFunc("/jobs/", jobHandler.HandleJobByID)

	http.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {

		w.Header().Set("Content-Type", "application/json")

		json.NewEncoder(w).Encode(map[string]string{
			"status": "UP",
		})
	})

	fmt.Println("Distributed Task Scheduler")
	fmt.Println("Server running on http://localhost:8080")

	log.Fatal(http.ListenAndServe(":8080", nil))
}
