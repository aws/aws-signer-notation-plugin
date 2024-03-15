BASE_DIR := $(dir $(realpath -s $(firstword $(MAKEFILE_LIST))))

.PHONY: build
build: | generate-mocks
	go build -o $(BASE_DIR)/build/bin/notation-com.amazonaws.signer.notation.plugin $(BASE_DIR)/cmd

.PHONY: generate-mocks
generate-mocks:
	@if ! command -v mockgen &> /dev/null; then \
		echo "Installing mockgen as it is not present in the system..."; \
		go install github.com/golang/mock/mockgen@v1.6.0; \
	fi
	@echo "Generating Mocks..."
	$(GOPATH)/bin/mockgen -package client -destination=./internal/client/mock_client.go "github.com/aws/aws-signer-notation-plugin/internal/client" Interface
	@echo "Mocks generated successfully."

.PHONY: clean-mocks
clean-mocks:
	rm -rf ./internal/client/mock_client.go

.PHONY: test
test: check-line-endings
	go test -v -race -coverprofile=coverage.txt -covermode=atomic ./...

.PHONY: check-line-endings
check-line-endings:
	! find . -name "*.go" -type f -exec file "{}" ";" | grep CRLF

.PHONY: fix-line-endings
fix-line-endings:
	find . -type f -name "*.go" -exec sed -i -e "s/\r//g" {} +
