GO_VERSION := 1.23

LOCAL_BIN:=$(CURDIR)/bin/

setup: install-golangci-lint

install-golangci-lint:
	GOBIN=$(LOCAL_BIN) go install github.com/golangci/golangci-lint/cmd/golangci-lint@v1.53.3
lint:
	GOBIN=$(LOCAL_BIN) golangci-lint run ./... --config .golangci.yml
format:
	test -z $$(go fmt ./...)
test:
	go test -v -cover -race ./...

.PHONY: lint format test