package postgres

import (
	"context"
	"errors"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/shopspring/decimal"

	"gopricemon/internal/domain"
	"gopricemon/internal/repository"
)

func (r *Repository) GetExecutorByToken(token string) (domain.Executor, error) {
	const query = `
		SELECT id, name, token::text
		FROM executors
		WHERE token::text = $1
	`

	var executor domain.Executor
	err := r.pool.QueryRow(context.Background(), query, token).Scan(
		&executor.ID, &executor.Name, &executor.Token,
	)
	if errors.Is(err, pgx.ErrNoRows) {
		return domain.Executor{}, repository.ErrExecutorNotFound
	}

	return executor, err
}

func (r *Repository) ClaimTask(platformID, executorID, leaseSeconds int) (domain.Task, bool, error) {
	const query = `
		WITH next_task AS (
			SELECT id
			FROM tasks
			WHERE platform_id = $1
				AND (status = 'not started' OR (status = 'in progress' AND lease_until <= NOW()))
			ORDER BY id
			FOR UPDATE SKIP LOCKED
			LIMIT 1
		)
		UPDATE tasks
		SET
			status = 'in progress',
			executor_id = $2,
			lease_token = gen_random_uuid(),
			lease_until = NOW() + $3 * INTERVAL '1 second',
			attempts = attempts + 1
		FROM next_task
		WHERE tasks.id = next_task.id
		RETURNING tasks.id, tasks.platform_id, tasks.action_type, tasks.price, tasks.lease_token::text, tasks.lease_until
	`

	var task domain.Task
	err := r.pool.QueryRow(context.Background(), query, platformID, executorID, leaseSeconds).Scan(
		&task.ID,
		&task.PlatformID,
		&task.ActionType,
		&task.Price,
		&task.LeaseToken,
		&task.LeaseUntil,
	)
	if errors.Is(err, pgx.ErrNoRows) {
		return domain.Task{}, false, nil
	}

	return task, true, err
}

func (r *Repository) ExtendTaskLease(taskID, executorID int, leaseToken string, leaseSeconds int) (time.Time, bool, error) {
	const query = `
		UPDATE tasks
		SET lease_until = NOW() + $4 * INTERVAL '1 second'
		WHERE id = $1
			AND executor_id = $2
			AND lease_token::text = $3
			AND status = 'in progress'
			AND lease_until > NOW()
		RETURNING lease_until
	`

	var leaseUntil time.Time
	err := r.pool.QueryRow(context.Background(), query, taskID, executorID, leaseToken, leaseSeconds).Scan(&leaseUntil)
	if errors.Is(err, pgx.ErrNoRows) {
		return time.Time{}, false, nil
	}

	return leaseUntil, true, err
}

func (r *Repository) ReportTaskResult(taskID, executorID int, leaseToken string, status domain.TaskStatus, errorText *string, completedAt time.Time) (*domain.TaskReportState, bool, error) {
	const updateQuery = `
		UPDATE tasks
		SET
			status = $4,
			error = $5,
			completed_at = $6,
			lease_token = NULL,
			lease_until = NULL
		WHERE id = $1
			AND executor_id = $2
			AND lease_token::text = $3
			AND status = 'in progress'
			AND lease_until > NOW()
		RETURNING status
	`

	var updatedStatus domain.TaskStatus
	err := r.pool.QueryRow(context.Background(), updateQuery, taskID, executorID, leaseToken, status, errorText, completedAt).Scan(&updatedStatus)
	if err == nil {
		return nil, true, nil
	}
	if !errors.Is(err, pgx.ErrNoRows) {
		return nil, false, err
	}

	const stateQuery = `
		SELECT status, executor_id, lease_token::text, COALESCE(lease_until <= NOW(), TRUE)
		FROM tasks
		WHERE id = $1
	`

	var task domain.TaskReportState
	err = r.pool.QueryRow(context.Background(), stateQuery, taskID).Scan(
		&task.Status,
		&task.ExecutorID,
		&task.LeaseToken,
		&task.LeaseExpired,
	)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, false, nil
	}
	if err != nil {
		return nil, false, err
	}

	return &task, false, nil
}

