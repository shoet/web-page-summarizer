package list_task

import (
	"context"

	"github.com/shoet/webpagesummary/pkg/infrastracture/entities"
)

type SummaryRepository interface{}

type Usecase struct {
	SummaryRepository SummaryRepository
}

func NewUsecase(summaryRepository SummaryRepository) *Usecase {
	return &Usecase{SummaryRepository: summaryRepository}
}

type UsecaseInput struct {
	Status     *string
	PageOffset *int
	PageLimit  *int
}

func (u *Usecase) Run(ctx context.Context, input UsecaseInput) ([]*entities.Summary, error) {
	return nil, nil
}
