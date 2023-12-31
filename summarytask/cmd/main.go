package main

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"os"
	"strings"
	"sync"
	"time"

	awsConfig "github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/shoet/web-page-summarizer-task/pkg/chatgpt"
	"github.com/shoet/web-page-summarizer-task/pkg/crawler"
	"github.com/shoet/web-page-summarizer-task/pkg/task"
	"github.com/shoet/webpagesummary/config"
	"github.com/shoet/webpagesummary/entities"
	"github.com/shoet/webpagesummary/logging"
	"github.com/shoet/webpagesummary/queue"
	"github.com/shoet/webpagesummary/repository"
)

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

var TraceIdKey interface{}

func main() {
	ctx := context.Background()
	logger := logging.NewLogger(os.Stdout)

	cfg, err := config.NewConfig()
	if err != nil {
		logger.Fatal("failed to load config", err)
	}

	awsCfg, err := awsConfig.LoadDefaultConfig(ctx)
	if err != nil {
		logger.Fatal("failed to load aws config", err)
	}

	db := dynamodb.NewFromConfig(awsCfg)
	summaryRepository := repository.NewSummaryRepository(db)

	pageCrawler, err := crawler.NewPageCrawler(&crawler.PageCrawlerInput{
		BrowserPath: cfg.BrowserPath,
	})
	if err != nil {
		logger.Fatal("failed to initialize page crawler", err)
	}

	client := &http.Client{}
	chatgptService, err := chatgpt.NewChatGPTService(cfg.OpenAIApiKey, client)
	if err != nil {
		logger.Fatal("failed to initialize chatgpt service", err)
	}

	tasker := task.NewSummaryTask(summaryRepository, pageCrawler, chatgptService)

	queueClient := queue.NewQueueClient(awsCfg, cfg.QueueUrl)

	// TODO: graceful shutdown
	for {
		// sqs long polling
		tasks, err := FetchTaskId(ctx, queueClient, cfg.MaxTaskExecute)
		if err != nil {
			logger.Error("failed to fetch task", err)
			continue
		}

		if len(tasks) == 0 {
			logger.Info("no task")
			continue
		}

		logger.Info("Pull task:")
		logger.Info(strings.Join(tasks, "\n"))

		var wg sync.WaitGroup
		for _, t := range tasks {
			wg.Add(1)

			go func(taskId string) {
				defer wg.Done()
				traceIdLogger := logger.NewTraceIdLogger(taskId)
				ctx := logging.SetLogger(ctx, traceIdLogger)
				err := withExecTimeout(ctx, func(ctx context.Context) error {
					return tasker.ExecuteSummaryTask(ctx, taskId)
				}, time.Second*time.Duration(cfg.ExecTimeout))
				if err != nil {
					// タスク失敗時の処理
					if err := summaryRepository.UpdateSummary(context.Background(), &entities.Summary{
						Id:               taskId,
						TaskStatus:       "failed",
						TaskFailedReason: err.Error(),
					}); err != nil {
						traceIdLogger.Error("failed to update summary", err)
					}
					traceIdLogger.Error("task is failed", err)
					return
				}
				traceIdLogger.Info("task is complete")
				return
			}(t)
		}
		wg.Wait()
	}
}

func withExecTimeout(ctx context.Context, fn func(c context.Context) error, duration time.Duration) error {
	ctxTimeout, cancel := context.WithTimeout(ctx, duration)
	defer cancel()
	errCh := make(chan error)
	go func(ctx context.Context) {
		errCh <- fn(ctx)
	}(ctxTimeout)

	select {
	case <-ctxTimeout.Done():
		if ctxTimeout.Err() == context.Canceled {
			return nil
		}
		return fmt.Errorf("task is timeout: %s", ctxTimeout.Err())
	case err := <-errCh:
		return err
	}
}
