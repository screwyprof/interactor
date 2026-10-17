package interactor_test

import (
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

func TestUseCaseFuncWithCustomCheck(t *testing.T) {
	t.Parallel()

	testdsl.Test(t).
		Given(interactor.UseCaseFunc[Question, Answer](AnswerQuestion{}.Run)).
		When(Question{ID: 5}).
		ThenCheck(func(tb testing.TB, got Answer, err error) {
			tb.Helper()

			require.NoError(tb, err)
			assert.Positive(tb, got.Value)
		})
}
