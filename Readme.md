# Interactor

`Interactor` is a simple and efficient Go package for managing and dispatching requests to the appropriate use cases according to their request types. It is inspired by Uncle Bob's [Clean Architecture](https://blog.cleancoder.com/uncle-bob/2012/08/13/the-clean-architecture.html) and the interactor concept described therein.

The library is designed to be well-documented, thoroughly tested, and easy to use, ensuring maintainable and reliable code for your projects.

## Features

- Dispatcher for managing different use cases.
- Flexible use cases as either pure functions or structures.
- Well-documented and tested code.

## Installation
To install the Interactor Library, run the following command:

```bash
go get -u github.com/screwyprof/interactor
```

## Usage

```go
package main

import (
	"context"
	"fmt"
	"log"

	"github.com/screwyprof/interactor"
)

// TestRequest represents a test request.
type TestRequest struct{}

// TestResponse represents a test response.
type TestResponse struct {
	Result int
}

// UseCase is an example use case implementation.
type UseCase struct {
	Res int
}

// Run runs the use case with the given request.
//
// The response is returned by reference to avoid extra allocations.
func (uc UseCase) Run(ctx context.Context, req TestRequest, res *TestResponse) error {
	res.Result = uc.Res
	return nil
}

func main() {
	// Create a use case.
	useCaseRunner := &UseCase{Res: 42}

	// Create a new dispatcher and register the use case runner.
	dispatcher := interactor.NewDispatcher()
	dispatcher.Register(TestRequest{}, interactor.MustAdapt(useCaseRunner.Run))

	// Run the use case
	var res TestResponse
	if err := dispatcher.Run(context.Background(), TestRequest{}, &res); err != nil {
		log.Fatal(err)
	}

	fmt.Printf("The answer to life, the universe, and everything: %d\n", res.Result)
}
```

## Documentation

For more information and detailed documentation, please refer to the GoDoc documentation:

[Interactor Library GoDoc](https://pkg.go.dev/github.com/screwyprof/interactor)

## Contributing

Contributions are welcome! If you have any suggestions, bug reports, or feature requests, please open a new issue or submit a pull request.

## License

The Interactor Library is released under the [MIT License](https://opensource.org/licenses/MIT).