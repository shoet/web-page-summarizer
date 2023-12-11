package task

import (
	"context"
	"fmt"

	"github.com/shoet/web-page-summarizer-task/pkg/chatgpt"
	"github.com/shoet/web-page-summarizer-task/pkg/crawler"
	"github.com/shoet/webpagesummary/logging"
	"github.com/shoet/webpagesummary/repository"
)

type Logger interface {
	Info(args ...interface{})
	Error(args ...interface{})
}

type SummaryTask struct {
	repo    *repository.SummaryRepository
	crawler *crawler.PageCrawler
	chatgpt *chatgpt.ChatGPTService
	logger  Logger
}

func NewSummaryTask(
	repo *repository.SummaryRepository,
	crawler *crawler.PageCrawler,
	chatgpt *chatgpt.ChatGPTService,
) *SummaryTask {
	return &SummaryTask{
		repo:    repo,
		crawler: crawler,
		chatgpt: chatgpt,
	}
}

func (st *SummaryTask) ExecuteSummaryTask(ctx context.Context, taskId string) error {
	logger := logging.GetLogger(ctx)
	logger.Info("start to execute task")

	// get task from dynamodb
	logger.Info("get task from dynamodb")
	s, err := st.repo.GetSummary(ctx, taskId)
	if err != nil {
		return fmt.Errorf("failed to get summary: %w", err)
	}

	if s.PageUrl == "" {
		return fmt.Errorf("pageurl is empty")
	}

	// dynamodb update taskId status to processing
	s.TaskStatus = "processing"
	logger.Info("task is processing")
	if err := st.repo.UpdateSummary(ctx, s); err != nil {
		return fmt.Errorf("failed to update summary: %w", err)
	}

	if err := checkContextTimeout(ctx); err != nil {
		return fmt.Errorf("context timeout: %w", err)
	}

	// scrape title, content
	logger.Info("processing scrape contents")
	title, content, err := st.crawler.FetchContents(s.PageUrl)
	if err != nil {
		return fmt.Errorf("failed to scrape body: %w", err)
	}

	if err := checkContextTimeout(ctx); err != nil {
		return fmt.Errorf("context timeout: %w", err)
	}

	// dynamodb update title, content
	logger.Info("update title, content")
	s.Title = title
	s.Content = content
	if err := st.repo.UpdateSummary(ctx, s); err != nil {
		return fmt.Errorf("failed to update summary: %w", err)
	}

	if err := checkContextTimeout(ctx); err != nil {
		return fmt.Errorf("context timeout: %w", err)
	}

	// request chatgpt api get content summary
	logger.Info("processing text summary")
	summaryTemplate, err := chatgpt.SummaryTemplateBuilder(&chatgpt.SummaryTemplateInput{
		Title:   title,
		Content: content,
	})
	if err != nil {
		return fmt.Errorf("failed to build summary template: %w", err)
	}

	logger.Info("request chatgpt api")
	summary, err := st.chatgpt.ChatCompletions(&chatgpt.ChatCompletionsInput{
		Text: summaryTemplate,
	})
	s.Summary = summary
	s.TaskStatus = "complete"

	if err := checkContextTimeout(ctx); err != nil {
		return fmt.Errorf("context timeout: %w", err)
	}

	// dynamodb update summary, status complete
	logger.Info("update summary, status complete")
	if err := st.repo.UpdateSummary(ctx, s); err != nil {
		return fmt.Errorf("failed to update summary: %w", err)
	}
	return nil
}

func checkContextTimeout(ctx context.Context) error {
	select {
	case <-ctx.Done():
		return ctx.Err()
	default:
		return nil
	}
}
