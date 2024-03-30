package main

import (
	"context"
	"errors"
	"fmt"
	"log"
	"net/http"
	"os"
	"runtime"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	awsConfig "github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/joho/godotenv"
	cp "github.com/otiai10/copy"
	"github.com/shoet/web-page-summarizer-task/pkg/chatgpt"
	"github.com/shoet/web-page-summarizer-task/pkg/crawler"
	"github.com/shoet/web-page-summarizer-task/pkg/task"
	"github.com/shoet/webpagesummary/pkg/config"
	"github.com/shoet/webpagesummary/pkg/infrastracture/adapter"
	"github.com/shoet/webpagesummary/pkg/infrastracture/entities"
	"github.com/shoet/webpagesummary/pkg/infrastracture/repository"
	"github.com/shoet/webpagesummary/pkg/logging"
)

func FailTask(traceId string, err error) {
	fmt.Printf("failed to execute task: %v\n", err)
}

var TraceIdKey interface{}

func CopyBrowser() (string, error) {
	src := "/var/playwright/browser/chromium-1091"
	dst := "/tmp/playwright/browser/chromium-1091"

	if _, err := os.Stat(dst); os.IsNotExist(err) {
		if err := cp.Copy(src, dst); err != nil {
			return "", fmt.Errorf("could not copy browser: %v", err)
		}
	}
	return dst, nil
}

type TaskExecutor struct {
	config            *config.Config
	logger            *logging.Logger
	queue             *adapter.QueueClient
	summaryRepository *repository.SummaryRepository
}

func NewTaskExecutor(ctx context.Context, cfg *config.Config) (*TaskExecutor, error) {
	logger := logging.NewLogger(os.Stdout)
	awsCfg, err := awsConfig.LoadDefaultConfig(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to load aws config: %w", err)
	}
	queueClient := adapter.NewQueueClient(awsCfg, cfg.QueueUrl)
	db := dynamodb.NewFromConfig(awsCfg)
	summaryRepository := repository.NewSummaryRepository(db, &cfg.Env)
	return &TaskExecutor{
		config:            cfg,
		logger:            logger,
		queue:             queueClient,
		summaryRepository: summaryRepository,
	}, nil
}

func (t *TaskExecutor) FetchTaskId(ctx context.Context, maxExecute int) ([]string, error) {
	var tasks []string
	for i := 0; i < maxExecute; i++ {
		taskId, err := t.queue.Dequeue(ctx)
		if err != nil {
			if errors.Is(err, adapter.ErrEmptyQueue) {
				break
			}
			return nil, fmt.Errorf("failed to dequeue: %w", err)
		}
		tasks = append(tasks, taskId)
	}
	return tasks, nil
}

type RunTaskInput struct {
	TaskId           string
	SQSReceiptHandle string
}

func (t *TaskExecutor) RunTask(ctx context.Context, input *RunTaskInput) error {
	playwrightConfig := &crawler.PlaywrightClientConfig{
		BrowserLaunchTimeoutSec: 120,
		SkipInstallBrowsers:     false,
	}
	if runtime.GOOS == "linux" {
		// Lambdaでの実行時は/varに用意したブラウザを/tmpにコピーする
		if _, err := CopyBrowser(); err != nil {
			t.logger.Fatal("failed to copy browser", err)
		}
		// Lambdaでの実行時はブラウザのインストールをスキップする
		playwrightConfig.SkipInstallBrowsers = true
	} else {
		os.Setenv("PLAYWRIGHT_BROWSERS_PATH", t.config.BrowserDownloadPath)
	}
	pageCrawler, browserCloser, err := crawler.NewPlaywrightClient(playwrightConfig)
	defer browserCloser()
	if err != nil {
		return fmt.Errorf("failed to initialize playwright client: %w", err)
	}

	client := &http.Client{}
	chatgptService, err := chatgpt.NewChatGPTService(t.config.OpenAIApiKey, client)
	if err != nil {
		t.logger.Fatal("failed to initialize chatgpt service", err)
	}

	tasker := task.NewSummaryTask(t.summaryRepository, pageCrawler, chatgptService)

	traceIdLogger := t.logger.NewTraceIdLogger(input.TaskId)
	ctx = logging.SetLogger(ctx, traceIdLogger)
	if err := tasker.ExecuteSummaryTask(ctx, input.TaskId); err != nil {
		traceIdLogger.Error("failed to execute task", err)
		// タスク失敗時はqueueから削除する
		input := &RunTaskInput{
			TaskId:           input.TaskId,
			SQSReceiptHandle: input.SQSReceiptHandle,
		}
		if err := t.queue.DeleteMessage(ctx, input.SQSReceiptHandle); err != nil {
			traceIdLogger.Error("failed to delete queue", err)
		}
		// タスク失敗時はsummaryのstatusをfailedにする
		if err := t.summaryRepository.UpdateSummary(context.Background(), &entities.Summary{
			Id:               input.TaskId,
			TaskStatus:       "failed",
			TaskFailedReason: err.Error(),
		}); err != nil {
			traceIdLogger.Error("failed to update summary", err)
		}
		return fmt.Errorf("failed to execute task: %w", err)
	}
	traceIdLogger.Info("task is complete")
	return nil

}

var executor *TaskExecutor

func init() {
	if err := godotenv.Load(); err != nil {
		fmt.Printf("load env: %v\n", err)
	}
	cfg, err := config.NewConfig()
	if err != nil {
		log.Fatalf("failed to load config: %v", err)
	}
	executor, err = NewTaskExecutor(context.Background(), cfg)
	if err != nil {
		log.Fatalf("failed to initialize task executor: %v", err)
	}
}

func Handler(ctx context.Context, sqsEvent events.SQSEvent) error {
	fmt.Println("start handler")
	taskId := sqsEvent.Records[0].Body
	sqsReceiptHandle := sqsEvent.Records[0].ReceiptHandle
	input := &RunTaskInput{
		TaskId:           taskId,
		SQSReceiptHandle: sqsReceiptHandle,
	}
	if err := executor.RunTask(ctx, input); err != nil {
		return fmt.Errorf("failed to execute task: %w", err)
	}
	return nil
}

func main() {
	ctx := context.Background()
	if os.Getenv("ENV") == "local" {
		tasks, err := executor.FetchTaskId(ctx, 1)
		if err != nil {
			log.Fatalf("failed to fetch task: %v", err)
		}

		if len(tasks) == 0 {
			fmt.Println("no task")
			return
		}

		input := &RunTaskInput{
			TaskId:           tasks[0],
			SQSReceiptHandle: "", // TODO
		}
		if err := executor.RunTask(ctx, input); err != nil {
			log.Fatalf("failed to run task: %v", err)
		}
	} else {
		lambda.Start(Handler)
	}
}
