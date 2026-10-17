# Interactor

[![codecov](https://codecov.io/gh/screwyprof/interactor/graph/badge.svg)](https://codecov.io/gh/screwyprof/interactor) [![Go](https://github.com/screwyprof/interactor/actions/workflows/go.yml/badge.svg)](https://github.com/screwyprof/interactor/actions/workflows/go.yml) [![Vulnerabilities](https://github.com/screwyprof/interactor/actions/workflows/vulns.yml/badge.svg)](https://github.com/screwyprof/interactor/actions/workflows/vulns.yml) [![Release](https://img.shields.io/github/v/release/screwyprof/interactor)](https://github.com/screwyprof/interactor/releases/latest) [![Go Reference](https://pkg.go.dev/badge/github.com/screwyprof/interactor/v3.svg)](https://pkg.go.dev/github.com/screwyprof/interactor/v3) [![License](https://img.shields.io/github/license/screwyprof/interactor)](LICENSE)

`Interactor` runs use cases and dispatches requests to them by request type. It is inspired by Uncle Bob's [Clean Architecture](https://blog.cleancoder.com/uncle-bob/2012/08/13/the-clean-architecture.html) and the interactor concept described there.

A use case is any type with a `Run` method from a request to a response. The compiler checks the types; there is no reflection on the call path.

## Installation

Requires Go 1.27 or later.

```bash
go get github.com/screwyprof/interactor/v3@latest
```

## Usage

```go
package main

import (
	"context"
	"fmt"
	"log"

	"github.com/screwyprof/interactor/v3"
)

type Hello struct{ Name string }

type Greeting struct{ Text string }

// Greet is a use case: a type with a Run method from a request to a response.
type Greet struct{}

func (Greet) Run(ctx context.Context, req Hello) (Greeting, error) {
	return Greeting{Text: "Hello, " + req.Name + "!"}, nil
}

func main() {
	var d interactor.Dispatcher
	d.MustRegister(Greet{})

	greeting, err := d.Dispatch[Greeting](context.Background(), Hello{Name: "Gopher"})
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println(greeting.Text) // Hello, Gopher!
}
```

- `MustRegister` panics on a wiring mistake, which is what you want at startup; `Register` returns the error instead.
- `Dispatch` names only the response type; the request type is inferred from the argument.
- A plain function becomes a use case with `interactor.UseCaseFunc[Hello, Greeting](fn)`.
- The zero `Dispatcher` is ready to use and safe for concurrent use; don't copy it after first use.
- The use case is found by the request's static type, so pass the concrete request type, not one held in `any`.
- Request types match exactly: `Hello` and `*Hello` are different types.

### Dispatcher or a single use case?

The `Dispatcher` gives a component one dependency through which it runs many use cases:

```go
type Handler struct {
	uc *interactor.Dispatcher
}

func (h Handler) Checkout(ctx context.Context, cart Cart) (Receipt, error) {
	order, err := h.uc.Dispatch[Order](ctx, PlaceOrder{Cart: cart})
	if err != nil {
		return Receipt{}, err
	}

	return h.uc.Dispatch[Receipt](ctx, ChargeOrder{OrderID: order.ID})
}
```

In tests, register fake use cases on a real `Dispatcher`; no mock is needed.

A component that runs a single use case can depend on it directly, since `UseCase` is a plain interface: `greet interactor.UseCase[Hello, Greeting]`, then `h.greet.Run(ctx, req)`.

### Errors

| Error | When |
|---|---|
| `ErrAlreadyRegistered` | `Register` for a request type that already has a use case (`MustRegister` panics with it) |
| `ErrNilUseCase` | `Register` with a nil use case (`MustRegister` panics with it) |
| `ErrUseCaseNotFound` | `Dispatch` for a request type nothing is registered for |
| `*ResponseTypeError` | `Dispatch` asks for a response type other than the registered one (match with `errors.AsType[*interactor.ResponseTypeError]`) |
| `ErrNilContext` | `Dispatch` with a nil context |

Errors returned by the use case itself come back unchanged.

## Testing your use cases

Use cases are plain types, so any testing style works. The `testdsl` package is there if you want each test to read as a scenario: given this use case, when it runs with this request, then expect this.

```go
import "github.com/screwyprof/interactor/v3/testdsl"

func TestGreet(t *testing.T) {
	testdsl.Test(t).
		Given(Greet{}).
		When(Hello{Name: "Gopher"}).
		Then(Greeting{Text: "Hello, Gopher!"})
}
```

`Given` fixes the use case's request and response types for the rest of the chain, so `When` only accepts its request and `Then` only its response: `Then(42)` would not compile.

Expect an error with `ThenFailWith`, which matches like `errors.Is`. Say `Greet` rejects an empty name with `ErrNoName`:

```go
func TestGreetRejectsAnEmptyName(t *testing.T) {
	testdsl.Test(t).
		Given(Greet{}).
		When(Hello{}).
		ThenFailWith(ErrNoName)
}
```

For anything else, `ThenCheck` hands you the typed result:

```go
func TestGreetMentionsTheName(t *testing.T) {
	testdsl.Test(t).
		Given(Greet{}).
		When(Hello{Name: "Gopher"}).
		ThenCheck(func(tb testing.TB, got Greeting, err error) {
			require.NoError(tb, err)
			assert.Contains(tb, got.Text, "Gopher")
		})
}
```

Scenarios fit table-driven tests too:

```go
func TestGreetScenarios(t *testing.T) {
	tests := map[string]struct {
		when Hello
		then func(*testdsl.When[Greeting])
	}{
		"greets by name": {
			when: Hello{Name: "Gopher"},
			then: func(w *testdsl.When[Greeting]) { w.Then(Greeting{Text: "Hello, Gopher!"}) },
		},
		"rejects empty names": {
			when: Hello{},
			then: func(w *testdsl.When[Greeting]) { w.ThenFailWith(ErrNoName) },
		},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			tc.then(testdsl.Test(t).Given(Greet{}).When(tc.when))
		})
	}
}
```

`When` runs the use case under the test's context (`t.Context()`). `testdsl` uses [testify](https://github.com/stretchr/testify), so your tests depend on it too.

## Migrating from v2

v3 replaces the reflection-based API with generics. Most of what v2 checked at runtime is now a compile error.

| v2 | v3 |
|---|---|
| `Run(ctx, req Req, res *Res) error` | `Run(ctx, req Req) (Res, error)`, or keep it and wrap with `FromOutParam` |
| `UseCaseRunner`, `UseCaseRunnerFn` | `UseCase[Req, Res]`, `UseCaseFunc[Req, Res]` |
| code depending on `UseCaseRunner` to accept the dispatcher | `*interactor.Dispatcher`, or an interface you define for the use case you need: a generic method can't satisfy an interface, so the dispatcher no longer implements one |
| `interactor.NewDispatcher()` | `var d interactor.Dispatcher` |
| `d.Register(Req{}, interactor.MustAdapt(uc))` | `d.MustRegister(uc)` once `Run` has the v3 shape, or `d.MustRegister(interactor.FromOutParam(uc.Run))` unchanged |
| `d.Register(Req{}, interactor.Must(interactor.Func(fn)))` | `d.MustRegister(interactor.FromOutParam(fn))` unchanged, or `interactor.UseCaseFunc[Req, Res](fn)` once `fn` returns its response |
| `var res Res; err := d.Run(ctx, req, &res)` | `res, err := d.Dispatch[Res](ctx, req)` |
| a request held in `any` was routed by its dynamic type | routing uses the static type: pass the concrete request |
| registering a type twice overwrote the first | `ErrAlreadyRegistered`, and `MustRegister` panics |
| a nil runner failed at `Run` with `ErrUseCaseRunnerNotFound` | `ErrNilUseCase` at `Register` |
| `ErrUseCaseRunnerNotFound` | `ErrUseCaseNotFound` |
| `ErrResultTypeMismatch` | `*ResponseTypeError` |
| `Request`, `Response`, `Func`, `Adapt`, `Must`, `MustAdapt` and their signature errors | removed: the compiler checks what they checked at runtime |

To move one use case at a time, wrap a v2-style one with `FromOutParam`; it registers unchanged:

```go
d.MustRegister(interactor.FromOutParam(legacy.Run))
```

## Documentation

[pkg.go.dev/github.com/screwyprof/interactor/v3](https://pkg.go.dev/github.com/screwyprof/interactor/v3)

## Contributing

Contributions are welcome. Run `make check` before opening a pull request: it runs the tests with a 100% coverage floor, the linters, a `go mod tidy` check and `govulncheck`, the same gates as CI.

## License

Released under the [MIT License](LICENSE).
