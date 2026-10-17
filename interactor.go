// Package interactor provides a flexible library for running use cases.
//
// This package allows registering use cases for different request types,
// efficiently dispatching requests to the appropriate handlers.
//
// Register a use case for its request type, then run requests through the dispatcher:
//
//	dispatcher := interactor.NewDispatcher()
//	dispatcher.Register(MyRequest{}, interactor.MustAdapt(&MyUseCase{}))
//
//	var res MyResponse
//	err := dispatcher.Run(ctx, MyRequest{}, &res)
//
// See the examples for complete, runnable code.
package interactor

import "context"

// Request is an interface representing the input for the use case.
type Request any

// Response is an interface representing the output from the use case.
type Response any

// UseCaseRunner is an interface that defines a contract for running a use case.
type UseCaseRunner interface {
	// Run executes the given request and writes the result to the provided response.
	Run(ctx context.Context, req Request, resp Response) error
}

// UseCaseRunnerFn allows using pure functions as a UseCaseRunner.
type UseCaseRunnerFn func(ctx context.Context, req Request, resp Response) error

// Run executes the given request and writes the result to the provided response.
func (fn UseCaseRunnerFn) Run(ctx context.Context, req Request, resp Response) error {
	return fn(ctx, req, resp)
}
