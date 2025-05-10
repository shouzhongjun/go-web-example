package scheduler

import (
	"context"
	"errors"
	"fmt"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/robfig/cron/v3"
	"go.uber.org/zap"
	repo "goWebExample/internal/repository/scheduler"
)

const ServiceName = "scheduler"

// Task represents a scheduled task
type Task struct {
	ID          string
	Description string
	Schedule    string
	Interval    time.Duration
	Func        func() error
	IsRunning   bool
	LastRun     time.Time
	NextRun     time.Time
	Error       error
	Status      repo.TaskStatus
	entryID     cron.EntryID
}

// SchedulerService interface defines the methods for a scheduler service
type SchedulerService interface {
	// AddTask adds a new task to the scheduler with a duration interval
	AddTask(id, description string, interval time.Duration, taskFunc func() error) (string, error)
	// AddTaskWithSchedule adds a new task with a schedule string (e.g. "5s", "1m", "2h")
	AddTaskWithSchedule(id, description, schedule string, taskFunc func() error) (string, error)
	// RemoveTask removes a task from the scheduler
	RemoveTask(id string) error
	// GetTasks returns all registered tasks
	GetTasks() []*Task
	// GetTask returns a specific task by ID
	GetTask(id string) (*Task, error)
	// Start starts the scheduler
	Start() error
	// Stop stops the scheduler
	Stop() error
}

// schedulerService implements the SchedulerService interface
type schedulerService struct {
	tasks      map[string]*Task
	taskMutex  sync.RWMutex
	logger     *zap.Logger
	ctx        context.Context
	cancel     context.CancelFunc
	running    bool
	cron       *cron.Cron
	repository repo.TaskRepository
	cache      repo.TaskCache
}

// NewSchedulerService creates a new scheduler service
func NewSchedulerService(logger *zap.Logger, repository repo.TaskRepository, cache repo.TaskCache) SchedulerService {
	ctx, cancel := context.WithCancel(context.Background())

	return &schedulerService{
		tasks:      make(map[string]*Task),
		taskMutex:  sync.RWMutex{},
		logger:     logger,
		ctx:        ctx,
		cancel:     cancel,
		running:    false,
		cron:       cron.New(cron.WithSeconds()),
		repository: repository,
		cache:      cache,
	}
}

// durationToCronExpression converts a time.Duration to a cron expression
func durationToCronExpression(d time.Duration) string {
	seconds := int(d.Seconds())
	if seconds < 60 {
		// For durations less than a minute, run every N seconds
		return "@every " + d.String()
	} else if seconds < 3600 {
		// For durations less than an hour, run every N minutes
		return "@every " + d.String()
	} else {
		// For longer durations, run every N hours
		return "@every " + d.String()
	}
}

// parseSchedule parses a schedule string (e.g. "5s", "1m", "2h", "*/5 * * * *") into a cron expression
func parseSchedule(schedule string) (string, time.Duration, error) {
	// Try to parse as a duration directly
	duration, err := time.ParseDuration(schedule)
	if err == nil {
		return durationToCronExpression(duration), duration, nil
	}

	// If it's a simple number, assume seconds
	if _, err := strconv.Atoi(schedule); err == nil {
		duration, _ := time.ParseDuration(schedule + "s")
		return durationToCronExpression(duration), duration, nil
	}

	// Check if it's a standard cron expression (5 or 6 fields)
	// Standard cron: minute hour day-of-month month day-of-week
	// With seconds: second minute hour day-of-month month day-of-week
	fields := strings.Fields(schedule)
	if len(fields) == 5 || len(fields) == 6 || strings.HasPrefix(schedule, "@") {
		// Validate the cron expression by parsing it
		parser := cron.NewParser(cron.Second | cron.Minute | cron.Hour | cron.Dom | cron.Month | cron.Dow)
		_, err := parser.Parse(schedule)
		if err == nil {
			// It's a valid cron expression
			return schedule, 0, nil
		}

		return "", 0, fmt.Errorf("invalid cron expression: %w", err)
	}

	return "", 0, ErrInvalidSchedule
}

