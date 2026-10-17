package interactor

import "context"

// FromOutParam adapts a v2-style use case, which writes its response through a pointer, to a
// [UseCase]. It returns nil for a nil run, which [Dispatcher.Register] then rejects.
//
//	err := d.Register(interactor.FromOutParam(legacy.Run))
func FromOutParam[Req, Res any](run func(ctx context.Context, req Req, res *Res) error) UseCase[Req, Res] {
	if run == nil {
		return nil
	}

	return UseCaseFunc[Req, Res](func(ctx context.Context, req Req) (Res, error) {
		var res Res

		err := run(ctx, req, &res)

		return res, err
	})
}
