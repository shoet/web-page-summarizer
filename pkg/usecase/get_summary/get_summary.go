package get_summary

import (
	"context"
	"fmt"

	"github.com/shoet/webpagesummary/pkg/infrastracture/entities"
	"github.com/shoet/webpagesummary/pkg/util"
)

type SummaryRepository interface {
	GetSummary(ctx context.Context, id string, userId *string) (*entities.Summary, error)
}

type Usecase struct {
	SummaryRepository SummaryRepository
}

func NewUsecase(summaryRepository SummaryRepository) *Usecase {
	return &Usecase{SummaryRepository: summaryRepository}
}

func (u *Usecase) Run(ctx context.Context, taskId string) (*entities.Summary, error) {
	var userIdPtr *string
	userId, err := util.GetUserSub(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed GetUserSub: %w", err)
	}
	if userId != util.APIKeyUserSub {
		userIdPtr = &userId
	}
	summary, err := u.SummaryRepository.GetSummary(ctx, taskId, userIdPtr)
	if err != nil {
		return nil, fmt.Errorf("failed get summary: %w", err)
	}
	return summary, nil
}
