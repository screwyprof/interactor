package interactor_test

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/screwyprof/interactor"
)

func TestFunc(t *testing.T) {
	t.Parallel()

	t.Run("a use case runner must be a function", func(t *testing.T) {
		t.Parallel()

		// act
		_, err := interactor.Func(struct{}{})

		// assert
		assertUseCaseRunnerIsAFunction(t, err)
	})

	t.Run("a use case runner must have 3 arguments", func(t *testing.T) {
		t.Parallel()

		// arrange
		invalidRunner := func(ctx context.Context, req TestRequest) error {
			return nil
		}

		// act
		_, err := interactor.Func(invalidRunner)

		// assert
		assertUseCaseRunnerHasInvalidSignature(t, err)
	})

	t.Run("first input param must be context.Context", func(t *testing.T) {
		t.Parallel()

		// arrange
		invalidRunner := func(ctx struct{}, req TestRequest, resp *TestResponse) error {
			return nil
		}

		// act
		_, err := interactor.Func(invalidRunner)

		// assert
		assertFirstArgHasContextType(t, err)
	})

	t.Run("second input param must be a struct implementing Request", func(t *testing.T) {
		t.Parallel()

		// arrange
		invalidRunner := func(ctx context.Context, req int, resp *TestResponse) error {
			return nil
		}

		// act
		_, err := interactor.Func(invalidRunner)

		// assert
		assertSecondArgIsAreRequestType(t, err)
	})
	t.Run("second input param maybe be a pointer to a struct implementing Request", func(t *testing.T) {
		t.Parallel()

		// arrange
		runner := func(ctx context.Context, req *TestRequest, resp *TestResponse) error {
			return nil
		}

		// act
		_, err := interactor.Func(runner)

		// assert
		require.NoError(t, err)
	})

	t.Run("third input param must be a pointer type implementing Response", func(t *testing.T) {
		t.Parallel()

		// arrange
		invalidRunner := func(ctx context.Context, req TestRequest, resp struct{}) error {
			return nil
		}

		// act
		_, err := interactor.Func(invalidRunner)

		// assert
		assertThirdArgIsAResponseType(t, err)
	})

	t.Run("provided response type must match expected response type", func(t *testing.T) {
		t.Parallel()

		// arrange
		type AnotherResponse struct{}

		runner, err := interactor.Func(ConcreteUseCase{}.Run)
		assert.NoError(t, err)

		// act
		err = runner(context.Background(), TestRequest{}, &AnotherResponse{})

		// assert
		assert.ErrorIs(t, err, interactor.ErrResultTypeMismatch)
	})

	t.Run("provided uses cases successfully adapted to comply with UseCaseRunner interface", func(t *testing.T) {
		t.Parallel()

		// act
		_, err := interactor.Func(ConcreteUseCase{}.Run)

		// assert
		assert.NoError(t, err)
	})

	t.Run("ensure that the given valid concrete use case runner can return valid result", func(t *testing.T) {
		t.Parallel()

		// arrange
		want := TestResponse{result: 123}

		// act
		runner, err := interactor.Func(ConcreteUseCase{}.Run)
		assert.NoError(t, err)

		var res TestResponse
		err = runner(context.Background(), TestRequest{id: 123}, &res)

		// assert
		assert.NoError(t, err)
		assert.Equal(t, want, res)
	})

	t.Run("ensure that the given valid concrete use case runner can return error result", func(t *testing.T) {
		t.Parallel()

		// arrange
		want := errSomeErr

		// act
		runner, err := interactor.Func(ConcreteUseCase{err: want}.Run)
		assert.NoError(t, err)

		err = runner(context.Background(), TestRequest{}, &TestResponse{})

		// assert
		assert.ErrorIs(t, err, want)
	})
}

func TestMust(t *testing.T) {
	t.Parallel()

	t.Run("it panics if it cannot adapt a use case runner", func(t *testing.T) {
		t.Parallel()

		assert.Panics(t, func() {
			interactor.Must(interactor.Func(struct{}{}))
		})
	})

	t.Run("it adapts a use case runner", func(t *testing.T) {
		t.Parallel()

		// act
		runner := interactor.Must(interactor.Func(ConcreteUseCase{}.Run))

		// assert
		assert.NotNil(t, runner)
	})
}

func assertUseCaseRunnerHasInvalidSignature(t *testing.T, err error) {
	t.Helper()
	assert.ErrorIs(t, err, interactor.ErrInvalidUseCaseRunnerSignature)
}

func assertUseCaseRunnerIsAFunction(t *testing.T, err error) {
	t.Helper()
	assert.ErrorIs(t, err, interactor.ErrUseCaseRunnerIsNotAFunction)
}

func assertFirstArgHasContextType(t *testing.T, err error) {
	t.Helper()
	assert.ErrorIs(t, err, interactor.ErrFirstArgHasInvalidType)
}

func assertSecondArgIsAreRequestType(t *testing.T, err error) {
	t.Helper()
	assert.ErrorIs(t, err, interactor.ErrSecondArgHasInvalidType)
}

func assertThirdArgIsAResponseType(t *testing.T, err error) {
	t.Helper()
	assert.ErrorIs(t, err, interactor.ErrThirdArgHasInvalidType)
}
