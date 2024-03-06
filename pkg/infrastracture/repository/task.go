package repository

import (
	"context"
	"fmt"
	"time"

	"github.com/shoet/webpagesummary/pkg/infrastracture"
	"github.com/shoet/webpagesummary/pkg/infrastracture/entities"
)

/*
task.goはRDB上のtaskテーブルにアクセスするためのリポジトリを提供するファイルです。
*/

type TaskRepository struct {
	db *infrastracture.DBHandler
}

func NewTaskRepository(db *infrastracture.DBHandler) *TaskRepository {
	return &TaskRepository{db: db}
}

func (r *TaskRepository) AddTask(ctx context.Context, tx infrastracture.Transactor, t *entities.Summary) error {
	now := time.Now()
	query := `
	INSERT INTO tasks
		(task_id, task_status, title, page_url, created_at, updated_at)
	VALUES
		($1, $2, $3, $4, $5, $6)
	`
	if _, err := tx.ExecContext(
		ctx, query,
		t.Id, t.TaskStatus, t.Title, t.PageUrl, now, now,
	); err != nil {
		return fmt.Errorf("failed ExecContext: %w", err)
	}

	return nil
}

func (r *TaskRepository) UpdateTask(ctx context.Context, tx infrastracture.Transactor, t *entities.Summary) error {
	now := time.Now()
	query := `
	UPDATE tasks
	SET	
		task_status = $2,
		title = $3,
		page_url = $4,
		updated_at = $5
	WHERE task_id = $1
	`
	if _, err := tx.ExecContext(
		ctx, query,
		t.Id, t.TaskStatus, t.Title, t.PageUrl, now,
	); err != nil {
		return fmt.Errorf("failed ExecContext: %w", err)
	}
	return nil
}
