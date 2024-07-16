.PHONY: lint test mock-gen

CUR_DIR := $(shell pwd)
MOCKS_DIR=$(CUR_DIR)/testutil/mocks

mock-gen:
	go install go.uber.org/mock/mockgen@latest
	mockgen -source=sdk/client/expected_clients.go -package mocks -destination $(MOCKS_DIR)/expected_clients_mock.go
	mockgen -source=sdk/client/interface.go -package mocks -destination $(MOCKS_DIR)/sdkclient_mock.go

test:
	go test -race ./... -v

lint:
	golangci-lint run