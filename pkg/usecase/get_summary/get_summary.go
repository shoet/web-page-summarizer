package get_summary

import (
	"context"
	"fmt"

	"github.com/shoet/webpagesummary/pkg/infrastracture/entities"
)

type SummaryRepository interface {
	GetSummary(ctx context.Context, id string) (*entities.Summary, error)
}

type Usecase struct {
	SummaryRepository SummaryRepository
}

func NewUsecase(summaryRepository SummaryRepository) *Usecase {
	return &Usecase{SummaryRepository: summaryRepository}
}

func (u *Usecase) Run(ctx context.Context, taskId string) (*entities.Summary, error) {
	summary, err := u.SummaryRepository.GetSummary(ctx, taskId)
	if err != nil {
		return nil, fmt.Errorf("failed get summary: %w", err)
	}
	return summary, nil
}
