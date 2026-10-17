package interactor_test

import (
	"context"
	"errors"
	"fmt"

	"github.com/screwyprof/interactor/v3"
)

type Hello struct{ Name string }

type Greeting struct{ Text string }

// Greet is a use case: a type with a Run method from a request to a response.
type Greet struct{}

func (Greet) Run(_ context.Context, req Hello) (Greeting, error) {
	return Greeting{Text: "Hello, " + req.Name + "!"}, nil
}

func Example() {
	var d interactor.Dispatcher

	d.MustRegister(Greet{})

	greeting, err := d.Dispatch[Greeting](context.Background(), Hello{Name: "Gopher"})
	if err != nil {
		fmt.Println(err)

		return
	}

	fmt.Println(greeting.Text)

	// Output:
	// Hello, Gopher!
}

func ExampleDispatcher_Dispatch_notFound() {
	var d interactor.Dispatcher

	_, err := d.Dispatch[Greeting](context.Background(), Hello{Name: "Gopher"})

	fmt.Println(errors.Is(err, interactor.ErrUseCaseNotFound))

	// Output:
	// true
}

// legacyGreet is a v2-style use case, writing its response through a pointer.
func legacyGreet(_ context.Context, req Hello, res *Greeting) error {
	res.Text = "Hello again, " + req.Name + "!"

	return nil
}

func ExampleFromOutParam() {
	var d interactor.Dispatcher

	d.MustRegister(interactor.FromOutParam(legacyGreet))

	greeting, _ := d.Dispatch[Greeting](context.Background(), Hello{Name: "Gopher"})

	fmt.Println(greeting.Text)

	// Output:
	// Hello again, Gopher!
}

func ExampleUseCaseFunc() {
	shout := interactor.UseCaseFunc[Hello, Greeting](func(_ context.Context, req Hello) (Greeting, error) {
		return Greeting{Text: "HELLO, " + req.Name + "!"}, nil
	})

	var d interactor.Dispatcher

	d.MustRegister(shout)

	greeting, _ := d.Dispatch[Greeting](context.Background(), Hello{Name: "Gopher"})

	fmt.Println(greeting.Text)

	// Output:
	// HELLO, Gopher!
}

func ExampleDispatcher_Register() {
	var d interactor.Dispatcher

	d.MustRegister(Greet{})

	// Register returns the error instead, for wiring that is decided at runtime.
	err := d.Register(Greet{})

	fmt.Println(errors.Is(err, interactor.ErrAlreadyRegistered))

	// Output:
	// true
}
