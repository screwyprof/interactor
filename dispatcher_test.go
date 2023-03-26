package interactor_test

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/screwyprof/interactor"
)

func TestDispatcher(t *testing.T) {
	t.Parallel()

	t.Run("dispatcher implements UseCaseRunner interface", func(t *testing.T) {
		t.Parallel()

		dispatcher := interactor.NewDispatcher()
		var _ interactor.UseCaseRunner = dispatcher
	})

	t.Run("when use case not found, an error returned", func(t *testing.T) {
		t.Parallel()

		// arrange
		dispatcher := interactor.NewDispatcher()

		// act
		var res TestResponse
		err := dispatcher.Run(context.Background(), TestRequest{}, &res)

		// assert
		assertUseCaseRunnerNotFound(t, err)
	})

	t.Run("when use case registered, it is being run", func(t *testing.T) {
		t.Parallel()

		// arrange
		useCaseRunner := &GeneralUseCaseSpy{}

		dispatcher := interactor.NewDispatcher()
		dispatcher.Register(TestRequest{}, useCaseRunner.Run)

		// act
		var res TestResponse
		err := dispatcher.Run(context.Background(), TestRequest{}, &res)

		// assert
		assertUseCaseWasRunSuccessfully(t, err, useCaseRunner)
	})
}

func BenchmarkDispatcher(b *testing.B) {
	dispatcher := interactor.NewDispatcher()
	useCaseRunner := &GeneralUseCaseSpy{}
	dispatcher.Register(TestRequest{}, useCaseRunner.Run)

	var res TestResponse

	req := TestRequest{id: 123}
	ctx := context.Background()

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		_ = dispatcher.Run(ctx, req, &res)
	}
}

func assertUseCaseWasRunSuccessfully(t *testing.T, err error, useCaseRunner *GeneralUseCaseSpy) {
	t.Helper()

	assert.NoError(t, err)
	assert.True(t, useCaseRunner.wasCalled)
}

func assertUseCaseRunnerNotFound(t *testing.T, err error) {
	t.Helper()

	assert.True(t, errors.Is(err, interactor.ErrUseCaseRunnerNotFound))
}
