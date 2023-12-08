package main

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"os"

	awsConfig "github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/shoet/web-page-summarizer-task/pkg/chatgpt"
	"github.com/shoet/web-page-summarizer-task/pkg/crawler"
	"github.com/shoet/web-page-summarizer-task/pkg/task"
	"github.com/shoet/webpagesummary/config"
	"github.com/shoet/webpagesummary/entities"
	"github.com/shoet/webpagesummary/queue"
	"github.com/shoet/webpagesummary/repository"
	"golang.org/x/sync/errgroup"
)

func FailExit(err error) {
	fmt.Printf("failed to execute: %v\n", err)
	os.Exit(1)
}

func FailTask(traceId string, err error) {
	fmt.Printf("failed to execute task: %v\n", err)
}

func FetchTaskId(ctx context.Context, q *queue.QueueClient, maxExecute int) ([]string, error) {
	var tasks []string
	for i := 0; i < maxExecute; i++ {
		taskId, err := q.Dequeue(ctx)
		if err != nil {
			if errors.Is(err, queue.ErrEmptyQueue) {
				break
			}
			return nil, fmt.Errorf("failed to dequeue: %w", err)
		}
		tasks = append(tasks, taskId)
	}
	return tasks, nil
}

func main() {
	ctx := context.Background()

	cfg, err := config.NewConfig()
	if err != nil {
		FailExit(err)
	}

	awsCfg, err := awsConfig.LoadDefaultConfig(ctx)
	if err != nil {
		FailExit(err)
	}

	db := dynamodb.NewFromConfig(awsCfg)
	summaryRepository := repository.NewSummaryRepository(db)

	pageCrawler, err := crawler.NewPageCrawler(&crawler.PageCrawlerInput{
		BrowserPath: cfg.BrowserPath,
	})
	if err != nil {
		FailExit(err)
	}

	client := &http.Client{}
	chatgptService := chatgpt.NewChatGPTService(cfg.OpenAIApiKey, client)

	// dequeue taskId from sqs
	queueClient := queue.NewQueueClient(awsCfg, cfg.QueueUrl)
	tasks, err := FetchTaskId(ctx, queueClient, cfg.MaxTaskExecute)
	if err != nil {
		FailExit(err)
	}

	tasker := task.NewSummaryTask(summaryRepository, pageCrawler, chatgptService)

	var eg errgroup.Group
	for _, task := range tasks {
		t := task
		eg.Go(func() error {
			err := tasker.ExecuteSummaryTask(ctx, t)
			if err != nil {
				// if failed task update dynamodb status failed
				if err := summaryRepository.UpdateSummary(ctx, &entities.Summary{
					Id:         t,
					TaskStatus: "failed",
				}); err != nil {
					return fmt.Errorf("failed to update summary [%s]: %w", t, err)
				}
			}
			return nil
		})
	}

	if err := eg.Wait(); err != nil {
		FailExit(err)
	}
}
