.PHONY: all

VERSION := $(shell cz version -p)
COMMIT := $(shell git rev-parse --short HEAD)
TAG ?= latest

# Run unit tests with coverage
unittest:
	@echo "Running unit tests with coverage for v$(VERSION)-$(COMMIT)"
	@go test -timeout 5s -covermode atomic -coverprofile unittest.coverage.out -v ./...
	@go tool cover -html=unittest.coverage.out -o unittest.coverage.html

# Run the free@home CLI locally
cli-run-local:
	@echo "Starting free@home CLI v$(VERSION)-$(COMMIT)"
	@go run -ldflags "-X 'main.version=$(VERSION)' -X 'main.commit=$(COMMIT)'" ./cmd/cli/main.go

# Build the free@home CLI
cli-build:
	@echo "Building free@home CLI v$(VERSION)-$(COMMIT)"
	@go build -ldflags "-X 'main.version=$(VERSION)' -X 'main.commit=$(COMMIT)'" -o fh ./cmd/cli/main.go
	@chmod +x fh

# Build the Docker image for the free@home CLI
cli-build-docker:
	@echo "Building Docker image for free@home CLI v$(VERSION) from $(COMMIT) with tag $(TAG)."
	@docker build --build-arg version=$(VERSION) --build-arg commit=$(COMMIT) -t ghcr.io/lion-and-bear/freeathome-cli:${TAG} -f ./cmd/cli.dockerfile .

# Build the multi-arch Docker image for the free@home CLI
cli-build-docker-multiarch:
	@echo "Building multi-arch Docker image for free@home CLI v$(VERSION) from $(COMMIT) with tag $(TAG)."
	@docker buildx build --platform linux/amd64,linux/arm64 --build-arg version=$(VERSION) --build-arg commit=$(COMMIT) -t ghcr.io/lion-and-bear/freeathome-cli:${TAG} -f ./cmd/cli.dockerfile --push .

# Run the free@home CLI integration tests
cli-integration-test:
	@echo "Running integration tests for free@home CLI v$(VERSION)-$(COMMIT)"
	@rm -rf ./coverage-cli && mkdir -p ./coverage-cli
	@go build -covermode atomic -coverpkg=./... -o cli-integration.test ./cmd/cli
	@go test -timeout 5s -tags integration ./integration/
	@go tool covdata textfmt -i coverage-cli -o cli-integration.coverage.out
	@go tool cover -html=cli-integration.coverage.out -o cli-integration.coverage.html

# Run all unit and integration tests and aggregate the coverage reports
test-ci: unittest cli-integration-test
	@$(shell go env GOPATH)/bin/gocovmerge unittest.coverage.out cli-integration.coverage.out > coverage.out
	@go tool cover -html=coverage.out -o coverage.html
