package queue

import (
	"context"
	"fmt"
	"testing"

	"github.com/shoet/webpagesummary/testutil"
)

func TestMain(m *testing.M) {
	setupOutput := MustSetup()
	m.Run()
	MustTeardown(&TeardownInput{QueueUrl: setupOutput.QueueUrl})
}

var queueName = "test-queue2"
var queueUrl = fmt.Sprintf(
	"http://sqs.ap-northeast-1.localhost.localstack.cloud:4566/000000000000/%s", queueName)

type SetUpOutput struct {
	QueueUrl string
}

type TeardownInput struct {
	QueueUrl string
}

func MustSetup() *SetUpOutput {
	ctx := context.Background()
	awsConfig, err := testutil.NewAwsConfigWithLocalStack(ctx)
	if err != nil {
		panic(fmt.Sprintf("failed load aws config: %s\n", err.Error()))
	}
	queueUrl, err := testutil.CreateSQSStandardQueueForTest(ctx, *awsConfig, queueName)
	if err != nil {
		panic(fmt.Sprintf("failed create sqs queue: %s\n", err.Error()))
	}
	return &SetUpOutput{
		QueueUrl: queueUrl,
	}
}

func MustTeardown(input *TeardownInput) {
	ctx := context.Background()
	awsConfig, err := testutil.NewAwsConfigWithLocalStack(ctx)
	if err != nil {
		panic(fmt.Sprintf("failed load aws config: %s\n", err.Error()))
	}
	if err := testutil.DeleteSQSQueueForTest(ctx, *awsConfig, input.QueueUrl); err != nil {
		panic(fmt.Sprintf("failed delete sqs queue: %s\n", err.Error()))
	}
}

func Test_QueueClient_Queue(t *testing.T) {
	// TODO
	ctx := context.Background()

	testAwsCfg, err := testutil.NewAwsConfigForTest(t, ctx)
	if err != nil {
		t.Fatalf("failed load aws config: %s\n", err.Error())
	}

	sut := NewQueueClient(*testAwsCfg, queueUrl)
	if err := sut.Queue(ctx, "test"); err != nil {
		t.Fatalf("failed Queue: %s\n", err.Error())
	}
}

func Test_QueueClient_Dequeue(t *testing.T) {
	// TODO
	ctx := context.Background()

	testAwsCfg, err := testutil.NewAwsConfigForTest(t, ctx)
	if err != nil {
		t.Fatalf("failed load aws config: %s\n", err.Error())
	}

	sut := NewQueueClient(*testAwsCfg, queueUrl)
	msg, err := sut.Dequeue(ctx)
	if err != nil {
		t.Fatalf("failed Dequeue: %s\n", err.Error())
	}
	fmt.Println(msg)
}

func Test_QueueClient_QueueDequeue(t *testing.T) {
	// TODO
	ctx := context.Background()
	testAwsCfg, err := testutil.NewAwsConfigForTest(t, ctx)
	if err != nil {
		t.Fatalf("failed load aws config: %s\n", err.Error())
	}

	sut := NewQueueClient(*testAwsCfg, queueUrl)

	sut.Queue(ctx, "test1")
	sut.Queue(ctx, "test2")
	sut.Queue(ctx, "test3")
	sut.Queue(ctx, "test4")

	for i := 0; i < 4; i++ {
		m, err := sut.Dequeue(ctx)
		if err != nil {
			t.Fatalf("failed Dequeue: %s\n", err.Error())
		}
		fmt.Println(m)
	}
}
