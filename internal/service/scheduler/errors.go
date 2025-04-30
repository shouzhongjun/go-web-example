package scheduler

import (
	"errors"
)

// Error constants for the scheduler service
var (
	ErrTaskAlreadyExists = errors.New("task with this ID already exists")
	ErrTaskNotFound      = errors.New("task not found")
	ErrInvalidSchedule   = errors.New("invalid schedule format")
)
