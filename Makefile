# Asked of git rather than found on disk: tracked files plus untracked ones that are not ignored,
# so a new file counts before it is added, but nothing else the working tree holds.
GIT_LS := git -c core.quotePath=false ls-files -co --exclude-standard
# wildcard drops tracked files deleted but not yet staged. Expanded only by the targets that use it.
GO_FILES = $(wildcard $(shell $(GIT_LS) "*.go"))
LOCAL_PACKAGES := github.com/screwyprof/interactor/v3
COVERAGE := coverage.out
# The same figure as codecov.yml's project target: one is what a contributor sees before pushing,
# the other is what blocks the merge.
COVERAGE_FLOOR := 100

# Every tool is fetched by `go run pkg@version` at the version named here, so this block is the
# whole toolchain. The `# renovate:` lines let the bot read a Makefile it would otherwise ignore.
# renovate: datasource=go depName=github.com/golangci/golangci-lint/v2
GOLANGCI_VERSION := v2.15.0
# renovate: datasource=go depName=github.com/reviewdog/reviewdog
REVIEWDOG_VERSION := v0.21.2
# renovate: datasource=go depName=golang.org/x/vuln
GOVULNCHECK_VERSION := v1.9.0
# renovate: datasource=go depName=github.com/screwyprof/prettycov
PRETTYCOV_VERSION := v0.16.0
# renovate: datasource=go depName=github.com/go-gremlins/gremlins
GREMLINS_VERSION := v0.6.0

GOLANGCI := go run github.com/golangci/golangci-lint/v2/cmd/golangci-lint@$(GOLANGCI_VERSION)
PRETTYCOV := go run github.com/screwyprof/prettycov/cmd/prettycov@$(PRETTYCOV_VERSION)

# -o pipefail fails a recipe when any command in a pipe fails, not just the last one.
SHELL := bash
.SHELLFLAGS := -eu -o pipefail -c

# go test writes the profile even when the suite fails; without this a partial one survives.
.DELETE_ON_ERROR:

OK_COLOR=\033[32;01m
NO_COLOR=\033[0m
MAKE_COLOR=\033[33;01m%-20s\033[0m

all: fmt check ## format code, then run every gate

fmt: ## format go files
	@echo -e "$(OK_COLOR)==> Formatting$(NO_COLOR)"
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

# -diff reports what tidy would change without writing it, so a failing gate leaves no dirty tree.
tidy: ## check go.mod and go.sum are what the imports say
	@echo -e "$(OK_COLOR)==> Checking go.mod$(NO_COLOR)"
	@go mod tidy -diff

# Reachability-aware: reports only advisories in code this module can actually call.
vulns: ## report known vulnerabilities reachable from this module
	@echo -e "$(OK_COLOR)==> Vulnerabilities$(NO_COLOR)"
	@go run golang.org/x/vuln/cmd/govulncheck@$(GOVULNCHECK_VERSION) ./...

# Coverage says a line ran; mutation testing says whether a test would notice it being wrong.
# gremlins flips conditions and arithmetic, then counts the mutants. The gate reads its counts, not
# its percentages or thresholds (v0.6.0 ignores its --threshold-* flags and fails a run that exactly
# meets one): every mutant must be killed, none lived, none uncovered and none timed out, since a
# timed-out mutant counts as neither killed nor lived.
#
# gremlins works on a copy of the module root, so only the files GIT_LS lists go in, NUL-separated
# so any file name survives, skipping any deleted but not yet staged.
#
# --timeout-coefficient: the default limit is a small multiple of the suite's own runtime, and
# back-to-back runs then timed out every dispatcher mutant; at 30 none do.
mutate: ## run mutation testing; fails unless every mutant is killed
	@echo -e "$(OK_COLOR)==> Mutation testing$(NO_COLOR)"
	@tmp=$$(mktemp -d) && trap 'rm -rf "$$tmp"' EXIT; \
	 $(GIT_LS) -z | while IFS= read -r -d '' f; do if [ -e "$$f" ]; then printf '%s\0' "$$f"; fi; done \
		| tar --null -cf - -T - | (cd "$$tmp" && tar -xf -); \
	 out=$$(cd "$$tmp" && go run github.com/go-gremlins/gremlins/cmd/gremlins@$(GREMLINS_VERSION) \
		unleash --timeout-coefficient=30 . 2>&1) || { echo "$$out"; exit 1; }; \
	 echo "$$out"; \
	 echo "$$out" | grep -qE '^Killed: [1-9][0-9]*, Lived: 0, Not covered: 0$$' \
		&& echo "$$out" | grep -qE '^Timed out: 0,' \
		|| { echo "mutation testing: every mutant must be killed"; exit 1; }

test: ## run all tests and write the coverage profile
	@echo -e "$(OK_COLOR)==> Running tests$(NO_COLOR)"
	go test -race -count=1 -shuffle=on -timeout=120s -covermode=atomic -coverpkg=./... -coverprofile=$(COVERAGE) ./...

coverage: test ## show the coverage tree
	@$(PRETTYCOV) report --profile=$(COVERAGE) --old=$(LOCAL_PACKAGES) --new=interactor --depth=max --files

# Prints $(2) when the gate passes, the whole output when it fails.
define summarise
out=$$($(MAKE) --no-print-directory $(1) 2>&1) || { echo "$$out"; exit 1; }; \
	echo "$$out" | grep -E '$(2)' || { echo "$$out"; exit 1; }
endef

# Every gate CI runs, one line per result, so a pull request quotes the tools rather than a retyping
# of them. The coverage line asserts the same floor codecov.yml enforces on the merge.
check: ## run every quality gate
	@echo -e "$(OK_COLOR)==> Checking$(NO_COLOR)" >&2
	@echo '$$ make test'
	@$(MAKE) --no-print-directory test >/dev/null
	@$(PRETTYCOV) report --profile=$(COVERAGE) --old=$(LOCAL_PACKAGES) --new=interactor \
		--depth=max --files --hide-covered --fail-under=$(COVERAGE_FLOOR)
	@echo; echo '$$ make lint-all'
	@$(call summarise,lint-all,^[0-9]+ issues\.)
	@echo; echo '$$ make tidy'
	@$(MAKE) --no-print-directory tidy && echo "go.mod and go.sum are tidy"
	@echo; echo '$$ make vulns'
	@$(call summarise,vulns,No vulnerabilities|Vulnerability #)
	@echo; echo '$$ make mutate'
	@$(call summarise,mutate,^(Killed:|Timed out:))

clean: ## remove build artifacts and generated files
	@echo -e "$(OK_COLOR)==> Cleaning up$(NO_COLOR)"
	rm -f $(COVERAGE)

help: ## show this help screen
	@awk 'BEGIN {FS = ":.*?## "} /^[a-zA-Z_-]+:.*?## / {sub("\\\\n",sprintf("\n%22c"," "), $$2);printf "$(MAKE_COLOR)  %s\n", $$1, $$2}' $(MAKEFILE_LIST)

# To avoid unintended conflicts with file names, always add to .PHONY
# unless there is a reason not to.
# https://www.gnu.org/software/make/manual/html_node/Phony-Targets.html
.PHONY: all fmt lint lint-all lint-fix lint-annotate tidy vulns mutate test coverage check clean help