// AddTaskWithSchedule adds a new task with a schedule string
func (s *schedulerService) AddTaskWithSchedule(id, description, schedule string, taskFunc func() error) (string, error) {
	// Parse the schedule string into a cron expression
	cronExpr, interval, err := parseSchedule(schedule)
	if err != nil {
		return "", err
	}

	s.taskMutex.Lock()
	defer s.taskMutex.Unlock()

	// Check if a task with this ID already exists
	if _, exists := s.tasks[id]; exists {
		return "", ErrTaskAlreadyExists
	}

	// Calculate next run time
	nextRun := time.Now().Add(interval)

	// Create a new task
	task := &Task{
		ID:          id,
		Description: description,
		Schedule:    schedule,
		Interval:    interval,
		Func:        taskFunc,
		IsRunning:   false,
		Status:      repo.TaskStatusPending,
		NextRun:     nextRun,
	}

	// Store the task in memory
	s.tasks[id] = task

	// Persist the task to database and cache
	if s.repository != nil || s.cache != nil {
		// Create a repository task model
		repoTask := &repo.Task{
			ID:          id,
			Description: description,
			Schedule:    schedule,
			Status:      repo.TaskStatusPending,
			NextRun:     &nextRun,
			CreatedAt:   time.Now(),
			UpdatedAt:   time.Now(),
		}

		// Save to database
		if s.repository != nil {
			taskModel := repoTask.ToModel()
			ctx := context.Background()

			if err := s.repository.Create(ctx, taskModel); err != nil {
				if !errors.Is(err, repo.ErrTaskAlreadyExists) {
					s.logger.Error("Failed to save task to database",
						zap.String("id", id),
						zap.Error(err))
				}
			} else {
				s.logger.Info("Task saved to database", zap.String("id", id))
			}
		}

		// Save to cache
		if s.cache != nil {
			taskModel := repoTask.ToModel()
			ctx := context.Background()

			if err := s.cache.Set(ctx, taskModel); err != nil {
				s.logger.Warn("Failed to save task to cache",
					zap.String("id", id),
					zap.Error(err))
			} else {
				s.logger.Info("Task saved to cache", zap.String("id", id))
			}
		}
	}

	s.logger.Info("Task added to scheduler",
		zap.String("id", id),
		zap.String("description", description),
		zap.String("schedule", schedule))

	// If the scheduler is already running, start this task immediately
	if s.running {
		s.startTask(task, cronExpr)
	}

	return id, nil
}

// AddTask adds a new task to the scheduler
func (s *schedulerService) AddTask(id, description string, interval time.Duration, taskFunc func() error) (string, error) {
	// Convert duration to cron expression
	cronExpr := durationToCronExpression(interval)

	s.taskMutex.Lock()
	defer s.taskMutex.Unlock()

	// Check if task with this ID already exists
	if _, exists := s.tasks[id]; exists {
		return "", ErrTaskAlreadyExists
	}

	// Calculate next run time
	nextRun := time.Now().Add(interval)

	// Create a new task
	task := &Task{
		ID:          id,
		Description: description,
		Schedule:    interval.String(),
		Interval:    interval,
		Func:        taskFunc,
		IsRunning:   false,
		Status:      repo.TaskStatusPending,
		NextRun:     nextRun,
	}

	// Store the task in memory
	s.tasks[id] = task

	// Persist the task to database and cache
	if s.repository != nil || s.cache != nil {
		// Create a repository task model
		repoTask := &repo.Task{
			ID:          id,
			Description: description,
			Schedule:    interval.String(),
			Status:      repo.TaskStatusPending,
			NextRun:     &nextRun,
			CreatedAt:   time.Now(),
			UpdatedAt:   time.Now(),
		}

		// Save to database
		if s.repository != nil {
			taskModel := repoTask.ToModel()
			ctx := context.Background()

			if err := s.repository.Create(ctx, taskModel); err != nil {
				if !errors.Is(err, repo.ErrTaskAlreadyExists) {
					s.logger.Error("Failed to save task to database",
						zap.String("id", id),
						zap.Error(err))
				}
			} else {
				s.logger.Info("Task saved to database", zap.String("id", id))
			}
		}

		// Save to cache
		if s.cache != nil {
			taskModel := repoTask.ToModel()
			ctx := context.Background()

			if err := s.cache.Set(ctx, taskModel); err != nil {
				s.logger.Warn("Failed to save task to cache",
					zap.String("id", id),
					zap.Error(err))
			} else {
				s.logger.Info("Task saved to cache", zap.String("id", id))
			}
		}
	}

	s.logger.Info("Task added to scheduler",
		zap.String("id", id),
		zap.String("description", description),
		zap.String("interval", interval.String()))

	// If the scheduler is already running, start this task immediately
	if s.running {
		s.startTask(task, cronExpr)
	}

	return id, nil
}

