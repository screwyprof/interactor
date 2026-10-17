package interactor_test

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/screwyprof/interactor/v3"
	"github.com/screwyprof/interactor/v3/testdsl"
)

func TestUseCaseFunc(t *testing.T) {
	t.Parallel()

	testdsl.Test(t).
		Given(interactor.UseCaseFunc[Question, Answer](AnswerQuestion{}.Run)).
		When(t.Context(), Question{ID: 21}).
		Then(Answer{Value: 42})
}

func TestDispatchRunsTheUseCaseUnderTheCallersContext(t *testing.T) {
	t.Parallel()

	ctx := context.WithValue(t.Context(), ctxKey{}, "marker")

	readCtx := func(ctx context.Context, _ Question) (Answer, error) {
		v, _ := ctx.Value(ctxKey{}).(string)
		if v != "marker" {
			return Answer{}, errBoom
		}

		return Answer{Value: 1}, nil
	}

	var d interactor.Dispatcher
	d.MustRegister(interactor.UseCaseFunc[Question, Answer](readCtx))

	_, err := d.Dispatch[Answer](ctx, Question{})
	require.NoError(t, err)
}

type ctxKey struct{}
