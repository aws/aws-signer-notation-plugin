GO_SPACE := $(CURDIR)
GO_BIN_PATH := $(GO_SPACE)/build/bin
GO_PLUGIN_PATH := $(GO_SPACE)/cmd
RESOURCES := $(GO_SPACE)/internal/resources/notation*
GOARCH := $(shell go env GOARCH)
GOOS := $(shell go env GOOS)
GO_BUILD := CGO_ENABLED=0 go build
PLUGIN_NAME := notation-com.amazonaws.signer.notation.plugin
VERSION := 1.0.0-${GIT_HASH}
export GO_INSTALL_FLAGS := -ldflags "-X github.com/aws/aws-signer-notation-plugin/internal/version.Version=$(VERSION)"
export T := ./cmd/... ./internal/... ./plugin/...
LDFLAGS := -s -w
MOCKGEN_INSTALLED := $(shell which mockgen)

.PHONY: build
build: | generate-mocks
	@echo "Building for $(GOARCH) $(GOOS) agent"
	cd $(GO_PLUGIN_PATH) && GOOS=$(GOOS) GOARCH=$(GOARCH) $(GO_BUILD) $(GO_INSTALL_FLAGS) -o $(GO_BIN_PATH)/$(PLUGIN_NAME)_$(GOOS)_$(GOARCH) $(GO_PLUGIN_PATH)

.PHONY: clean
clean: | clean-mocks
	rm -rf $(GO_BIN_PATH) $(RESOURCES)

.PHONY: generate-mocks
generate-mocks:
ifndef MOCKGEN_INSTALLED
	@echo "Installing mockgen as it is not present in the system..."
	go install github.com/golang/mock/mockgen@v1.6.0
endif
	@echo "Generating Mocks..."
	$(GOPATH)/bin/mockgen -package client -destination=./internal/client/mock_client.go "github.com/aws/aws-signer-notation-plugin/internal/client" Interface
	@echo "Mocks generated successfully."

.PHONY: clean-mocks
clean-mocks:
	rm -rf ./internal/client/mock_client.go

.PHONY: test
test: check-line-endings
	go test -v -race -coverprofile=coverage.txt -covermode=atomic $(T)

.PHONY: check-line-endings
check-line-endings:
	! find . -name "*.go" -type f -exec file "{}" ";" | grep CRLF

.PHONY: fix-line-endings
fix-line-endings:
	find . -type f -name "*.go" -exec sed -i -e "s/\r//g" {} +