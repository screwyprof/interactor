package interactor_test

import (
	"testing"

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
