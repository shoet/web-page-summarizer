package entities

type Summary struct {
	Id               string `json:"id" dynamodbav:"id"`
	TaskStatus       string `json:"taskStatus" dynamodbav:"task_status"`
	PageUrl          string `json:"pageUrl" dynamodbav:"page_url"`
	Title            string `json:"title" dynamodbav:"title"`
	Content          string `json:"content" dynamodbav:"content"`
	Summary          string `json:"summary" dynamodbav:"summary"`
	TaskFailedReason string `json:"taskFailedReason" dynamodbav:"task_failed_reason"`
}
