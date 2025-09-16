package main

import (
	"WebPullWorker/config"
	"WebPullWorker/pkg/handlers"
	"WebPullWorker/pkg/queue"
	"log"
	"net/http"

	"fmt"
)

func main() {

	cfg := config.LoadConfig()
	fmt.Printf("Текущее значение Workers: %d, и QueueSize: %d\n", cfg.CountWorkers, cfg.QueueSize)

	taskQueue := queue.NewTaskQueue(cfg.CountWorkers, cfg.QueueSize)
	taskQueue.Start()
	defer taskQueue.Stop()

	http.HandleFunc("/enqueue", handlers.EnqueueHandler(taskQueue))
	http.HandleFunc("/healthz", handlers.HealthHandler)

	fmt.Println("Server started on :8080")
	if err := http.ListenAndServe(":8080", nil); err != nil {
		log.Fatal(err)
	}
}
