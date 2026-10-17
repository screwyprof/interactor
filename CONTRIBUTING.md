# Contributing

## What you need

Go 1.27, make, bash and git.

```shell
git clone https://github.com/screwyprof/interactor
cd interactor
make check
```

Every tool the build reaches for is a Go program fetched at its pinned version by `go run`:
golangci-lint, reviewdog, govulncheck and prettycov. Never from your PATH, so everyone runs the same
versions as CI. The first `make check` compiles them, so it is slow once and fast after. Renovate
keeps the pins in the [Makefile](Makefile) up to date.

The git pre-commit hooks are optional. They need `pre-commit`, which is Python rather than Go: run
`pip install pre-commit`, then `pre-commit install`.

## The gates

`make check` is what CI runs, so the two cannot drift. It holds coverage at 100% of every package
and prints one block per gate:

| target | what it asks |
| --- | --- |
| `make test` | tests with `-race` and `-shuffle=on`, writing the coverage profile |
| `make lint-all` | the linters [.golangci.yml](.golangci.yml) enables, over the whole tree |
| `make tidy` | `go mod tidy -diff`, so a stale go.mod or go.sum cannot reach main |
| `make vulns` | govulncheck, which reports only what this module can actually call |
| `make mutate` | mutation testing with gremlins: a mutant that lives is a test that asserts nothing |

All of them are expected to pass before a pull request. `make lint` narrows linting to your
branch, `make lint-fix` applies the linters' automatic fixes, `make coverage` shows the coverage
tree, and `make help` lists the rest.

## What a change looks like

### A fix needs a test that fails without it

Not a test that covers the line, but one you have watched fail with the fix reverted. Write the
test first, run it red, then fix the code. 100% coverage is the floor, not the proof: a test that
runs a line without checking what it does is worse than no test.

### Examples compile

Every Go snippet in the [README](README.md) has to compile against the current API, and the
`Example` functions in the tests are what pkg.go.dev shows, so they run as tests with checked
output. Change the API, change both.

### Comments say what the code cannot

A decision, a measurement, or the bug that forced the shape. Not what the next line plainly does.

### Performance claims come with numbers

Run the benchmarks before and after with `go test -run '^$' -bench . -benchmem -count 10` and
compare them with [benchstat](https://pkg.go.dev/golang.org/x/perf/cmd/benchstat). Paste the
benchstat table into the commit body. A result with `~` is not a result.

## Using an AI assistant

No objection to it. The bar above does not move: you are the author of what you send, so you have
read it, you can say why it is shaped that way, and you have checked it yourself before a reviewer
does. `make check` passing is the floor rather than the case for the change.

## Layout

| | |
| --- | --- |
| root package `interactor` | `UseCase`, `UseCaseFunc`, `Dispatcher` and `FromOutParam` |
| `testdsl` | the public Given/When/Then helpers for testing use cases |
| `internal/nilcheck` | the one definition of a nil use case, shared by both |

## Conventions

Commit subjects say what changed and why it matters, not which files moved. Breaking changes need a
new major version: the module path ends in `/v3`, and a v4 would change it again.
