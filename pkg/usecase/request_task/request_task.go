package request_task

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/shoet/webpagesummary/pkg/infrastracture/entities"
)

type SummaryRepository interface {
	CreateSummary(ctx context.Context, summary *entities.Summary) (string, error)
}

type QueueClient interface {
	Queue(ctx context.Context, message string) error
}

type Usecase struct {
	SummaryRepository SummaryRepository
	QueueClient       QueueClient
}

func NewUsecase(summaryRepository SummaryRepository, queueClient QueueClient) *Usecase {
	return &Usecase{
		SummaryRepository: summaryRepository,
		QueueClient:       queueClient,
	}
}

func (u *Usecase) Run(ctx context.Context, url string) (taskID string, error error) {
	id := uuid.New().String()
	newSummaryTask := &entities.Summary{
		Id:         id,
		PageUrl:    url,
		TaskStatus: "request",
		CreatedAt:  time.Now().Unix(),
	}
	_, err := u.SummaryRepository.CreateSummary(ctx, newSummaryTask)
	if err != nil {
		return "", err
	}

	// queue taskId to sqs
	if err := u.QueueClient.Queue(ctx, id); err != nil {
		return "", err
	}
	return id, nil
}