// RemoveTask removes a task from the scheduler
func (s *schedulerService) RemoveTask(id string) error {
	s.taskMutex.Lock()
	defer s.taskMutex.Unlock()

	task, exists := s.tasks[id]
	if !exists {
		return ErrTaskNotFound
	}

	// Remove the task from cron if it's running
	if s.running && task.entryID != 0 {
		s.cron.Remove(task.entryID)
	}

	// Remove the task from our map
	delete(s.tasks, id)

	// Remove from database and cache
	ctx := context.Background()

	// Remove from database
	if s.repository != nil {
		if err := s.repository.Delete(ctx, id); err != nil {
			if !errors.Is(err, repo.ErrTaskNotFound) {
				s.logger.Warn("Failed to remove task from database",
					zap.String("id", id),
					zap.Error(err))
			}
		} else {
			s.logger.Info("Task removed from database", zap.String("id", id))
		}
	}

	// Remove from cache
	if s.cache != nil {
		if err := s.cache.Delete(ctx, id); err != nil {
			if !errors.Is(err, repo.ErrTaskNotFound) {
				s.logger.Warn("Failed to remove task from cache",
					zap.String("id", id),
					zap.Error(err))
			}
		} else {
			s.logger.Info("Task removed from cache", zap.String("id", id))
		}
	}

	s.logger.Info("Task removed from scheduler", zap.String("id", id))

	return nil
}

// GetTasks returns all registered tasks
func (s *schedulerService) GetTasks() []*Task {
	s.taskMutex.RLock()
	defer s.taskMutex.RUnlock()

	tasks := make([]*Task, 0, len(s.tasks))
	for _, task := range s.tasks {
		tasks = append(tasks, task)
	}

	return tasks
}

// GetTask returns a specific task by ID
func (s *schedulerService) GetTask(id string) (*Task, error) {
	s.taskMutex.RLock()
	defer s.taskMutex.RUnlock()

	task, exists := s.tasks[id]
	if !exists {
		return nil, ErrTaskNotFound
	}

	return task, nil
}

// startTask starts a single task
func (s *schedulerService) startTask(task *Task, cronExpr string) {
	// Add the task to cron
	entryID, err := s.cron.AddFunc(cronExpr, func() {
		s.executeTask(task)
	})

	if err != nil {
		s.logger.Error("Failed to add task to cron scheduler",
			zap.String("id", task.ID),
			zap.String("cronExpr", cronExpr),
			zap.Error(err))
		return
	}

	// Store the entry ID
	task.entryID = entryID

	// Calculate the next run time based on the cron schedule
	entry := s.cron.Entry(entryID)
	if !entry.Next.IsZero() {
		task.NextRun = entry.Next
	} else {
		// Fallback to using interval if cron doesn't provide next time
		task.NextRun = time.Now().Add(task.Interval)
	}
}

