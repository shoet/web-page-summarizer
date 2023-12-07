package task

import (
	"context"
	"fmt"

	"github.com/shoet/web-page-summarizer-task/pkg/chatgpt"
	"github.com/shoet/web-page-summarizer-task/pkg/crawler"
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
	// get task from dynamodb
	s, err := st.repo.GetSummary(ctx, taskId)
	if err != nil {
		return fmt.Errorf("failed to get summary: %w", err)
	}
	fmt.Println("get summary: ")
	fmt.Println(s.Id, s.PageUrl)

	if s.PageUrl == "" {
		return fmt.Errorf("pageurl is empty")
	}

	// dynamodb update taskId status to processing
	s.TaskStatus = "processing"
	if err := st.repo.UpdateSummary(ctx, s); err != nil {
		return fmt.Errorf("failed to update summary: %w", err)
	}

	// scrape title, content
	title, content, err := st.crawler.FetchContents(s.PageUrl)
	if err != nil {
		return fmt.Errorf("failed to scrape body: %w", err)
	}

	// dynamodb update title, content
	s.Title = title
	s.Content = content
	if err := st.repo.UpdateSummary(ctx, s); err != nil {
		return fmt.Errorf("failed to update summary: %w", err)
	}

	// request chatgpt api get content summary
	summaryTemplate, err := chatgpt.SummaryTemplateBuilder(&chatgpt.SummaryTemplateInput{
		Title:   title,
		Content: content,
	})
	if err != nil {
		return fmt.Errorf("failed to build summary template: %w", err)
	}
	summary, err := st.chatgpt.ChatCompletions(&chatgpt.ChatCompletionsInput{
		Text: summaryTemplate,
	})
	s.Summary = summary
	s.TaskStatus = "complete"

	// dynamodb update summary, status complete
	if err := st.repo.UpdateSummary(ctx, s); err != nil {
		return fmt.Errorf("failed to update summary: %w", err)
	}
	return nil
}
