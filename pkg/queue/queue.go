package queue

import (
	"fmt"
	"math/rand"
	"sync"
	"time"
)

type TaskStatus string

const (
	StatusQueued  TaskStatus = "queued"
	StatusRunning TaskStatus = "running"
	StatusDone    TaskStatus = "done"
	StatusFailed  TaskStatus = "failed"
)

// структура задачи
type Task struct {
	ID         string     `json:"id"`
	Payload    string     `json:"payload"`
	MaxRetries int        `json:"max_retriers"`
	NextRunAt  int64      `json:"next_run_at"`
	Status     TaskStatus `json:"status"`
	Attempts   int        `json:"attempts"`
}

// стуркутра очереди задач
type TaskQueue struct {
	workers int
	queue   chan *Task
	delayed []*Task

	tasks  map[string]*Task
	mu     sync.Mutex
	wg     sync.WaitGroup
	stopCh chan struct{}
}

// конструктор очереди
func NewTaskQueue(workers, queueSize int) *TaskQueue {
	return &TaskQueue{
		workers: workers,
		queue:   make(chan *Task, queueSize),
		tasks:   make(map[string]*Task),
		delayed: make([]*Task, 0),
		stopCh:  make(chan struct{}),
	}
}

// добавление задачи в очередь
func (q *TaskQueue) Enqueue(task *Task) error {
	select {

	case q.queue <- task:
		q.mu.Lock()
		defer q.mu.Unlock()

		q.tasks[task.ID] = task
		task.Status = StatusQueued

		return nil

	default:
		return fmt.Errorf("queue is full")
	}
}

// запуск воркеров
func (q *TaskQueue) Start() {

	for i := 0; i < q.workers; i++ {
		q.wg.Add(1)
		go q.worker(i)
	}

	go q.runDelayedChecker()
}

// воркеры(горутины)
func (q *TaskQueue) worker(id int) {
	defer q.wg.Done()

	for {
		select {
		case task, ok := <-q.queue:
			if !ok {

				fmt.Printf("Worker %d завершил работу\n", id)
				return
			}
			if task == nil {
				continue
			}

			now := time.Now().UnixMilli()
			if task.NextRunAt > now {
				q.mu.Lock()
				q.delayed = append(q.delayed, task)
				q.mu.Unlock()
				continue
			}
			q.processTask(task)

		case <-q.stopCh:
			close(q.queue)
		}
	}
}

// имитация выполнения
func (q *TaskQueue) processTask(task *Task) {
	q.mu.Lock()
	task.Status = StatusRunning
	q.mu.Unlock()

	delay := time.Duration(100+rand.Intn(400)) * time.Millisecond
	time.Sleep(delay)

	if rand.Intn(100) < 20 {
		task.Attempts++
		if task.Attempts <= task.MaxRetries {
			backoff := calcBackoff(task.Attempts)
			task.NextRunAt = time.Now().Add(backoff).UnixMilli()

			q.mu.Lock()
			q.delayed = append(q.delayed, task)
			task.Status = StatusQueued
			q.mu.Unlock()

			fmt.Printf("Task %s провалена, retry #%d через %v\n", task.ID, task.Attempts, backoff)
		} else {
			q.mu.Lock()
			task.Status = StatusFailed
			q.mu.Unlock()

			fmt.Printf("Task %s окончательно упала после %d попыток\n", task.ID, task.Attempts)
		}
		return
	}

	// Успешное выполнение
	q.mu.Lock()
	task.Status = StatusDone
	q.mu.Unlock()

	fmt.Printf("Task %s успешно выполнена\n", task.ID)
}

// вычисление времени ожидания
func calcBackoff(attempt int) time.Duration {
	base := 100 * time.Millisecond
	max := 5 * time.Second
	backoff := base * (1 << attempt)
	if backoff > max {
		backoff = max
	}

	jitter := time.Duration(rand.Int63n(int64(backoff)))
	return backoff/2 + jitter
}

// проверка подочереди
func (q *TaskQueue) runDelayedChecker() {
	ticker := time.NewTicker(100 * time.Millisecond)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			now := time.Now().UnixMilli()

			q.mu.Lock()
			var ready []*Task
			var waiting []*Task

			for _, task := range q.delayed {
				if task.NextRunAt <= now {
					ready = append(ready, task)
				} else {
					waiting = append(waiting, task)
				}
			}

			q.delayed = waiting
			q.mu.Unlock()

			for _, task := range ready {
				select {
				case q.queue <- task:
					fmt.Printf("Task %s вернулась в очередь на retry\n", task.ID)
				default:
					fmt.Printf("Очередь переполнена, задача %s осталась в delayed\n", task.ID)
					q.mu.Lock()
					q.delayed = append(q.delayed, task)
					q.mu.Unlock()
				}
			}

		case <-q.stopCh:
			return
		}
	}
}

// остановка воркеров
func (q *TaskQueue) Stop() {

	close(q.stopCh)
	q.wg.Wait()

}
