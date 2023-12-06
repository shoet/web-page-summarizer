package entities

type Summary struct {
	Id      string `json:"id" dynamodbav:"id"`
	Status  string `json:"status" dynamodbav:"status"`
	Url     string `json:"url" dynamodbav:"url"`
	Title   string `json:"title" dynamodbav:"title"`
	Content string `json:"content" dynamodbav:"content"`
	Summary string `json:"summary" dynamodbav:"summary"`
}
