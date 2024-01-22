package server

import (
	"context"

	"github.com/shoet/webpagesummary/entities"
)

type SummaryRepository interface {
	GetSummary(ctx context.Context, id string) (*entities.Summary, error)
	CreateSummary(ctx context.Context, summary *entities.Summary) (string, error)
}

type QueueClient interface {
	Queue(ctx context.Context, message string) error
}
