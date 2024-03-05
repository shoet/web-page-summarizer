package task

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"testing"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
	"github.com/shoet/web-page-summarizer-task/pkg/chatgpt"
	"github.com/shoet/web-page-summarizer-task/pkg/crawler"
	"github.com/shoet/webpagesummary/pkg/infrastracture/entities"
	"github.com/shoet/webpagesummary/pkg/infrastracture/queue"
	"github.com/shoet/webpagesummary/pkg/infrastracture/repository"
	"github.com/shoet/webpagesummary/pkg/testutil"
)

func TestMain(m *testing.M) {
	setupOutput := MustSetup()
	m.Run()
	MustTeardown(&TeardownInput{
		QueueUrl: setupOutput.QueueUrl,
	})
}

var tableName = "web_page_summary"

var queueName = "test-queue"
var queueUrl = fmt.Sprintf(
	"http://sqs.ap-northeast-1.localhost.localstack.cloud:4566/000000000000/%s", queueName)

type SetUpOutput struct {
	QueueUrl string
}

type TeardownInput struct {
	QueueUrl string
}

func MustSetup() *SetUpOutput {
	fmt.Println("setup")
	ctx := context.Background()
	awsConfig, err := testutil.NewAwsConfigWithLocalStack(ctx)
	if err != nil {
		panic(fmt.Sprintf("failed load aws config: %s\n", err.Error()))
	}
	if err := testutil.CreateDynamoDBForTest(ctx, *awsConfig, &dynamodb.CreateTableInput{
		TableName: &tableName,
		AttributeDefinitions: []types.AttributeDefinition{
			{
				AttributeName: aws.String("id"),
				AttributeType: types.ScalarAttributeTypeS,
			},
		},
		KeySchema: []types.KeySchemaElement{
			{
				AttributeName: aws.String("id"),
				KeyType:       types.KeyTypeHash,
			},
		},
		ProvisionedThroughput: &types.ProvisionedThroughput{
			ReadCapacityUnits:  aws.Int64(10),
			WriteCapacityUnits: aws.Int64(10),
		},
	}); err != nil {
		panic(fmt.Sprintf("failed create dynamodb: %s\n", err.Error()))
	}

	queueUrl, err := testutil.CreateSQSStandardQueueForTest(ctx, *awsConfig, queueName)
	if err != nil {
		panic(fmt.Sprintf("failed create sqs queue: %s\n", err.Error()))
	}
	fmt.Println(queueUrl)
	return &SetUpOutput{
		QueueUrl: queueUrl,
	}
}

func MustTeardown(input *TeardownInput) {
	fmt.Println("teardown")
	fmt.Println(input.QueueUrl)
	ctx := context.Background()
	awsConfig, err := testutil.NewAwsConfigWithLocalStack(ctx)
	if err != nil {
		panic(fmt.Sprintf("failed load aws config: %s\n", err.Error()))
	}
	if err := testutil.DropDynamoDBForTest(ctx, *awsConfig, tableName); err != nil {
		panic(fmt.Sprintf("failed delete sqs queue: %s\n", err.Error()))
	}

	if err := testutil.DeleteSQSQueueForTest(ctx, *awsConfig, input.QueueUrl); err != nil {
		panic(fmt.Sprintf("failed delete sqs queue: %s\n", err.Error()))
	}
}

func Prepare_ExecuteSummaryTask(t *testing.T, ctx context.Context, cfg aws.Config) string {
	t.Helper()
	taskId := "test_Test_ExecuteSummaryTask"
	queueClient := queue.NewQueueClient(cfg, queueUrl)
	if err := queueClient.Queue(ctx, taskId); err != nil {
		t.Fatalf("failed to queue: %v", err)
	}

	repo := repository.NewSummaryRepository(dynamodb.NewFromConfig(cfg))
	_, err := repo.CreateSummary(ctx, &entities.Summary{
		Id:         taskId,
		PageUrl:    "https://news.yahoo.co.jp/pickup/6484213",
		TaskStatus: "request",
		CreatedAt:  time.Now().Unix(),
	})
	if err != nil {
		t.Fatalf("failed to create summary: %v", err)
	}
	return taskId
}

func Test_ExecuteSummaryTask(t *testing.T) {
	fmt.Println("Test_ExecuteSummaryTask")
	ctx := context.Background()

	testAwsCfg, err := testutil.NewAwsConfigForTest(t, ctx)
	if err != nil {
		t.Fatalf("failed load aws config: %s\n", err.Error())
	}

	taskId := Prepare_ExecuteSummaryTask(t, ctx, *testAwsCfg)

	db := dynamodb.NewFromConfig(*testAwsCfg)
	pageRepository := repository.NewSummaryRepository(db)

	pageCrawler, err := crawler.NewPageCrawler(&crawler.PageCrawlerInput{
		BrowserPath: "/opt/homebrew/bin/chromium", // TODO local
	})
	if err != nil {
		t.Fatalf("failed to create PageCrawler: %v", err)
	}

	apiKey, ok := os.LookupEnv("CHATGPT_API_SECRET")
	if !ok {
		t.Fatalf("failed to get api key")
	}
	client := &http.Client{}
	chatgptApi, err := chatgpt.NewChatGPTService(apiKey, client)
	if err != nil {
		t.Fatalf("failed to create ChatGPTService: %v", err)
	}

	sut := NewSummaryTask(pageRepository, pageCrawler, chatgptApi)
	if err := sut.ExecuteSummaryTask(ctx, taskId); err != nil {
		t.Fatalf("failed to execute summary task: %v", err)
	}
}
