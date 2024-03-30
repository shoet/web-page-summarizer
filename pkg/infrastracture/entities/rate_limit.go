package entities

/*
AuthRateLimitはID単位でアクセス回数を表現する構造体
*/
type AuthRateLimit struct {
	ID    string `dynamodbav:"id"`
	Count uint   `dynamodbav:"count"`
}
