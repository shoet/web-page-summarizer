package list_task

import (
	"context"
	"fmt"

	"github.com/shoet/webpagesummary/pkg/infrastracture"
	"github.com/shoet/webpagesummary/pkg/infrastracture/entities"
	"github.com/shoet/webpagesummary/pkg/infrastracture/repository"
)

type TaskRepository interface {
	ListTask(ctx context.Context, tx infrastracture.Transactor, input *repository.ListTaskInput) ([]*entities.Task, error)
}

type Usecase struct {
	DBHandler      *infrastracture.DBHandler
	TaskRepository TaskRepository
}

func NewUsecase(dbHandler *infrastracture.DBHandler, taskRepository TaskRepository) *Usecase {
	return &Usecase{
		DBHandler:      dbHandler,
		TaskRepository: taskRepository,
	}
}

type UsecaseInput struct {
	UserId string
	Status *string
	Limit  uint
	Offset uint
}

func (u *Usecase) Run(ctx context.Context, input UsecaseInput) ([]*entities.Task, error) {
	repoInput := &repository.ListTaskInput{
		Status: input.Status,
		Limit:  func() *uint { return &input.Limit }(),
		Offset: func() *uint { return &input.Offset }(),
	}
	tx, err := u.DBHandler.GetTransaction()
	if err != nil {
		return nil, fmt.Errorf("failed GetTransaction: %w", err)
	}
	tasks, err := u.TaskRepository.ListTask(ctx, tx, repoInput)
	if err != nil {
		return nil, fmt.Errorf("failed ListTask: %w", err)
	}
	return tasks, nil
}