// Start starts the scheduler
func (s *schedulerService) Start() error {
	s.taskMutex.Lock()
	defer s.taskMutex.Unlock()

	if s.running {
		return nil // Already running
	}

	s.logger.Info("Starting scheduler service")

	// Load tasks from database
	if err := s.loadTasks(); err != nil {
		s.logger.Error("Failed to load tasks from database", zap.Error(err))
		// Continue anyway, as we might have in-memory tasks
	}

	// Start the cron scheduler
	s.cron.Start()

	// Start all tasks
	for _, task := range s.tasks {
		var cronExpr string

		// Check if the schedule is a cron expression
		if strings.Count(task.Schedule, " ") >= 4 || strings.HasPrefix(task.Schedule, "@") {
			// It's a cron expression, use it directly
			cronExpr = task.Schedule
		} else {
			// It's a duration-based schedule, convert it
			cronExpr = durationToCronExpression(task.Interval)
		}

		s.startTask(task, cronExpr)
	}

	s.running = true
	return nil
}

// loadTasks loads tasks from the database and cache
func (s *schedulerService) loadTasks() error {
	ctx := context.Background()

	// Try to get tasks from cache first if cache is available
	if s.cache != nil {
		cachedTasks, err := s.cache.GetAll(ctx)
		if err == nil && len(cachedTasks) > 0 {
			s.logger.Info("Loaded tasks from cache", zap.Int("count", len(cachedTasks)))

			// Convert cached tasks to service tasks
			for _, cachedTask := range cachedTasks {
				// Skip disabled tasks
				if cachedTask.Status == repo.TaskStatusDisabled {
					s.logger.Info("Skipping disabled task", zap.String("id", cachedTask.ID))
					continue
				}

				task := s.convertRepoTaskToServiceTask(cachedTask.ToTask())
				s.tasks[task.ID] = task
			}

			return nil
		}
	} else {
		s.logger.Warn("Cache is not available, skipping cache lookup")
	}

	// If cache failed, is empty, or not available, load from database
	if s.repository == nil {
		s.logger.Warn("Repository is not available, cannot load tasks")
		return nil
	}

	dbTasks, err := s.repository.FindAll(ctx)
	if err != nil {
		return err
	}

	s.logger.Info("Loaded tasks from database", zap.Int("count", len(dbTasks)))

	// Convert database tasks to service tasks and update cache
	for _, dbTask := range dbTasks {
		// Skip disabled tasks
		if dbTask.Status == repo.TaskStatusDisabled {
			s.logger.Info("Skipping disabled task", zap.String("id", dbTask.ID))
			continue
		}

		// Update cache if available
		if s.cache != nil {
			if err := s.cache.Set(ctx, dbTask); err != nil {
				s.logger.Warn("Failed to update task in cache", zap.String("id", dbTask.ID), zap.Error(err))
			}
		}

		task := s.convertRepoTaskToServiceTask(dbTask.ToTask())
		s.tasks[task.ID] = task
	}

	return nil
}

// convertRepoTaskToServiceTask converts a repository task to a service task
func (s *schedulerService) convertRepoTaskToServiceTask(repoTask *repo.Task) *Task {
	// Create a placeholder function that logs the task execution
	// Real functions will be registered separately
	placeholderFunc := func() error {
		s.logger.Info("Executing task loaded from database",
			zap.String("id", repoTask.ID),
			zap.String("description", repoTask.Description))
		return nil
	}

	// Parse the schedule
	var interval time.Duration

	// Check if the schedule is a cron expression
	if strings.Count(repoTask.Schedule, " ") >= 4 || strings.HasPrefix(repoTask.Schedule, "@") {
		// It's a cron expression, interval will be 0
		interval = 0
	} else {
		// Try to parse as a duration
		var err error
		interval, err = time.ParseDuration(repoTask.Schedule)
		if err != nil {
			s.logger.Warn("Failed to parse schedule as duration, treating as cron expression",
				zap.String("id", repoTask.ID),
				zap.String("schedule", repoTask.Schedule),
				zap.Error(err))
			interval = 0
		}
	}

	task := &Task{
		ID:          repoTask.ID,
		Description: repoTask.Description,
		Schedule:    repoTask.Schedule,
		Interval:    interval,
		Func:        placeholderFunc,
		IsRunning:   false,
		Status:      repoTask.Status,
	}

	// Set LastRun if available
	if repoTask.LastRun != nil {
		task.LastRun = *repoTask.LastRun
	}

	// Set NextRun if available
	if repoTask.NextRun != nil {
		task.NextRun = *repoTask.NextRun
	}

	return task
}

