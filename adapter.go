package interactor

import "context"

// FromOutParam adapts a v2-style use case, which writes its response through a pointer, to a
// [UseCaseFunc]. It returns nil for a nil run, which [Dispatcher.Register] then rejects.
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
