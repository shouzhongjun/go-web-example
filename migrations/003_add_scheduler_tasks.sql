-- Create a scheduler_tasks table
DROP TABLE IF EXISTS `scheduler_tasks`;

CREATE TABLE IF NOT EXISTS `scheduler_tasks` (
  `id` VARCHAR(64) NOT NULL,
  `description` VARCHAR(255) NOT NULL,
  `schedule` VARCHAR(100) NOT NULL,
  `status` VARCHAR(20) NOT NULL DEFAULT 'pending',
  `last_run` DATETIME NULL,
  `next_run` DATETIME NULL,
  `last_error` TEXT NULL,
  `created_at` DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
  `updated_at` DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  PRIMARY KEY (`id`),
  INDEX `idx_scheduler_tasks_status` (`status`),
  INDEX `idx_scheduler_tasks_next_run` (`next_run`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

-- Insert demo tasks
INSERT INTO `scheduler_tasks` (`id`, `description`, `schedule`, `status`, `next_run`)
VALUES 
  ('demo-task-1', 'Demo task that logs a message every 30 seconds', '30s', 'pending', DATE_ADD(NOW(), INTERVAL 30 SECOND)),
  ('demo-task-2', 'Demo task that logs a message every minute', '1m', 'pending', DATE_ADD(NOW(), INTERVAL 1 MINUTE)),
  ('demo-task-3', 'Demo task that logs a message every 5 minutes using cron expression', '0 */5 * * * *', 'pending', DATE_ADD(NOW(), INTERVAL 5 MINUTE));
