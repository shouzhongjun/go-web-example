package scheduler

import (
	"time"
)

// TaskStatus represents the status of a scheduled task
type TaskStatus string

const (
	// TaskStatusPending indicates the task is waiting to be executed
	TaskStatusPending TaskStatus = "pending"
	// TaskStatusRunning indicates the task is currently running
	TaskStatusRunning TaskStatus = "running"
	// TaskStatusCompleted indicates the task has completed successfully
	TaskStatusCompleted TaskStatus = "completed"
	// TaskStatusFailed indicates the task has failed
	TaskStatusFailed TaskStatus = "failed"
	// TaskStatusDisabled indicates the task is disabled and should not be executed
	TaskStatusDisabled TaskStatus = "disabled"
)

// TaskModel represents a scheduled task in the database
type TaskModel struct {
	ID          string     `gorm:"primaryKey;type:varchar(64)" json:"id"`
	Description string     `gorm:"type:varchar(255)" json:"description"`
	Schedule    string     `gorm:"type:varchar(100)" json:"schedule"`
	Status      TaskStatus `gorm:"type:varchar(20);default:'pending'" json:"status"`
	LastRun     *time.Time `gorm:"type:datetime" json:"last_run"`
	NextRun     *time.Time `gorm:"type:datetime" json:"next_run"`
	LastError   string     `gorm:"type:text" json:"last_error"`
	CreatedAt   time.Time  `gorm:"type:datetime;autoCreateTime" json:"created_at"`
	UpdatedAt   time.Time  `gorm:"type:datetime;autoUpdateTime" json:"updated_at"`
}

// TableName specifies the table name for the TaskModel
func (TaskModel) TableName() string {
	return "scheduler_tasks"
}

// ToTask converts a TaskModel to a Task (used in the service layer)
func (m *TaskModel) ToTask() *Task {
	task := &Task{
		ID:          m.ID,
		Description: m.Description,
		Schedule:    m.Schedule,
		Status:      m.Status,
		LastRun:     m.LastRun,
		NextRun:     m.NextRun,
		LastError:   m.LastError,
		CreatedAt:   m.CreatedAt,
		UpdatedAt:   m.UpdatedAt,
	}
	return task
}

// Task represents a scheduled task in the service layer
type Task struct {
	ID          string     `json:"id"`
	Description string     `json:"description"`
	Schedule    string     `json:"schedule"`
	Status      TaskStatus `json:"status"`
	LastRun     *time.Time `json:"last_run"`
	NextRun     *time.Time `json:"next_run"`
	LastError   string     `json:"last_error"`
	CreatedAt   time.Time  `json:"created_at"`
	UpdatedAt   time.Time  `json:"updated_at"`
	// Fields not stored in the database
	Func      func() error `json:"-"`
	IsRunning bool         `json:"-"`
}

// ToModel converts a Task to a TaskModel (used in the repository layer)
func (t *Task) ToModel() *TaskModel {
	model := &TaskModel{
		ID:          t.ID,
		Description: t.Description,
		Schedule:    t.Schedule,
		Status:      t.Status,
		LastRun:     t.LastRun,
		NextRun:     t.NextRun,
		LastError:   t.LastError,
		CreatedAt:   t.CreatedAt,
		UpdatedAt:   t.UpdatedAt,
	}
	return model
}