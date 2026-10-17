package interactor

import "context"

// FromOutParam adapts a v2-style use case, which writes its response through a pointer, to a
// [UseCaseFunc]. It returns nil for a nil run, which [Dispatcher.Register] rejects; calling Run on
// that nil directly panics. A method value bound to a nil pointer, such as uc.Run on a nil *T, is
// not nil and is not detected: check the receiver before wrapping it.
//
// On error the response is the zero value, even if run wrote to it before failing: a v2 use case
// that reports details alongside its error, such as validation problems, must return them in the
// error instead.
//
//	d.MustRegister(interactor.FromOutParam(legacy.Run))
func FromOutParam[Req, Res any](run func(ctx context.Context, req Req, res *Res) error) UseCaseFunc[Req, Res] {
	if run == nil {
		return nil
	}

	return UseCaseFunc[Req, Res](func(ctx context.Context, req Req) (Res, error) {
		var res Res

		if err := run(ctx, req, &res); err != nil {
			var zero Res

			return zero, err
		}

		return res, nil
	})
}
