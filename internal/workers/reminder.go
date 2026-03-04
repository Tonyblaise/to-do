package workers

import (
	"context"
	"database/sql"
	"log/slog"
	"sync"
	"time"

	"github.com/Tonyblaise/to-do/internal/repository"
)

type ReminderWorker struct {
	taskRepo *repository.TaskRepository
	logger   *slog.Logger
	jobs     chan string
}

func NewReminderWorker(db *sql.DB, logger *slog.Logger) *ReminderWorker {
	return &ReminderWorker{
		taskRepo: repository.NewTaskRepository(db),
		logger:   logger,
		jobs:     make(chan string, 100),
	}
}

func (w *ReminderWorker) Start(ctx context.Context) {
	const numWorkers = 3
	const pollInterval = 1 * time.Minute
	const lookaheadWindow = 24 * time.Hour

	var wg sync.WaitGroup
	
	for i := 0; i < numWorkers; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			w.processJobs(ctx, id)
		}(i)
	}

	ticker := time.NewTicker(pollInterval)
	defer ticker.Stop()

	w.logger.Info("reminder worker started", "workers", numWorkers, "poll_interval", pollInterval)

	for {
		select {
		case <-ctx.Done():
			w.logger.Info("reminder worker stopping")
			close(w.jobs)
			wg.Wait()
			return
		case <-ticker.C:
			w.poll(lookaheadWindow)
		}
	}
}

func (w *ReminderWorker) poll(window time.Duration) {
	tasks, err := w.taskRepo.GetUpcomingDue(window)
	if err != nil {
		w.logger.Error("reminder worker poll failed", "error", err)
		return
	}

	w.logger.Info("reminder worker polled", "tasks_found", len(tasks))

	for _, t := range tasks {
		select {
		case w.jobs <- t.ID:
		default:
			w.logger.Warn("reminder job queue full, skipping task", "task_id", t.ID)
		}
	}
}

func (w *ReminderWorker) processJobs(ctx context.Context, workerID int) {
	for {
		select {
		case taskID, ok := <-w.jobs:
			if !ok {
				return
			}
			w.sendReminder(workerID, taskID)
		case <-ctx.Done():
			return
		}
	}
}

func (w *ReminderWorker) sendReminder(workerID int, taskID string) {
	w.logger.Info("sending reminder (mocked)",
		"worker_id", workerID,
		"task_id", taskID,
	)

	time.Sleep(10 * time.Millisecond)
}
