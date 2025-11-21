PACKAGE=github.com/String-Reconciliation-Ditributed-System/RCDS_GO
MAIN_PACKAGE=$(PACKAGE)/cmd

GO_BUILD_ARGS=CGO_ENABLED=0 GO111MODULE=on

.PHONY: all
all: fmt vet lint test build

.PHONY: build
build: 
	$(GO_BUILD_ARGS) go build -o bin/rcds $(MAIN_PACKAGE)

.PHONY: test
test:
	$(GO_BUILD_ARGS) go test ./pkg/... -v

.PHONY: unit-test
unit-test: test

.PHONY: test-coverage
test-coverage:
	$(GO_BUILD_ARGS) go test ./pkg/... -coverprofile=coverage.out -covermode=atomic
	go tool cover -html=coverage.out -o coverage.html

.PHONY: fmt
fmt:
	$(GO_BUILD_ARGS) gofmt -w -s .

.PHONY: vet
vet:
	$(GO_BUILD_ARGS) go vet ./...

.PHONY: lint
lint:
	@which golangci-lint > /dev/null || (echo "golangci-lint not found, skipping lint" && exit 0)
	golangci-lint run ./...

.PHONY: vendor
vendor:
	$(GO_BUILD_ARGS) go mod vendor

.PHONY: vendor-fmt
vendor-fmt: fmt vendor
	$(GO_BUILD_ARGS) go mod tidy

.PHONY: clean
clean:
	rm -rf bin/ coverage.out coverage.html vendor/

.PHONY: help
help:
	@echo "Available targets:"
	@echo "  all           - Run fmt, vet, lint, test, and build"
	@echo "  build         - Build the binary"
	@echo "  test          - Run all tests"
	@echo "  test-coverage - Run tests with coverage report"
	@echo "  fmt           - Format code"
	@echo "  vet           - Run go vet"
	@echo "  lint          - Run golangci-lint"
	@echo "  vendor        - Vendor dependencies"
	@echo "  clean         - Clean build artifacts"
