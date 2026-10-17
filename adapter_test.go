package interactor_test

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/screwyprof/interactor/v3"
	"github.com/screwyprof/interactor/v3/testdsl"
)

func TestUseCaseFunc(t *testing.T) {
	t.Parallel()

	testdsl.Test(t).
		Given(interactor.UseCaseFunc[Question, Answer](AnswerQuestion{}.Run)).
		When(Question{ID: 21}).
		Then(Answer{Value: 42})
}

func TestFromOutParam(t *testing.T) {
	t.Parallel()

	t.Run("returns what the use case wrote", func(t *testing.T) {
		t.Parallel()

		testdsl.Test(t).
			Given(interactor.FromOutParam(legacyAnswerQuestion)).
			When(Question{ID: 21}).
			Then(Answer{Value: 42})
	})

	t.Run("returns the use case's error", func(t *testing.T) {
		t.Parallel()

		testdsl.Test(t).
			Given(interactor.FromOutParam(legacyAnswerQuestion)).
			When(Question{ID: -1}).
			ThenFailWith(errBoom)
	})
}

func TestFromOutParamReturnsNoPartialResponseOnError(t *testing.T) {
	t.Parallel()

	halfDone := func(_ context.Context, _ Question, res *Answer) error {
		res.Value = 99 // written before failing

		return errBoom
	}

	got, err := interactor.FromOutParam(halfDone).Run(t.Context(), Question{})

	require.ErrorIs(t, err, errBoom)
	assert.Zero(t, got, "a failed use case's response must not leak through")
}

func TestUseCasesRunUnderTheCallersContext(t *testing.T) {
	t.Parallel()

	ctx := context.WithValue(t.Context(), ctxKey{}, "marker")

	readCtx := func(ctx context.Context, _ Question) (Answer, error) {
		v, _ := ctx.Value(ctxKey{}).(string)
		if v != "marker" {
			return Answer{}, errBoom
		}

		return Answer{Value: 1}, nil
	}
	readCtxOutParam := func(ctx context.Context, q Question, res *Answer) error {
		got, err := readCtx(ctx, q)
		*res = got

		return err
	}

	t.Run("through Dispatch", func(t *testing.T) {
		t.Parallel()

		var d interactor.Dispatcher
		d.MustRegister(interactor.UseCaseFunc[Question, Answer](readCtx))

		_, err := d.Dispatch[Answer](ctx, Question{})
		require.NoError(t, err)
	})

	t.Run("through FromOutParam", func(t *testing.T) {
		t.Parallel()

		_, err := interactor.FromOutParam(readCtxOutParam).Run(ctx, Question{})
		require.NoError(t, err)
	})
}

type ctxKey struct{}

func TestFromOutParamStartsEachCallFromAZeroResponse(t *testing.T) {
	t.Parallel()

	// Writes only when asked to: a second call must not see what the first one wrote.
	writeOnce := func(_ context.Context, q Question, res *Answer) error {
		if q.ID > 0 {
			res.Value = q.ID
		}

		return nil
	}

	uc := interactor.FromOutParam(writeOnce)

	first, err := uc.Run(t.Context(), Question{ID: 7})
	require.NoError(t, err)
	require.Equal(t, Answer{Value: 7}, first)

	second, err := uc.Run(t.Context(), Question{})
	require.NoError(t, err)
	assert.Zero(t, second, "each call must start from a zero response")
}
