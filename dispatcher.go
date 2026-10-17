package interactor

import (
	"context"
	"fmt"
	"maps"
	"reflect"
	"sync"
	"sync/atomic"
)

// Dispatcher routes a request to the use case registered for its type.
//
// The zero value is ready to use. A Dispatcher is safe for concurrent use and must not be copied
// after first use.
type Dispatcher struct {
	mu       sync.Mutex                           // serialises Register
	useCases atomic.Pointer[map[reflect.Type]any] // read lock-free; replaced, never modified
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

	if isNil(uc) {
		return fmt.Errorf("%w: %v", ErrNilUseCase, reqType)
	}

	d.mu.Lock()
	defer d.mu.Unlock()

	var current map[reflect.Type]any
	if p := d.useCases.Load(); p != nil {
		current = *p
	}

	if _, ok := current[reqType]; ok {
		return fmt.Errorf("%w: %v", ErrAlreadyRegistered, reqType)
	}

	// Copy on write: dispatching reads the published map without a lock, so it is never modified.
	// Sized to the existing entries: the new one may grow it once, which at startup costs nothing
	// measurable.
	next := make(map[reflect.Type]any, len(current))
	maps.Copy(next, current)
	next[reqType] = uc

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
// It returns an error wrapping [ErrUseCaseNotFound] if nothing is registered for Req, or
// [ErrResponseTypeMismatch] if the registered use case returns a type other than Res. ctx is passed
// to the use case as given. Errors from the use case itself are returned unchanged, including any
// it gets from dispatching further requests, together with whatever response it returned alongside
// the error.
func (d *Dispatcher) Dispatch[Res, Req any](ctx context.Context, req Req) (Res, error) {
	var zero Res

	reqType := reflect.TypeFor[Req]()

	var (
		registered any
		ok         bool
	)
	if m := d.useCases.Load(); m != nil {
		registered, ok = (*m)[reqType]
	}

	if !ok {
		return zero, fmt.Errorf("%w: %v", ErrUseCaseNotFound, reqType)
	}

	uc, ok := registered.(UseCase[Req, Res])
	if !ok {
		return zero, fmt.Errorf("%w: the use case for %v does not return %v",
			ErrResponseTypeMismatch, reqType, reflect.TypeFor[Res]())
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

// isNil reports whether uc is a nil interface, nil pointer or nil func. These are rejected at
// registration rather than left to fail inside Dispatch.
func isNil(uc any) bool {
	if uc == nil {
		return true
	}

	v := reflect.ValueOf(uc)

	return (v.Kind() == reflect.Pointer || v.Kind() == reflect.Func) && v.IsNil()
}
