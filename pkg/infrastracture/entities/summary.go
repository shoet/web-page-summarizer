package entities

import (
	"encoding/json"

	"github.com/rs/zerolog"
)

type Summary struct {
	Id               string `json:"id" dynamodbav:"id"`
	TaskStatus       string `json:"taskStatus" dynamodbav:"task_status,omitempty"`
	PageUrl          string `json:"pageUrl" dynamodbav:"page_url,omitempty"`
	Title            string `json:"title,omitempty" dynamodbav:"title,omitempty"`
	Content          string `json:"content,omitempty" dynamodbav:"content,omitempty"`
	UserId           string `json:"userId,omitempty" dynamodbav:"user_id,omitempty"`
	Summary          string `json:"summary,omitempty" dynamodbav:"summary,omitempty"`
	TaskFailedReason string `json:"taskFailedReason,omitempty" dynamodbav:"task_failed_reason,omitempty"`
	CreatedAt        int64  `json:"createdAt" dynamodbav:"created_at,omitempty"`
}

func (s Summary) MarshalZerologObject(e *zerolog.Event) {
	e.Str("id", s.Id).
		Str("taskStatus", s.TaskStatus).
		Str("taskFailedReason", s.TaskFailedReason).
		Int64("createdAt", s.CreatedAt)
}

type Task struct {
	Id         uint   `json:"id" db:"id" goqu:"skipinsert"`
	TaskId     string `json:"taskId" db:"task_id"`
	TaskStatus string `json:"taskStatus" db:"task_status"`
	PageUrl    string `json:"pageUrl" db:"page_url"`
	Title      string `json:"title" db:"title"`
	CreatedAt  uint   `json:"createdAt" db:"created_at"`
	UpdatedAt  uint   `json:"updatedAt" db:"updated_at"`
}

func (t *Task) JSON() string {
	b, err := json.MarshalIndent(t, "", "  ")
	if err != nil {
		panic(err)
	}
	return string(b)
}

type Tasks []*Task

func (t Tasks) JSON() string {
	b, err := json.MarshalIndent(t, "", "  ")
	if err != nil {
		panic(err)
	}
	return string(b)
}
