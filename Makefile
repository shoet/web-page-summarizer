.DEFAULT_GOAL := help

DOCKER_IMAGE := web-page-summarizer-task

.PHONY: build
build: ## build go binary to bootstrap
	env GOARCH=amd64 GOOS=linux go build -trimpath -ldflags="-s -w" -o ./bin/api/bootstrap functions/api/main.go \
	&& zip -j ./bin/api.zip ./bin/api/bootstrap

.PHONY: clean
clean: ## Clean Lambda functions binary
	rm -rf ./bin

.PHONY: deploy
deploy: clean build ## Deploy by Serverless Framework
	sls deploy --verbose

.PHONY: build-tasker
build-tasker: ## Build tasker binary
	cd summarytask && \
		env GOOS=linux go build -trimpath -ldflags="-s -w" -o cmd/bin/main cmd/main.go

.PHONY: build-tasker-local
build-tasker-local: ## Build tasker binary on Arm64
	cd summarytask && \
		go build -trimpath -ldflags="-s -w" -o cmd/bin/main cmd/main.go

.PHONY: build-image-tasker
build-image-tasker: ## Build tasker container image
	docker build -t ${DOCKER_IMAGE}:latest \
		--platform linux/amd64 \
		--target deploy \
		-f summarytask/Dockerfile \
		.

.PHONY: build-image-tasker-local
build-image-tasker-local: ## Build crawler container image on Arm64
	docker build -t ${DOCKER_IMAGE}:local \
		--target deploy \
		-f summarytask/Dockerfile \
		--no-cache \
		.

.PHONY: push-container-tasker
push-container-tasker: ## Push tasker container image
	bash ./container_push.sh

.PHONY: run-tasker
run-tasker: ## run tasker development
	cd summarytask && \
		go run cmd/crawltask/main.go

.PHONY: generate
generate: ## Generate codes
	go generate ./...

.PHONY: help
help: ## Show options
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | \
		awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-20s\033[0m %s\n", $$1, $$2}'
