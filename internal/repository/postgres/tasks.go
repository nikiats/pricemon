package postgres

import (
	"context"
	"errors"
	"time"

	"github.com/jackc/pgx/v5"

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

func (r *Repository) ReportTaskResult(taskID, executorID int, leaseToken string, status domain.TaskStatus, errorText *string) (bool, error) {
	const query = `
		UPDATE tasks
		SET
			status = $4,
			error = $5,
			lease_token = NULL,
			lease_until = NULL
		WHERE id = $1
			AND executor_id = $2
			AND lease_token::text = $3
			AND status = 'in progress'
			AND lease_until > NOW()
	`

	result, err := r.pool.Exec(context.Background(), query, taskID, executorID, leaseToken, status, errorText)
	if err != nil {
		return false, err
	}

	return result.RowsAffected() > 0, nil
}
