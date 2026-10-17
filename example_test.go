package interactor_test

import (
	"context"
	"fmt"
	"log"

	"github.com/screwyprof/interactor/v2"
)

func ExampleDispatcher() {
	// arrange
	useCaseRunner := &ConcreteUseCase{}

	dispatcher := interactor.NewDispatcher()
	dispatcher.Register(TestRequest{}, interactor.MustAdapt(useCaseRunner))

	// act
	var res TestResponse
	if err := dispatcher.Run(context.Background(), TestRequest{id: 42}, &res); err != nil {
		log.Fatal(err)
	}

	fmt.Printf("The answer to life the universe and everything: %d\n", res.result)

	// Output:
	// The answer to life the universe and everything: 42
}

// The README usage example, compiled and run here. Types are prefixed to avoid clashing with
// the test helpers and the parameter names are blanked for the linter; keep the rest in step
// with the README.

// ReadmeRequest represents a test request.
type ReadmeRequest struct{}

// ReadmeResponse represents a test response.
type ReadmeResponse struct {
	Result int
}

// ReadmeUseCase is an example use case implementation.
type ReadmeUseCase struct {
	Res int
}

// Run runs the use case with the given request.
func (uc ReadmeUseCase) Run(_ context.Context, _ ReadmeRequest, res *ReadmeResponse) error {
	res.Result = uc.Res

	return nil
}

func ExampleMustAdapt() {
	// Create a use case.
	useCaseRunner := &ReadmeUseCase{Res: 42}

	// Create a new dispatcher and register the use case runner.
	dispatcher := interactor.NewDispatcher()
	dispatcher.Register(ReadmeRequest{}, interactor.MustAdapt(useCaseRunner))

	// Run the use case
	var res ReadmeResponse
	if err := dispatcher.Run(context.Background(), ReadmeRequest{}, &res); err != nil {
		log.Fatal(err)
	}

	fmt.Printf("The answer to life, the universe, and everything: %d\n", res.Result)

	// Output:
	// The answer to life, the universe, and everything: 42
}
