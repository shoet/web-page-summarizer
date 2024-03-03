.DEFAULT_GOAL := help

DOCKER_IMAGE := web-page-summarizer-task

.PHONY: build
build: ## build go binary to bootstrap
	env GOARCH=amd64 GOOS=linux go build -trimpath -ldflags="-s -w" -o ./.bin/api/bootstrap functions/api/main.go \
	&& zip -j ./.bin/api.zip ./.bin/api/bootstrap

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
