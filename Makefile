PACKAGE=github.com/String-Reconciliation-Ditributed-System/RCDS_GO
MAIN_PACKAGE=$(PACKAGE)/cmd

GO_BUILD_ARGS=CGO_ENABLED=0 GO111MODULE=on


.PHONY: build
build: 
	$(GO_BUILD_ARGS) go build -o bin/rcds $(MAIN_PACKAGE)

.PHONY: unit-test
unit-test:
	$(GO_BUILD_ARGS) go test -mod=vendor ./pkg/... -v

.PHONY: vendor-fmt
vendor-fmt:
	$(GO_BUILD_ARGS) gofmt -w -s . && go mod vendor && go mod tidy && go vet ./...
