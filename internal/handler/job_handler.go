package handler

import (
	"encoding/json"
	"net/http"
	"strings"

	"task-scheduler/internal/service"
)

type JobHandler struct {
	service *service.JobService
}

func NewJobHandler(service *service.JobService) *JobHandler {
	return &JobHandler{
		service: service,
	}
}

func writeJSON(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(data)
}

func (h *JobHandler) HandleJobs(w http.ResponseWriter, r *http.Request) {

	switch r.Method {

	case http.MethodPost:
		h.createJob(w, r)

	case http.MethodGet:
		h.getAllJobs(w, r)

	default:
		writeJSON(w, http.StatusMethodNotAllowed, map[string]string{
			"error": "method not allowed",
		})
	}
}

func (h *JobHandler) createJob(w http.ResponseWriter, r *http.Request) {

	var request service.CreateJobRequest

	err := json.NewDecoder(r.Body).Decode(&request)

	if err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{
			"error": "invalid request body",
		})
		return
	}

	job, err := h.service.CreateJob(request)

	if err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{
			"error": err.Error(),
		})
		return
	}

	writeJSON(w, http.StatusCreated, job)
}

func (h *JobHandler) getAllJobs(w http.ResponseWriter, r *http.Request) {

	jobs, err := h.service.GetAllJobs()

	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{
			"error": err.Error(),
		})
		return
	}

	writeJSON(w, http.StatusOK, jobs)
}

func (h *JobHandler) HandleJobByID(w http.ResponseWriter, r *http.Request) {

	if r.Method != http.MethodGet {
		writeJSON(w, http.StatusMethodNotAllowed, map[string]string{
			"error": "method not allowed",
		})
		return
	}

	id := strings.TrimPrefix(r.URL.Path, "/jobs/")

	if id == "" {
		writeJSON(w, http.StatusBadRequest, map[string]string{
			"error": "job id is required",
		})
		return
	}

	job, err := h.service.GetJob(id)

	if err != nil {
		writeJSON(w, http.StatusNotFound, map[string]string{
			"error": err.Error(),
		})
		return
	}

	writeJSON(w, http.StatusOK, job)
}
