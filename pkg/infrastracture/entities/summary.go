package entities

import "time"

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

type Task struct {
	Id         uint      `json:"id" db:"id" goqu:"skipinsert"`
	TaskId     string    `json:"taskId" db:"task_id"`
	TaskStatus string    `json:"taskStatus" db:"task_status"`
	PageUrl    string    `json:"pageUrl" db:"page_url"`
	Title      string    `json:"title" db:"title"`
	CreatedAt  time.Time `json:"createdAt" db:"created_at"`
	UpdatedAt  time.Time `json:"updatedAt" db:"updated_at"`
}
