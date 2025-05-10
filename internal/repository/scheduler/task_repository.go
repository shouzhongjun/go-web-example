package scheduler

import (
	"context"
	"errors"
	"fmt"
	"time"

	"go.uber.org/zap"
	"gorm.io/gorm"
)

// Repository errors
var (
	ErrTaskNotFound      = errors.New("task not found")
	ErrTaskAlreadyExists = errors.New("task with this ID already exists")
)

// TaskRepository defines the interface for task persistence operations
type TaskRepository interface {
	// Create creates a new task in the database
	Create(ctx context.Context, task *TaskModel) error
	// Update updates an existing task in the database
	Update(ctx context.Context, task *TaskModel) error
	// Delete deletes a task from the database
	Delete(ctx context.Context, id string) error
	// FindByID finds a task by its ID
	FindByID(ctx context.Context, id string) (*TaskModel, error)
	// FindAll returns all tasks
	FindAll(ctx context.Context) ([]*TaskModel, error)
	// FindByStatus returns all tasks with the given status
	FindByStatus(ctx context.Context, status TaskStatus) ([]*TaskModel, error)
	// UpdateStatus updates the status of a task
	UpdateStatus(ctx context.Context, id string, status TaskStatus) error
	// UpdateNextRun updates the next run time of a task
	UpdateNextRun(ctx context.Context, id string, nextRun time.Time) error
	// UpdateLastRun updates the last run time and status of a task
	UpdateLastRun(ctx context.Context, id string, lastRun time.Time, status TaskStatus, lastError string) error
}

// taskRepository implements the TaskRepository interface
type taskRepository struct {
	db     *gorm.DB
	logger *zap.Logger
}

// NewTaskRepository creates a new task repository
func NewTaskRepository(db *gorm.DB, logger *zap.Logger) TaskRepository {
	return &taskRepository{
		db:     db,
		logger: logger,
	}
}

// Create creates a new task in the database
func (r *taskRepository) Create(ctx context.Context, task *TaskModel) error {
	// Check if task already exists
	var count int64
	if err := r.db.WithContext(ctx).Model(&TaskModel{}).Where("id = ?", task.ID).Count(&count).Error; err != nil {
		return fmt.Errorf("failed to check if task exists: %w", err)
	}

	if count > 0 {
		return ErrTaskAlreadyExists
	}

	// Create the task
	if err := r.db.WithContext(ctx).Create(task).Error; err != nil {
		return fmt.Errorf("failed to create task: %w", err)
	}

	r.logger.Info("Task created", zap.String("id", task.ID))
	return nil
}

// Update updates an existing task in the database
func (r *taskRepository) Update(ctx context.Context, task *TaskModel) error {
	result := r.db.WithContext(ctx).Save(task)
	if result.Error != nil {
		return fmt.Errorf("failed to update task: %w", result.Error)
	}

	if result.RowsAffected == 0 {
		return ErrTaskNotFound
	}

	r.logger.Info("Task updated", zap.String("id", task.ID))
	return nil
}

// Delete deletes a task from the database
func (r *taskRepository) Delete(ctx context.Context, id string) error {
	result := r.db.WithContext(ctx).Delete(&TaskModel{}, "id = ?", id)
	if result.Error != nil {
		return fmt.Errorf("failed to delete task: %w", result.Error)
	}

	if result.RowsAffected == 0 {
		return ErrTaskNotFound
	}

	r.logger.Info("Task deleted", zap.String("id", id))
	return nil
}

// FindByID finds a task by its ID
func (r *taskRepository) FindByID(ctx context.Context, id string) (*TaskModel, error) {
	var task TaskModel
	if err := r.db.WithContext(ctx).Where("id = ?", id).First(&task).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrTaskNotFound
		}
		return nil, fmt.Errorf("failed to find task: %w", err)
	}

	return &task, nil
}

// FindAll returns all tasks
func (r *taskRepository) FindAll(ctx context.Context) ([]*TaskModel, error) {
	var tasks []*TaskModel
	if err := r.db.WithContext(ctx).Find(&tasks).Error; err != nil {
		return nil, fmt.Errorf("failed to find tasks: %w", err)
	}

	return tasks, nil
}

// FindByStatus returns all tasks with the given status
func (r *taskRepository) FindByStatus(ctx context.Context, status TaskStatus) ([]*TaskModel, error) {
	var tasks []*TaskModel
	if err := r.db.WithContext(ctx).Where("status = ?", status).Find(&tasks).Error; err != nil {
		return nil, fmt.Errorf("failed to find tasks by status: %w", err)
	}

	return tasks, nil
}

// UpdateStatus updates the status of a task
func (r *taskRepository) UpdateStatus(ctx context.Context, id string, status TaskStatus) error {
	result := r.db.WithContext(ctx).Model(&TaskModel{}).Where("id = ?", id).Update("status", status)
	if result.Error != nil {
		return fmt.Errorf("failed to update task status: %w", result.Error)
	}

	if result.RowsAffected == 0 {
		return ErrTaskNotFound
	}

	r.logger.Info("Task status updated", zap.String("id", id), zap.String("status", string(status)))
	return nil
}

// UpdateNextRun updates the next run time of a task
func (r *taskRepository) UpdateNextRun(ctx context.Context, id string, nextRun time.Time) error {
	result := r.db.WithContext(ctx).Model(&TaskModel{}).Where("id = ?", id).Update("next_run", nextRun)
	if result.Error != nil {
		return fmt.Errorf("failed to update task next run: %w", result.Error)
	}

	if result.RowsAffected == 0 {
		return ErrTaskNotFound
	}

	r.logger.Info("Task next run updated", zap.String("id", id), zap.Time("next_run", nextRun))
	return nil
}

// UpdateLastRun updates the last run time and status of a task
func (r *taskRepository) UpdateLastRun(ctx context.Context, id string, lastRun time.Time, status TaskStatus, lastError string) error {
	updates := map[string]interface{}{
		"last_run":   lastRun,
		"status":     status,
		"last_error": lastError,
	}

	result := r.db.WithContext(ctx).Model(&TaskModel{}).Where("id = ?", id).Updates(updates)
	if result.Error != nil {
		return fmt.Errorf("failed to update task last run: %w", result.Error)
	}

	if result.RowsAffected == 0 {
		return ErrTaskNotFound
	}

	r.logger.Info("Task last run updated", 
		zap.String("id", id), 
		zap.Time("last_run", lastRun), 
		zap.String("status", string(status)))
	return nil
}