// Stop stops the scheduler
func (s *schedulerService) Stop() error {
	s.taskMutex.Lock()
	defer s.taskMutex.Unlock()

	if !s.running {
		return nil // Already stopped
	}

	s.logger.Info("Stopping scheduler service")

	// Stop the cron scheduler
	ctx := s.cron.Stop()

	// Wait for all running jobs to complete
	<-ctx.Done()

	// Cancel the context
	s.cancel()
	s.running = false
	return nil
}

// executeTask executes a task and updates its status
func (s *schedulerService) executeTask(task *Task) {
	ctx := context.Background()
	now := time.Now()

	s.taskMutex.Lock()
	task.IsRunning = true
	task.LastRun = now
	task.Status = repo.TaskStatusRunning

	// Update next run time if possible
	if task.entryID != 0 {
		entry := s.cron.Entry(task.entryID)
		if !entry.Next.IsZero() {
			task.NextRun = entry.Next
		} else {
			task.NextRun = now.Add(task.Interval)
		}
	}
	s.taskMutex.Unlock()

	// Update task status in database and cache
	nextRun := task.NextRun
	lastRun := task.LastRun
	go func() {
		if s.repository != nil {
			if err := s.repository.UpdateLastRun(ctx, task.ID, lastRun, repo.TaskStatusRunning, ""); err != nil {
				s.logger.Warn("Failed to update task status in database",
					zap.String("id", task.ID),
					zap.Error(err))
			}

			if err := s.repository.UpdateNextRun(ctx, task.ID, nextRun); err != nil {
				s.logger.Warn("Failed to update task next run in database",
					zap.String("id", task.ID),
					zap.Error(err))
			}
		}

		if s.cache != nil {
			if err := s.cache.UpdateLastRun(ctx, task.ID, lastRun, repo.TaskStatusRunning, ""); err != nil {
				s.logger.Warn("Failed to update task status in cache",
					zap.String("id", task.ID),
					zap.Error(err))
			}

			if err := s.cache.UpdateNextRun(ctx, task.ID, nextRun); err != nil {
				s.logger.Warn("Failed to update task next run in cache",
					zap.String("id", task.ID),
					zap.Error(err))
			}
		}
	}()

	s.logger.Info("Executing scheduled task",
		zap.String("id", task.ID),
		zap.String("description", task.Description))

	// Execute the task
	err := task.Func()
	lastError := ""
	status := repo.TaskStatusCompleted

	if err != nil {
		lastError = err.Error()
		status = repo.TaskStatusFailed
	}

	s.taskMutex.Lock()
	task.IsRunning = false
	task.Error = err
	task.Status = status
	s.taskMutex.Unlock()

	// Update task status in database and cache
	go func() {
		if s.repository != nil {
			if err := s.repository.UpdateLastRun(ctx, task.ID, lastRun, status, lastError); err != nil {
				s.logger.Warn("Failed to update task completion status in database",
					zap.String("id", task.ID),
					zap.Error(err))
			}
		}

		if s.cache != nil {
			if err := s.cache.UpdateLastRun(ctx, task.ID, lastRun, status, lastError); err != nil {
				s.logger.Warn("Failed to update task completion status in cache",
					zap.String("id", task.ID),
					zap.Error(err))
			}
		}
	}()

	if err != nil {
		s.logger.Error("Task execution failed",
			zap.String("id", task.ID),
			zap.Error(err))
	} else {
		s.logger.Info("Task executed successfully",
			zap.String("id", task.ID))
	}
}
