package task

import (
	"context"
	"fmt"

	"github.com/shoet/web-page-summarizer-task/pkg/chatgpt"
	"github.com/shoet/webpagesummary/pkg/infrastracture/repository"
	"github.com/shoet/webpagesummary/pkg/logging"
)

type Logger interface {
	Info(args ...interface{})
	Error(args ...interface{})
}

type Crawler interface {
	FetchContents(url string) (string, string, error)
}

type SummaryTask struct {
	repo    *repository.SummaryRepository
	crawler Crawler
	chatgpt *chatgpt.ChatGPTService
	logger  Logger
}

func NewSummaryTask(
	repo *repository.SummaryRepository,
	crawler Crawler,
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

	// scrape title, content
	logger.Info("processing scrape contents")
	title, content, err := st.crawler.FetchContents(s.PageUrl)
	if err != nil {
		return fmt.Errorf("failed to scrape body: %w", err)
	}

	// dynamodb update title, content
	logger.Info("update title, content")
	s.Title = title
	s.Content = content
	if err := st.repo.UpdateSummary(ctx, s); err != nil {
		return fmt.Errorf("failed to update summary: %w", err)
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
	if summary == "" {
		return fmt.Errorf("failed to get summary is empty: %w", err)
	}
	s.Summary = summary
	s.TaskStatus = "complete"

	// dynamodb update summary, status complete
	logger.Info("update summary, status complete")
	if err := st.repo.UpdateSummary(ctx, s); err != nil {
		return fmt.Errorf("failed to update summary: %w", err)
	}
	return nil
}
