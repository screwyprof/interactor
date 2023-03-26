// Package interactor provides a flexible library for running use cases.
//
// This package allows registering use cases for different request types,
// efficiently dispatching requests to the appropriate handlers.
//
// Example usage:
//
//	func ExampleDispatcher() {
//		// arrange
//		useCaseRunner := &ConcreteUseCase{res: 42}
//
//		dispatcher := interactor.NewDispatcher()
//		dispatcher.Register(TestRequest{}, interactor.MustAdapt(useCaseRunner.RunUseCase))
//
//		// act
//		var res TestResponse
//		if err := dispatcher.Run(context.Background(), TestRequest{}, &res); err != nil {
//			log.Fatal(err)
//		}
//
//		fmt.Printf("The answer to life the universe and everything: %d\n", res.result)
//
//		// Output:
//		// The answer to life the universe and everything: 42
//	}
package interactor

import "context"

// Request is an interface representing the input for the use case.
type Request interface{}

// Response is an interface representing the output from the use case.
type Response interface{}

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
