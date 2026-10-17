GO_FILES = $(shell git ls-files -co --exclude-standard "*.go")

# Tools run with `go run pkg@version`, so the version named here is the one that runs, locally and in CI.
GOLANGCI_VERSION := v2.15.0
REVIEWDOG_VERSION := v0.21.2
GOVULNCHECK_VERSION := v1.9.0
GOLANGCI := go run github.com/golangci/golangci-lint/v2/cmd/golangci-lint@$(GOLANGCI_VERSION)

SHELL := bash
.SHELLFLAGS := -eu -o pipefail -c

OK_COLOR=\033[32;01m
NO_COLOR=\033[0m
MAKE_COLOR=\033[33;01m%-20s\033[0m

all: fmt lint-all test ## format code, run linters and tests

fmt: ## format go files
	@echo -e "$(OK_COLOR)--> Formatting go files$(NO_COLOR)"
	@go mod tidy
	@$(GOLANGCI) fmt $(GO_FILES)

lint: ## run linters for current changes
	@echo -e "$(OK_COLOR)==> Linting current changes$(NO_COLOR)"
	$(GOLANGCI) run ./...

# --new-from-merge-base in .golangci.yml is a separate key from --new-from-rev; both must be cleared.
lint-all: ## run linters over the whole tree, unfiltered
	@echo -e "$(OK_COLOR)==> Linting$(NO_COLOR)"
	$(GOLANGCI) run ./... --new-from-rev="" --new-from-merge-base=""

# Applies the auto-fixes linters offer (modernize, usetesting, wsl_v5, ...) across the whole tree.
lint-fix: ## apply linter auto-fixes
	@echo -e "$(OK_COLOR)==> Applying lint fixes$(NO_COLOR)"
	$(GOLANGCI) run ./... --fix --new-from-rev="" --new-from-merge-base=""

# CI only: renders the findings `lint` reports as annotations on the pull request diff.
lint-annotate:
	@$(GOLANGCI) run ./... --output.text.print-issued-lines=false --output.text.colors=false --show-stats=false \
		| go run github.com/reviewdog/reviewdog/cmd/reviewdog@$(REVIEWDOG_VERSION) \
			-f=golangci-lint -name=golangci-lint -reporter=github-pr-check -fail-level=any

# Reachability-aware: reports only advisories in code this module can actually call.
vulns: ## report known vulnerabilities reachable from this module
	@echo -e "$(OK_COLOR)==> Vulnerabilities$(NO_COLOR)"
	go run golang.org/x/vuln/cmd/govulncheck@$(GOVULNCHECK_VERSION) ./...

test: ## run all tests
	@echo -e "$(OK_COLOR)--> Running unit tests$(NO_COLOR)"
	go test -race -count=1 -shuffle=on -timeout=120s -covermode=atomic -coverpkg=./... -coverprofile=coverage.out ./...

coverage: test ## show test coverage report
	@echo -e "$(OK_COLOR)--> Showing test coverage$(NO_COLOR)"
	go tool cover -func=coverage.out

clean: ## remove build artifacts and generated files
	@echo -e "$(OK_COLOR)--> Cleaning up$(NO_COLOR)"
	rm -f coverage.out

help: ## show this help screen
	@awk 'BEGIN {FS = ":.*?## "} /^[a-zA-Z_-]+:.*?## / {sub("\\\\n",sprintf("\n%22c"," "), $$2);printf "$(MAKE_COLOR)  %s\n", $$1, $$2}' $(MAKEFILE_LIST)

# To avoid unintended conflicts with file names, always add to .PHONY
# unless there is a reason not to.
# https://www.gnu.org/software/make/manual/html_node/Phony-Targets.html
.PHONY: all fmt lint lint-all lint-fix lint-annotate vulns test coverage clean help
