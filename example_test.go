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
