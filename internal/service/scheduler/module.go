package scheduler

import (
	"context"
	"goWebExample/internal/infra/cache"
	"goWebExample/internal/infra/di/container"
	"goWebExample/internal/pkg/handlers"
	"goWebExample/internal/pkg/module"
	repo "goWebExample/internal/repository/scheduler"

	"go.uber.org/zap"
	"gorm.io/gorm"
)

// addDemoTasks adds some demo tasks to the scheduler
func addDemoTasks(svc SchedulerService, logger *zap.Logger) {
	// Add a task that runs every 30 seconds
	_, err := svc.AddTaskWithSchedule(
		"demo-task-1",
		"Demo task that logs a message every 30 seconds",
		"30s",
		func() error {
			logger.Info("Demo task 1 executed")
			return nil
		},
	)
	if err != nil {
		logger.Error("Failed to add demo task 1", zap.Error(err))
	}

	// Add a task that runs every minute
	_, err = svc.AddTaskWithSchedule(
		"demo-task-2",
		"Demo task that logs a message every minute",
		"1m",
		func() error {
			logger.Info("Demo task 2 executed")
			return nil
		},
	)
	if err != nil {
		logger.Error("Failed to add demo task 2", zap.Error(err))
	}

	// Add a task that runs every 5 minutes using a cron expression
	_, err = svc.AddTaskWithSchedule(
		"demo-task-3",
		"Demo task that logs a message every 5 minutes using cron expression",
		"0 */5 * * * *", // Seconds Minutes Hours DayOfMonth Month DayOfWeek
		func() error {
			logger.Info("Demo task 3 executed (cron expression)")
			return nil
		},
	)
	if err != nil {
		logger.Error("Failed to add demo task 3", zap.Error(err))
	}
}

func init() {
	// Register the scheduler module
	module.GetRegistry().Register(module.NewBaseModule(
		"scheduler",
		// Service creator function
		func(logger *zap.Logger, container *container.ServiceContainer) (string, interface{}) {
			// Get database connector
			var db *gorm.DB
			dbConnector := container.GetDBConnector()
			if dbConnector == nil {
				logger.Warn("MySQL connector not available, scheduler will run without database persistence")
			} else {
				db = dbConnector.GetDB()

				// Check if db is not nil before auto migrating
				if db != nil {
					// Auto migrate the task model
					if err := db.AutoMigrate(&repo.TaskModel{}); err != nil {
						logger.Error("Failed to auto migrate scheduler task model", zap.Error(err))
					} else {
						logger.Info("Scheduler task model auto migrated")
					}
				} else {
					logger.Warn("Database instance is nil, skipping auto migration")
				}
			}

			// Get Redis connector from the factory
			var redisClient *cache.RedisConnector
			factory := container.GetFactory()
			if factory == nil {
				logger.Warn("Service factory not available, scheduler will run without cache")
			} else {
				redisConnector := factory.GetConnector("redis")
				if redisConnector == nil {
					logger.Warn("Redis connector not available, scheduler will run without cache")
				} else {
					if redisConn, ok := redisConnector.(*cache.RedisConnector); ok {
						redisClient = redisConn
					} else {
						logger.Warn("Redis connector is not of expected type, scheduler will run without cache")
					}
				}
			}

			// Create repository and cache
			var taskRepo repo.TaskRepository
			var taskCache repo.TaskCache

			if db != nil {
				taskRepo = repo.NewTaskRepository(db, logger)
				logger.Info("Scheduler task repository created")
			}

			if redisClient != nil {
				// Ensure Redis is connected before using it
				if err := redisClient.Connect(context.Background()); err != nil {
					logger.Warn("Failed to connect to Redis", zap.Error(err))
				} else if redisClient.GetClient() != nil {
					taskCache = repo.NewTaskCache(redisClient.GetClient(), logger)
					logger.Info("Scheduler task cache created")
				} else {
					logger.Warn("Redis client is nil after connection, scheduler will run without cache")
				}
			}

			// Create scheduler service
			schedulerSvc := NewSchedulerService(logger, taskRepo, taskCache)
			logger.Info("Scheduler service created with persistence")

			// Add demo tasks
			addDemoTasks(schedulerSvc, logger)

			// Start the scheduler service
			if err := schedulerSvc.Start(); err != nil {
				logger.Error("Failed to start scheduler service", zap.Error(err))
				return "", nil
			}

			logger.Info("Scheduler service registered and started")
			return ServiceName, schedulerSvc
		},
		// Handler creator function - no handler needed for this service
		func(logger *zap.Logger) handlers.Handler {
			return nil
		},
	))
}
