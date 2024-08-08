.PHONY: lint test mock-gen

MOCKS_DIR=./testutil/mocks

OPFGD_PKG := github.com/babylonlabs-io/finality-gadget/cmd/opfgd

mock-gen:
	go install go.uber.org/mock/mockgen@latest
	mockgen -source=sdk/client/expected_clients.go -package mocks -destination $(MOCKS_DIR)/expected_clients_mock.go
	mockgen -source=sdk/client/interface.go -package mocks -destination $(MOCKS_DIR)/sdkclient_mock.go

test:
	go test -race ./... -v

lint:
	golangci-lint run

install:
	go install -trimpath $(OPFGD_PKG)