# Copyright Amazon.com Inc. or its affiliates. All Rights Reserved.
#
# Licensed under the Apache License, Version 2.0 (the "License"). You may
# not use this file except in compliance with the License. A copy of the
# License is located at
#
# 	http://aws.amazon.com/apache2.0
#
# or in the "license" file accompanying this file. This file is distributed
# on an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either
# express or implied. See the License for the specific language governing
# permissions and limitations under the License.

.PHONY: help
help:
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | sort | awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-25s\033[0m %s\n", $$1, $$2}'

.PHONY: build
build: test ## build the aws signer notation plugin
	go build -o ./build/bin/notation-com.amazonaws.signer.notation.plugin ./cmd

.PHONY: test
test: generate-mocks ## run the unit tests
	go test -v -race -coverprofile=coverage.txt -covermode=atomic ./...

.PHONY: generate-mocks
generate-mocks: ## generate mocks required for unit tests
	@if ! command -v mockgen &> /dev/null; then \
		echo "Installing mockgen as it is not present in the system..."; \
		go install github.com/golang/mock/mockgen@v1.6.0; \
	fi
	@echo "Generating Mocks..."
	mockgen -package client -destination=./internal/client/mock_client.go "github.com/aws/aws-signer-notation-plugin/internal/client" Interface
	@echo "Mocks generated successfully."

.PHONY: clean
clean: ## remove build artifacts and mocks
	rm -rf ./internal/client/mock_client.go
	rm -rf ./build
	git status --ignored --short | grep '^!! ' | sed 's/!! //' | xargs rm -rf
