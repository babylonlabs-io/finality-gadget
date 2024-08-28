MOCKS_DIR := ./testutil/mocks

OPFGD_PKG := github.com/babylonlabs-io/finality-gadget/cmd/opfgd

BUILDDIR ?= $(CURDIR)/build
BUILD_FLAGS := --tags '$(BUILD_TAGS)' --ldflags '$(LDFLAGS)'
BUILD_ARGS := $(BUILD_ARGS) -o $(BUILDDIR)

DOCKER ?= $(shell which docker)
GIT_ROOT := $(shell git rev-parse --show-toplevel)

mock-gen:
	go install go.uber.org/mock/mockgen@latest
	mockgen -source=db/interface.go -package mocks -destination $(MOCKS_DIR)/db_mock.go
	mockgen -source=finalitygadget/expected_clients.go -package mocks -destination $(MOCKS_DIR)/expected_clients_mock.go

test:
	go test -race ./... -v

lint:
	golangci-lint run

install:
	go install -trimpath $(OPFGD_PKG)

.PHONY: lint test mock-gen install

$(BUILDDIR)/:
	mkdir -p $(BUILDDIR)/

build: go.sum $(BUILDDIR)/
	CGO_CFLAGS="-O -D__BLST_PORTABLE__" go $@ -mod=readonly $(BUILD_FLAGS) $(BUILD_ARGS) ./...

build-docker:
	$(DOCKER) build --secret id=sshKey,src=${BBN_PRIV_DEPLOY_KEY} \
	--tag babylonlabs-io/finality-gadget \
	-f Dockerfile \
	$(GIT_ROOT)

.PHONY: build build-docker