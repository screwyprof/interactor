# This repo's root import path.
PKG := gitlab.com/screwyprof/golibs

## DO NOT EDIT BELLOW THIS LINE
GO_FILES = $(shell find . -name "*.go" | uniq)
GO_PACKAGES = $(shell go list ./... | tr '\n', ',')
LOCAL_PACKAGES="github.com/screwyprof/"

SHELL := bash

OK_COLOR=\033[32;01m
NO_COLOR=\033[0m
MAKE_COLOR=\033[33;01m%-20s\033[0m

all: deps fmt test ## download deps, format code, run tests

deps: ## sync go mod deps
	@echo -e "$(OK_COLOR)--> Installing dev dependencies and tools$(NO_COLOR)"
	go install mvdan.cc/gofumpt@v0.3.1
	go install github.com/daixiang0/gci@v0.8.0

fmt: ## format go files
	@echo -e "$(OK_COLOR)--> Formatting go files$(NO_COLOR)"
	@go mod tidy
	@go fmt ./...
	@gofumpt -l -w .
	@gci write $(GO_FILES) -s standard  -s default -s "prefix($(LOCAL_PACKAGES))"

lint: ## run linters for current changes
	@echo -e "$(OK_COLOR)==> Linting current changes$(NO_COLOR)"
	golangci-lint run ./...

test: ## run all tests
	@echo -e "$(OK_COLOR)--> Running unit tests$(NO_COLOR)"
	go test -v --race --count=1 ./...

help: ## show this help screen
	@awk 'BEGIN {FS = ":.*?## "} /^[a-zA-Z_-]+:.*?## / {sub("\\\\n",sprintf("\n%22c"," "), $$2);printf "$(MAKE_COLOR)  %s\n", $$1, $$2}' $(MAKEFILE_LIST)

# To avoid unintended conflicts with file names, always add to .PHONY
# unless there is a reason not to.
# https://www.gnu.org/software/make/manual/html_node/Phony-Targets.html
.PHONY: all deps test fmt clean help