package handlers

import (
	"WebPullWorker/pkg/queue"
	"encoding/json"
	"net/http"
	"time"
)

type EnqueueRequest struct {
	ID         string `json:"id"`
	Payload    string `json:"payload"`
	MaxRetries int    `json:"max_retries"`
}

func EnqueueHandler(q *queue.TaskQueue) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {

		if r.Method != http.MethodPost {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}

		var req EnqueueRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, "invalid json", http.StatusBadRequest)
			return
		}
		if req.ID == "" {
			http.Error(w, "id is required", http.StatusBadRequest)
			return
		}
		if req.MaxRetries < 0 {
			req.MaxRetries = 0
		}

		task := &queue.Task{
			ID:         req.ID,
			Payload:    req.Payload,
			MaxRetries: req.MaxRetries,
			Attempts:   0,
			Status:     queue.StatusQueued,
			NextRunAt:  time.Now().UnixMilli(),
		}

		if err := q.Enqueue(task); err != nil {
			http.Error(w, "queue is full", http.StatusServiceUnavailable)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusAccepted)
		json.NewEncoder(w).Encode(map[string]string{
			"status": "enqueued",
			"id":     task.ID,
		})
	}
}

func HealthHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Write([]byte("OK"))
}
