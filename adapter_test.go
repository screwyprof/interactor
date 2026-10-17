package interactor_test

import (
	"testing"

	"github.com/screwyprof/interactor/v3"
	"github.com/screwyprof/interactor/v3/testdsl"
)

func TestUseCaseFunc(t *testing.T) {
	t.Parallel()

	uc := interactor.UseCaseFunc[Question, Answer](AnswerQuestion{}.Run)

	testdsl.Test(t, uc, Question{ID: 21}, testdsl.ThenReturns(Answer{Value: 42}))
}

func TestFromOutParam(t *testing.T) {
	t.Parallel()

	t.Run("returns what the use case wrote", func(t *testing.T) {
		t.Parallel()

		uc := interactor.FromOutParam(legacyAnswerQuestion)

		testdsl.Test(t, uc, Question{ID: 21}, testdsl.ThenReturns(Answer{Value: 42}))
	})

	t.Run("returns the use case's error", func(t *testing.T) {
		t.Parallel()

		uc := interactor.FromOutParam(legacyAnswerQuestion)

		testdsl.Test(t, uc, Question{ID: -1}, testdsl.ThenFails[Answer](errBoom))
	})
}
