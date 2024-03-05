package entities

type Summary struct {
	Id               string `json:"id" dynamodbav:"id"`
	TaskStatus       string `json:"taskStatus" dynamodbav:"task_status,omitempty"`
	PageUrl          string `json:"pageUrl" dynamodbav:"page_url,omitempty"`
	Title            string `json:"title,omitempty" dynamodbav:"title,omitempty"`
	Content          string `json:"content,omitempty" dynamodbav:"content,omitempty"`
	Summary          string `json:"summary,omitempty" dynamodbav:"summary,omitempty"`
	TaskFailedReason string `json:"taskFailedReason,omitempty" dynamodbav:"task_failed_reason,omitempty"`
	CreatedAt        int64  `json:"createdAt" dynamodbav:"created_at,omitempty"`
}
