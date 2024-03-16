.DEFAULT_GOAL := help

DOCKER_IMAGE := web-page-summarizer-task

.PHONY: up
up: ## Run local server
	docker-compose up -d

.PHONY: logs
logs: ## Show logs
	docker-compose logs -f

.PHONY: down
down: ## Stop local server
	docker-compose down

.PHONY: build
build: ## build go binary to bootstrap
	env GOARCH=amd64 GOOS=linux go build -trimpath -ldflags="-s -w" -o ./.bin/api/bootstrap functions/api/main.go \
	&& zip -j ./.bin/api.zip ./.bin/api/bootstrap \
	&& env GOARCH=amd64 GOOS=linux go build -trimpath -ldflags="-s -w" -o ./.bin/stream-event/bootstrap functions/stream-event/main.go \
	&& zip -j ./.bin/stream-event.zip ./.bin/stream-event/bootstrap \
	&& env GOARCH=amd64 GOOS=linux go build -trimpath -ldflags="-s -w" -o ./.bin/auth/bootstrap functions/auth/main.go \
	&& zip -j ./.bin/auth.zip ./.bin/auth/bootstrap \
	&& env GOARCH=amd64 GOOS=linux go build -trimpath -ldflags="-s -w" -o ./.bin/session/bootstrap functions/session/main.go \
	&& zip -j ./.bin/session.zip ./.bin/session/bootstrap

.PHONY: clean
clean: ## Clean Lambda functions binary
	rm -rf ./.bin

.PHONY: deploy
deploy: clean build ## Deploy by Serverless Framework
	sls deploy --verbose

.PHONY: generate
generate: ## Generate codes
	go generate ./...

.PHONY: help
help: ## Show options
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | \
		awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-20s\033[0m %s\n", $$1, $$2}'
