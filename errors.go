package interactor

import (
	"errors"
	"fmt"
	"reflect"
)

// Errors returned by [Dispatcher.Register] and [Dispatcher.Dispatch].
var (
	ErrUseCaseNotFound   = errors.New("interactor: no use case registered for request type")
	ErrAlreadyRegistered = errors.New("interactor: use case already registered for request type")
	ErrNilUseCase        = errors.New("interactor: use case must not be nil")
	ErrNilContext        = errors.New("interactor: context must not be nil")
)

// ResponseTypeError is returned by [Dispatcher.Dispatch] when the use case registered for the request type
// returns a different response type than the one asked for. Match it with [errors.As].
type ResponseTypeError struct {
	Request    reflect.Type // the request type dispatched
	Registered reflect.Type // the response type of the registered use case
	Requested  reflect.Type // the response type Dispatch was asked for
}

// Error describes the mismatch.
func (e *ResponseTypeError) Error() string {
	return fmt.Sprintf("interactor: use case for %v returns %v, not %v", e.Request, e.Registered, e.Requested)
}
