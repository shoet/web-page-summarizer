package repository

import (
	"context"
	"fmt"
	"time"

	"github.com/doug-martin/goqu/v9"
	"github.com/shoet/webpagesummary/pkg/infrastracture"
	"github.com/shoet/webpagesummary/pkg/infrastracture/entities"
)

/*
task.goはRDB上のtaskテーブルにアクセスするためのリポジトリを提供するファイルです。
*/

type TaskRepository struct {
}

func NewTaskRepository() *TaskRepository {
	return &TaskRepository{}
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
		t.Id, t.TaskStatus, t.Title, t.PageUrl, now.Unix(), now.Unix(),
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
		t.Id, t.TaskStatus, t.Title, t.PageUrl, now.Unix(),
	); err != nil {
		return fmt.Errorf("failed ExecContext: %w", err)
	}
	return nil
}

type ListTaskInput struct {
	Status *string
	Limit  *uint
	Offset *uint
}

func (r *TaskRepository) ListTask(
	ctx context.Context, tx infrastracture.Transactor, input *ListTaskInput,
) ([]*entities.Task, error) {

	var builder *goqu.SelectDataset
	builder = goqu.
		From("tasks").
		Select("id", "task_id", "task_status", "title", "page_url", "created_at", "updated_at")

	if input.Status != nil {
		builder = builder.Where(goqu.Ex{"task_status": *input.Status})
	}

	if input.Offset != nil {
		builder = builder.Offset(*input.Offset)
	}

	if input.Limit != nil {
		builder = builder.Limit(*input.Limit)
	}

	builder = builder.Order(goqu.I("id").Desc()) // AutoIncrementの降順で取得

	query, _, err := builder.ToSQL()
	if err != nil {
		return nil, fmt.Errorf("failed to goqu.ToSQL: %v", err)
	}

	var tasks []*entities.Task
	if err := tx.SelectContext(ctx, &tasks, query); err != nil {
		return nil, fmt.Errorf("failed to SelectContext: %v", err)
	}

	return tasks, nil
}
