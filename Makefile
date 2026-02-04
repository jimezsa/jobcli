TOOLS_DIR := .tools
GOBIN := $(abspath $(TOOLS_DIR))
export GOBIN

.DEFAULT_GOAL := build

GOFILES := $(shell find . -name '*.go' -not -path './.tools/*' -not -path './vendor/*')

.PHONY: build tools fmt fmt-check lint test

build:
	@go build -o jobcli ./cmd/jobcli

tools:
	@mkdir -p $(TOOLS_DIR)
	@go install mvdan.cc/gofumpt@v0.7.0
	@go install golang.org/x/tools/cmd/goimports@v0.38.0
	@go install github.com/golangci/golangci-lint/cmd/golangci-lint@v1.62.2

fmt:
	@$(GOBIN)/goimports -w $(GOFILES)
	@$(GOBIN)/gofumpt -w $(GOFILES)

fmt-check:
	@$(GOBIN)/goimports -w $(GOFILES)
	@$(GOBIN)/gofumpt -w $(GOFILES)
	@git diff --exit-code -- '*.go' go.mod go.sum

lint:
	@$(GOBIN)/golangci-lint run ./...

test:
	@go test ./...
