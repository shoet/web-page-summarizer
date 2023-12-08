package entities

type Summary struct {
	Id               string `json:"id" dynamodbav:"id"`
	TaskStatus       string `json:"taskStatus" dynamodbav:"task_status"`
	PageUrl          string `json:"pageUrl" dynamodbav:"page_url"`
	Title            string `json:"title,omitempty" dynamodbav:"title"`
	Content          string `json:"content,omitempty" dynamodbav:"content"`
	Summary          string `json:"summary" dynamodbav:"summary"`
	TaskFailedReason string `json:"taskFailedReason,omitempty" dynamodbav:"task_failed_reason"`
}
