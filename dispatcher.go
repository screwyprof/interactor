package interactor

import (
	"context"
	"errors"
	"fmt"
	"reflect"
)

// ErrUseCaseRunnerNotFound is returned when the use case runner is not registered for the given Request type.
var ErrUseCaseRunnerNotFound = errors.New("use case runner not registered for the given request type")

// Dispatcher manages registered UseCaseRunners and dispatches requests to the appropriate UseCaseRunner.
type Dispatcher struct {
	useCaseRunners map[reflect.Type]UseCaseRunnerFn
}

// NewDispatcher creates a new Dispatcher instance.
func NewDispatcher() *Dispatcher {
	return &Dispatcher{
		useCaseRunners: make(map[reflect.Type]UseCaseRunnerFn),
	}
}

// Register registers the given UseCaseRunner for the provided request type.
func (d *Dispatcher) Register(request Request, runner UseCaseRunnerFn) {
	requestType := reflect.TypeOf(request)
	d.useCaseRunners[requestType] = runner
}

// Run runs a use case with the given Request and writes the result to the provided Response.
//
// It returns nil if the use case was executed successfully.
// It returns ErrUseCaseRunnerNotFound  if the use case runner is not registered for the Request type.
func (d *Dispatcher) Run(ctx context.Context, req Request, resp Response) error {
	reqType := reflect.TypeOf(req)

	runner, ok := d.useCaseRunners[reqType]
	if !ok {
		return fmt.Errorf("%w: %s", ErrUseCaseRunnerNotFound, reqType)
	}

	return runner(ctx, req, resp)
}
