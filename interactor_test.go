package interactor_test

import (
	"context"
	"errors"
	"testing"

	"github.com/screwyprof/interactor"
	. "github.com/screwyprof/interactor/testdsl"
)

func TestInteractorAcceptance(t *testing.T) {
	t.Parallel()

	t.Run("ensure use case returns an error", func(t *testing.T) {
		t.Parallel()

		want := errors.New("some error")
		useCaseRunner := &ConcreteUseCase{err: want}
		adaptedUseCaseRunner := interactor.MustAdapt(useCaseRunner)

		Test(t)(
			Given(adaptedUseCaseRunner.Run),
			When(context.Background(), TestRequest{}, &TestResponse{}),
			ThenFailWith(want),
		)
	})

	t.Run("ensure use case returns valid result", func(t *testing.T) {
		t.Parallel()

		want := &TestResponse{result: 123}
		useCaseRunner := &ConcreteUseCase{}
		adaptedUseCaseRunner := interactor.Must(interactor.Func(useCaseRunner.Run))

		Test(t)(
			Given(adaptedUseCaseRunner.Run),
			When(context.Background(), TestRequest{123}, &TestResponse{}),
			Then(want),
		)
	})
}
