BINARY_NAME := komodor
BIN_DIR     := bin
CMD_PATH    := ./cmd/komodor

.PHONY: all generate build lint test test-integration clean

all: generate build

generate:
	$(shell go env GOPATH)/bin/oapi-codegen --generate types,client --package client --o internal/client/client.gen.go api/komodor-clean.yaml

build:
	go build -o $(BIN_DIR)/$(BINARY_NAME) $(CMD_PATH)

lint:
	golangci-lint run ./...

test:
	go test ./...

test-integration:
	go test -tags integration ./...

clean:
	rm -rf $(BIN_DIR)
