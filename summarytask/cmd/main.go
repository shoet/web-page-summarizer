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
	"github.com/shoet/webpagesummary/config"
	"github.com/shoet/webpagesummary/entities"
	"github.com/shoet/webpagesummary/logging"
	"github.com/shoet/webpagesummary/queue"
	"github.com/shoet/webpagesummary/repository"
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
	config *config.Config
	logger *logging.Logger
	queue  *queue.QueueClient
	db     *dynamodb.Client
}

func NewTaskExecutor(ctx context.Context, cfg *config.Config) (*TaskExecutor, error) {
	logger := logging.NewLogger(os.Stdout)
	awsCfg, err := awsConfig.LoadDefaultConfig(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to load aws config: %w", err)
	}
	queueClient := queue.NewQueueClient(awsCfg, cfg.QueueUrl)
	db := dynamodb.NewFromConfig(awsCfg)
	return &TaskExecutor{
		config: cfg,
		logger: logger,
		queue:  queueClient,
		db:     db,
	}, nil
}

func (t *TaskExecutor) FetchTaskId(ctx context.Context, maxExecute int) ([]string, error) {
	var tasks []string
	for i := 0; i < maxExecute; i++ {
		taskId, err := t.queue.Dequeue(ctx)
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

func (t *TaskExecutor) RunTask(ctx context.Context, taskId string) (string, error) {
	summaryRepository := repository.NewSummaryRepository(t.db)

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
		t.logger.Fatal("failed to initialize page crawler", err)
	}

	client := &http.Client{}
	chatgptService, err := chatgpt.NewChatGPTService(t.config.OpenAIApiKey, client)
	if err != nil {
		t.logger.Fatal("failed to initialize chatgpt service", err)
	}

	tasker := task.NewSummaryTask(summaryRepository, pageCrawler, chatgptService)

	traceIdLogger := t.logger.NewTraceIdLogger(taskId)
	ctx = logging.SetLogger(ctx, traceIdLogger)
	if err := tasker.ExecuteSummaryTask(ctx, taskId); err != nil {
		traceIdLogger.Error("failed to execute task", err)
		// タスク失敗時はsummaryのstatusをfailedにする
		if err := summaryRepository.UpdateSummary(context.Background(), &entities.Summary{
			Id:               taskId,
			TaskStatus:       "failed",
			TaskFailedReason: err.Error(),
		}); err != nil {
			traceIdLogger.Error("failed to update summary", err)
		}
		traceIdLogger.Error("task is failed", err)
		return "failed", fmt.Errorf("failed to execute task: %w", err)
	}
	traceIdLogger.Info("task is complete")
	return "success", nil

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

func Handler(ctx context.Context, sqsEvent events.SQSEvent) (string, error) {
	fmt.Println("start handler")
	taskId := sqsEvent.Records[0].Body
	return executor.RunTask(ctx, taskId)
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

		result, err := executor.RunTask(ctx, tasks[0])
		if err != nil {
			fmt.Println(err)
		}
		fmt.Println(result)
	} else {
		lambda.Start(Handler)
	}
}
