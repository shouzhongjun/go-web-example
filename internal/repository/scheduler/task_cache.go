package scheduler

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/go-redis/redis/v8"
	"go.uber.org/zap"
)

const (
	// TaskKeyPrefix is the prefix for task keys in Redis
	TaskKeyPrefix = "scheduler:task:"
	// TaskListKey is the key for the list of all task IDs in Redis
	TaskListKey = "scheduler:tasks"
	// DefaultTaskTTL is the default TTL for task cache entries
	DefaultTaskTTL = 24 * time.Hour
)

// TaskCache defines the interface for task caching operations
type TaskCache interface {
	// Set stores a task in the cache
	Set(ctx context.Context, task *TaskModel) error
	// Get retrieves a task from the cache
	Get(ctx context.Context, id string) (*TaskModel, error)
	// Delete removes a task from the cache
	Delete(ctx context.Context, id string) error
	// GetAll retrieves all tasks from the cache
	GetAll(ctx context.Context) ([]*TaskModel, error)
	// UpdateStatus updates the status of a task in the cache
	UpdateStatus(ctx context.Context, id string, status TaskStatus) error
	// UpdateLastRun updates the last run information of a task in the cache
	UpdateLastRun(ctx context.Context, id string, lastRun time.Time, status TaskStatus, lastError string) error
	// UpdateNextRun updates the next run time of a task in the cache
	UpdateNextRun(ctx context.Context, id string, nextRun time.Time) error
}

// taskCache implements the TaskCache interface
type taskCache struct {
	client *redis.Client
	logger *zap.Logger
	ttl    time.Duration
}

// NewTaskCache creates a new task cache
func NewTaskCache(client *redis.Client, logger *zap.Logger) TaskCache {
	return &taskCache{
		client: client,
		logger: logger,
		ttl:    DefaultTaskTTL,
	}
}

// taskKey returns the Redis key for a task
func taskKey(id string) string {
	return TaskKeyPrefix + id
}

// Set stores a task in the cache
func (c *taskCache) Set(ctx context.Context, task *TaskModel) error {
	// Marshal the task to JSON
	taskJSON, err := json.Marshal(task)
	if err != nil {
		return fmt.Errorf("failed to marshal task: %w", err)
	}

	// Store the task in Redis
	key := taskKey(task.ID)
	if err := c.client.Set(ctx, key, taskJSON, c.ttl).Err(); err != nil {
		return fmt.Errorf("failed to set task in cache: %w", err)
	}

	// Add the task ID to the list of all tasks
	if err := c.client.SAdd(ctx, TaskListKey, task.ID).Err(); err != nil {
		c.logger.Warn("Failed to add task ID to list", zap.String("id", task.ID), zap.Error(err))
	}

	c.logger.Debug("Task stored in cache", zap.String("id", task.ID))
	return nil
}

// Get retrieves a task from the cache
func (c *taskCache) Get(ctx context.Context, id string) (*TaskModel, error) {
	// Get the task from Redis
	key := taskKey(id)
	taskJSON, err := c.client.Get(ctx, key).Result()
	if err != nil {
		if err == redis.Nil {
			return nil, ErrTaskNotFound
		}
		return nil, fmt.Errorf("failed to get task from cache: %w", err)
	}

	// Unmarshal the task from JSON
	var task TaskModel
	if err := json.Unmarshal([]byte(taskJSON), &task); err != nil {
		return nil, fmt.Errorf("failed to unmarshal task: %w", err)
	}

	return &task, nil
}

// Delete removes a task from the cache
func (c *taskCache) Delete(ctx context.Context, id string) error {
	// Delete the task from Redis
	key := taskKey(id)
	if err := c.client.Del(ctx, key).Err(); err != nil {
		return fmt.Errorf("failed to delete task from cache: %w", err)
	}

	// Remove the task ID from the list of all tasks
	if err := c.client.SRem(ctx, TaskListKey, id).Err(); err != nil {
		c.logger.Warn("Failed to remove task ID from list", zap.String("id", id), zap.Error(err))
	}

	c.logger.Debug("Task deleted from cache", zap.String("id", id))
	return nil
}

// GetAll retrieves all tasks from the cache
func (c *taskCache) GetAll(ctx context.Context) ([]*TaskModel, error) {
	// Get all task IDs from the list
	taskIDs, err := c.client.SMembers(ctx, TaskListKey).Result()
	if err != nil {
		return nil, fmt.Errorf("failed to get task IDs from cache: %w", err)
	}

	if len(taskIDs) == 0 {
		return []*TaskModel{}, nil
	}

	// Get all tasks in parallel using a pipeline
	pipe := c.client.Pipeline()
	cmds := make(map[string]*redis.StringCmd)

	for _, id := range taskIDs {
		cmds[id] = pipe.Get(ctx, taskKey(id))
	}

	_, err = pipe.Exec(ctx)
	if err != nil && err != redis.Nil {
		return nil, fmt.Errorf("failed to execute pipeline: %w", err)
	}

	// Process the results
	tasks := make([]*TaskModel, 0, len(taskIDs))
	for id, cmd := range cmds {
		taskJSON, err := cmd.Result()
		if err != nil {
			if err == redis.Nil {
				// Task was deleted between getting the list and retrieving it
				continue
			}
			c.logger.Warn("Failed to get task from cache", zap.String("id", id), zap.Error(err))
			continue
		}

		var task TaskModel
		if err := json.Unmarshal([]byte(taskJSON), &task); err != nil {
			c.logger.Warn("Failed to unmarshal task", zap.String("id", id), zap.Error(err))
			continue
		}

		tasks = append(tasks, &task)
	}

	return tasks, nil
}

// UpdateStatus updates the status of a task in the cache
func (c *taskCache) UpdateStatus(ctx context.Context, id string, status TaskStatus) error {
	// Get the task from cache
	task, err := c.Get(ctx, id)
	if err != nil {
		return err
	}

	// Update the status
	task.Status = status
	task.UpdatedAt = time.Now()

	// Store the updated task
	return c.Set(ctx, task)
}

// UpdateLastRun updates the last run information of a task in the cache
func (c *taskCache) UpdateLastRun(ctx context.Context, id string, lastRun time.Time, status TaskStatus, lastError string) error {
	// Get the task from cache
	task, err := c.Get(ctx, id)
	if err != nil {
		return err
	}

	// Update the last run information
	task.LastRun = &lastRun
	task.Status = status
	task.LastError = lastError
	task.UpdatedAt = time.Now()

	// Store the updated task
	return c.Set(ctx, task)
}

// UpdateNextRun updates the next run time of a task in the cache
func (c *taskCache) UpdateNextRun(ctx context.Context, id string, nextRun time.Time) error {
	// Get the task from cache
	task, err := c.Get(ctx, id)
	if err != nil {
		return err
	}

	// Update the next run time
	task.NextRun = &nextRun
	task.UpdatedAt = time.Now()

	// Store the updated task
	return c.Set(ctx, task)
}