func (r *Repository) GetTasks() ([]domain.TaskInfo, error) {
	const query = `
		SELECT
			t.id,
			i.name,
			c.name,
			p.name,
			e.name,
			t.action_type,
			t.price,
			t.status,
			t.error,
			t.lease_until,
			t.completed_at
		FROM tasks AS t
		LEFT JOIN items AS i ON i.id = t.item_id
		LEFT JOIN categories AS c ON c.id = i.category_id
		JOIN platforms AS p ON p.id = t.platform_id
		LEFT JOIN executors AS e ON e.id = t.executor_id
		ORDER BY t.id DESC
	`

	rows, err := r.pool.Query(context.Background(), query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var tasks []domain.TaskInfo
	for rows.Next() {
		var task domain.TaskInfo
		if err = rows.Scan(
			&task.ID,
			&task.ItemName,
			&task.CategoryName,
			&task.PlatformName,
			&task.ExecutorName,
			&task.ActionType,
			&task.Price,
			&task.Status,
			&task.Error,
			&task.LeaseUntil,
			&task.CompletedAt,
		); err != nil {
			return nil, err
		}
		tasks = append(tasks, task)
	}

	return tasks, rows.Err()
}

func (r *Repository) GetTask(taskID int) (*domain.TaskInfo, error) {
	const query = `
		SELECT
			t.id,
			i.name,
			c.name,
			p.name,
			e.name,
			t.action_type,
			t.price,
			t.status,
			t.error,
			t.lease_until,
			t.completed_at
		FROM tasks AS t
		LEFT JOIN items AS i ON i.id = t.item_id
		LEFT JOIN categories AS c ON c.id = i.category_id
		JOIN platforms AS p ON p.id = t.platform_id
		LEFT JOIN executors AS e ON e.id = t.executor_id
		WHERE t.id = $1
	`

	var task domain.TaskInfo
	err := r.pool.QueryRow(context.Background(), query, taskID).Scan(
		&task.ID,
		&task.ItemName,
		&task.CategoryName,
		&task.PlatformName,
		&task.ExecutorName,
		&task.ActionType,
		&task.Price,
		&task.Status,
		&task.Error,
		&task.LeaseUntil,
		&task.CompletedAt,
	)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}

	return &task, nil
}

func (r *Repository) CreateTask(itemID int, platformName, actionType string, price decimal.Decimal, taskKey string) (domain.TaskInfo, error) {
	const query = `
		INSERT INTO tasks (item_id, platform_id, action_type, price, status, task_key)
		SELECT i.id, p.id, $3, $4, 'not started', NULLIF($5, '')
		FROM items AS i
		JOIN platforms AS p ON p.name = $2
		WHERE i.id = $1
		ON CONFLICT (task_key) DO UPDATE
		SET task_key = EXCLUDED.task_key
		RETURNING
			id,
			(SELECT name FROM items WHERE id = tasks.item_id),
			(SELECT c.name FROM items AS i JOIN categories AS c ON c.id = i.category_id WHERE i.id = tasks.item_id),
			(SELECT name FROM platforms WHERE id = tasks.platform_id),
			NULL::TEXT,
			action_type,
			price,
			status,
			error,
			lease_until
	`

	var task domain.TaskInfo
	err := r.pool.QueryRow(context.Background(), query, itemID, platformName, actionType, price, taskKey).Scan(
		&task.ID,
		&task.ItemName,
		&task.CategoryName,
		&task.PlatformName,
		&task.ExecutorName,
		&task.ActionType,
		&task.Price,
		&task.Status,
		&task.Error,
		&task.LeaseUntil,
	)
	if errors.Is(err, pgx.ErrNoRows) {
		return domain.TaskInfo{}, repository.ErrTaskReferencesNotFound
	}
	return task, err
}

func (r *Repository) DeleteTask(taskID int) (bool, error) {
	result, err := r.pool.Exec(
		context.Background(), `DELETE FROM tasks WHERE id = $1 AND status <> 'in progress'`, taskID,
	)
	if err != nil {
		return false, err
	}

	return result.RowsAffected() > 0, nil
}
