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
	"github.com/shoet/webpagesummary/queue"
	"github.com/shoet/webpagesummary/repository"
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
	chatgptService, err := chatgpt.NewChatGPTService(cfg.OpenAIApiKey, client)
	if err != nil {
		FailExit(err)
	}

	tasker := task.NewSummaryTask(summaryRepository, pageCrawler, chatgptService)

	// dequeue taskId from sqs
	queueClient := queue.NewQueueClient(awsCfg, cfg.QueueUrl)
	tasks, err := FetchTaskId(ctx, queueClient, cfg.MaxTaskExecute)
	if err != nil {
		FailExit(err)
	}

	if len(tasks) == 0 {
		fmt.Println("no task")
		return
	}

	fmt.Println("Pull task:")
	fmt.Println(strings.Join(tasks, "\n"))

	var wg sync.WaitGroup
	for _, t := range tasks {
		wg.Add(1)
		go func(taskId string) {
			defer wg.Done()
			// err := tasker.ExecuteSummaryTask(ctx, taskId)
			err := withExecTimeout(func() error {
				return tasker.ExecuteSummaryTask(ctx, taskId)
			}, time.Second*time.Duration(cfg.ExecTimeout))
			if err != nil {
				// タスク失敗時の処理
				if err := summaryRepository.UpdateSummary(context.Background(), &entities.Summary{
					Id:               taskId,
					TaskStatus:       "failed",
					TaskFailedReason: err.Error(),
				}); err != nil {
					fmt.Printf("failed to update summary [%s]: %v\n", taskId, err)
				}
				fmt.Printf("task is failed [%s]: %v\n", taskId, err)
				return
			}
			fmt.Printf("task is complete [%s]\n", taskId)
			return
		}(t)
	}
	wg.Wait()
}

func withExecTimeout(fn func() error, duration time.Duration) error {
	ctxTimeout, cancel := context.WithTimeout(context.Background(), duration)
	defer cancel()
	go func() {
		fn()
		cancel()
	}()
	for {
		time.Sleep(10 * time.Microsecond)
		select {
		case <-ctxTimeout.Done():
			if ctxTimeout.Err() == context.Canceled {
				return nil
			}
			return fmt.Errorf("task is timeout: %s", ctxTimeout.Err())
		}
	}
}
