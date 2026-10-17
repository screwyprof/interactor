package interactor

import (
	"context"
	"fmt"
	"reflect"
	"sync"
)

// Dispatcher routes a request to the use case registered for its type.
//
// The zero value is ready to use. A Dispatcher is safe for concurrent use and must not be copied
// after first use.
type Dispatcher struct {
	mu       sync.RWMutex
	useCases map[reflect.Type]entry
}

// entry keeps the use case with its response type, so a mismatched Dispatch can say what was
// registered.
type entry struct {
	useCase  any
	response reflect.Type
}

// NewDispatcher returns an empty Dispatcher. It is equivalent to new(Dispatcher).
func NewDispatcher() *Dispatcher {
	return &Dispatcher{}
}

// Register adds uc as the use case for requests of type Req.
//
// Request types match exactly: Req and *Req are different types. It returns [ErrNilUseCase] for
// a nil use case and [ErrAlreadyRegistered] if Req already has one.
func (d *Dispatcher) Register[Req, Res any](uc UseCase[Req, Res]) error {
	if isNil(uc) {
		return ErrNilUseCase
	}

	reqType := reflect.TypeFor[Req]()

	d.mu.Lock()
	defer d.mu.Unlock()

	if _, ok := d.useCases[reqType]; ok {
		return fmt.Errorf("%w: %v", ErrAlreadyRegistered, reqType)
	}

	if d.useCases == nil {
		d.useCases = make(map[reflect.Type]entry)
	}

	d.useCases[reqType] = entry{useCase: uc, response: reflect.TypeFor[Res]()}

	return nil
}

// Dispatch runs the use case registered for req's type and returns its response.
//
// Res comes first so callers name only the response type and the request type is inferred:
//
//	greeting, err := d.Dispatch[Greeting](ctx, Hello{})
//
// It returns [ErrNilContext] for a nil ctx, [ErrUseCaseNotFound] if nothing is registered for
// Req, and a [*ResponseTypeError] if the registered use case returns a type other than Res.
// Errors from the use case itself are returned unchanged.
func (d *Dispatcher) Dispatch[Res, Req any](ctx context.Context, req Req) (Res, error) {
	var zero Res

	if ctx == nil {
		return zero, ErrNilContext
	}

	reqType := reflect.TypeFor[Req]()

	d.mu.RLock()
	e, ok := d.useCases[reqType]
	d.mu.RUnlock()

	if !ok {
		return zero, fmt.Errorf("%w: %v", ErrUseCaseNotFound, reqType)
	}

	uc, ok := e.useCase.(UseCase[Req, Res])
	if !ok {
		return zero, &ResponseTypeError{Request: reqType, Registered: e.response, Requested: reflect.TypeFor[Res]()}
	}

	return uc.Run(ctx, req) //nolint:wrapcheck // the use case's own error, returned unchanged as documented.
}

// isNil reports whether uc is nil or holds a nil pointer or func, any of which would panic in Run.
func isNil(uc any) bool {
	if uc == nil {
		return true
	}

	v := reflect.ValueOf(uc)

	return (v.Kind() == reflect.Pointer || v.Kind() == reflect.Func) && v.IsNil()
}
