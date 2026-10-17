package interactor

import (
	"context"
	"fmt"
	"maps"
	"reflect"
	"sync"
	"sync/atomic"

	"github.com/screwyprof/interactor/v3/internal/nilcheck"
)

// Dispatcher routes a request to the use case registered for its type.
//
// The zero value is ready to use. A Dispatcher is safe for concurrent use and must not be copied
// after first use.
type Dispatcher struct {
	mu       sync.Mutex                             // serialises Register
	useCases atomic.Pointer[map[reflect.Type]entry] // read lock-free; replaced, never modified
}

// entry keeps the use case with its response type, so a mismatched Dispatch can say what was
// registered.
type entry struct {
	useCase  any
	response reflect.Type
}

// Register adds uc as the use case for requests of type Req.
//
// Request types match exactly: Req and *Req are different types. For a nil use case the error
// wraps [ErrNilUseCase], and if Req already has one it wraps [ErrAlreadyRegistered]; both name
// the request type.
//
// Register is meant for wiring at startup: each call copies the registrations, so that Dispatch
// can read them without locking.
func (d *Dispatcher) Register[Req, Res any](uc UseCase[Req, Res]) error {
	reqType := reflect.TypeFor[Req]()

	if nilcheck.IsNil(uc) {
		return fmt.Errorf("%w: %v", ErrNilUseCase, reqType)
	}

	d.mu.Lock()
	defer d.mu.Unlock()

	var current map[reflect.Type]entry
	if p := d.useCases.Load(); p != nil {
		current = *p
	}

	if _, ok := current[reqType]; ok {
		return fmt.Errorf("%w: %v", ErrAlreadyRegistered, reqType)
	}

	// Copy on write: dispatching reads the published map without a lock, so it is never modified.
	// Sized to the existing entries: the new one may grow it once, which at startup costs nothing
	// measurable, and leaves no size arithmetic for a mutant to change without a test noticing.
	next := make(map[reflect.Type]entry, len(current))
	maps.Copy(next, current)
	next[reqType] = entry{useCase: uc, response: reflect.TypeFor[Res]()}

	d.useCases.Store(&next)

	return nil
}

// Dispatch runs the use case registered for req's type and returns its response.
//
// Res comes first so callers name only the response type and the request type is inferred:
//
//	greeting, err := d.Dispatch[Greeting](ctx, Hello{})
//
// The use case is found by Req, the request's static type: a request held in an interface
// variable is looked up as that interface type, not its dynamic type. Res must be exactly the
// registered use case's response type.
//
// It returns [ErrNilContext] for a nil ctx, [ErrUseCaseNotFound] if nothing is registered for
// Req, and a [*ResponseTypeError] if the registered use case returns a type other than Res.
// Errors from the use case itself are returned unchanged, including any it gets from dispatching
// further requests, together with whatever response it returned alongside the error.
func (d *Dispatcher) Dispatch[Res, Req any](ctx context.Context, req Req) (Res, error) {
	var zero Res

	if ctx == nil {
		return zero, ErrNilContext
	}

	reqType := reflect.TypeFor[Req]()

	var (
		e  entry
		ok bool
	)
	if m := d.useCases.Load(); m != nil {
		e, ok = (*m)[reqType]
	}

	if !ok {
		return zero, fmt.Errorf("%w: %v", ErrUseCaseNotFound, reqType)
	}

	uc, ok := e.useCase.(UseCase[Req, Res])
	if !ok {
		return zero, &ResponseTypeError{Request: reqType, Registered: e.response, Requested: reflect.TypeFor[Res]()}
	}

	return uc.Run(ctx, req) //nolint:wrapcheck // the use case's own error, returned unchanged as documented.
}

// MustRegister is like [Dispatcher.Register] but panics if registration fails. Use it when wiring
// an application at startup, where a duplicate or nil use case is a programming mistake.
func (d *Dispatcher) MustRegister[Req, Res any](uc UseCase[Req, Res]) {
	if err := d.Register(uc); err != nil {
		panic(err)
	}
}
