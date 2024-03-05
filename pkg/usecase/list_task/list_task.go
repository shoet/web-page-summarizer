package list_task

import (
	"context"
	"fmt"

	"github.com/shoet/webpagesummary/pkg/infrastracture/entities"
)

type SummaryRepository interface {
	ListTask(ctx context.Context, status *string, nextToken *string, limit int32) ([]*entities.Summary, *string, error)
}

type Usecase struct {
	SummaryRepository SummaryRepository
}

func NewUsecase(summaryRepository SummaryRepository) *Usecase {
	return &Usecase{SummaryRepository: summaryRepository}
}

type UsecaseInput struct {
	Status    *string
	NextToken *string
	Limit     int32
}

func (u *Usecase) Run(ctx context.Context, input UsecaseInput) ([]*entities.Summary, *string, error) {
	tasks, nextToken, err := u.SummaryRepository.ListTask(ctx, input.Status, input.NextToken, input.Limit)
	if err != nil {
		return nil, nil, fmt.Errorf("failed ListTask: %w", err)
	}
	return tasks, nextToken, nil
}
