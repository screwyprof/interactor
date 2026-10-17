// Package interactor runs use cases and dispatches requests to them by request type.
//
// A use case is any type with a Run method taking a request and returning a response:
//
//	type Greet struct{}
//
//	func (Greet) Run(ctx context.Context, req Hello) (Greeting, error) { ... }
//
// Register it on a [Dispatcher], then dispatch requests of that type:
//
//	var d interactor.Dispatcher
//	if err := d.Register(Greet{}); err != nil { ... }
//
//	greeting, err := d.Dispatch[Greeting](ctx, Hello{Name: "Gopher"})
//
// Types are checked by the compiler; only the wiring (a request type nobody registered, or a
// response type other than the registered one) is checked when dispatching.
package interactor

import "context"

// UseCase handles one request type and returns its response.
type UseCase[Req, Res any] interface {
	Run(ctx context.Context, req Req) (Res, error)
}

// UseCaseFunc lets a plain function be a [UseCase], like [net/http.HandlerFunc].
type UseCaseFunc[Req, Res any] func(ctx context.Context, req Req) (Res, error)

// Run calls f(ctx, req).
func (f UseCaseFunc[Req, Res]) Run(ctx context.Context, req Req) (Res, error) {
	return f(ctx, req)
}
