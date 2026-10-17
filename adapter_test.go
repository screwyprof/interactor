package interactor_test

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/screwyprof/interactor/v2"
)

func TestFunc(t *testing.T) {
	t.Parallel()

	testCases := []testCase{
		{
			name:    "a use case runner must be a function",
			runner:  struct{}{},
			wantErr: interactor.ErrUseCaseRunnerIsNotAFunction,
		},
		{
			name:    "a use case runner must not be nil",
			runner:  nil,
			wantErr: interactor.ErrUseCaseRunnerIsNotAFunction,
		},
		{
			name:    "a use case runner must not be a nil function",
			runner:  (func(context.Context, TestRequest, *TestResponse) error)(nil),
			wantErr: interactor.ErrUseCaseRunnerIsNotAFunction,
		},
		{
			name: "first input param must be context.Context, not a type implementing it",
			runner: func(_ *customContext, _ TestRequest, _ *TestResponse) error {
				return nil
			},
			wantErr: interactor.ErrFirstArgHasInvalidType,
		},
		{
			name: "first input param must be context.Context, not a wider interface",
			runner: func(_ extendedContext, _ TestRequest, _ *TestResponse) error {
				return nil
			},
			wantErr: interactor.ErrFirstArgHasInvalidType,
		},
		{
			name:    "a use case runner must return an error",
			runner:  func(_ context.Context, _ TestRequest, _ *TestResponse) {},
			wantErr: interactor.ErrInvalidReturnType,
		},
		{
			name:    "a use case runner must not return a non-error value",
			runner:  func(_ context.Context, _ TestRequest, _ *TestResponse) int { return 0 },
			wantErr: interactor.ErrInvalidReturnType,
		},
		{
			name: "a use case runner must return only an error",
			runner: func(_ context.Context, _ TestRequest, _ *TestResponse) (int, error) {
				return 0, nil
			},
			wantErr: interactor.ErrInvalidReturnType,
		},
		{
			name:          "request must have the runner's request type",
			runner:        ConcreteUseCase{}.Run,
			request:       "oops",
			response:      &TestResponse{},
			wantRunnerErr: interactor.ErrRequestTypeMismatch,
		},
		{
			name:          "request must not be a pointer when the runner takes a value",
			runner:        ConcreteUseCase{}.Run,
			request:       &TestRequest{},
			response:      &TestResponse{},
			wantRunnerErr: interactor.ErrRequestTypeMismatch,
		},
		{
			name:          "request must not be nil",
			runner:        ConcreteUseCase{}.Run,
			request:       nil,
			response:      &TestResponse{},
			wantRunnerErr: interactor.ErrRequestTypeMismatch,
		},
		{
			name: "a use case runner must have 3 arguments",
			runner: func(_ context.Context, _ TestRequest) error {
				return nil
			},
			wantErr: interactor.ErrInvalidUseCaseRunnerSignature,
		},
		{
			name: "first input param must be context.Context",
			runner: func(_ struct{}, _ TestRequest, _ *TestResponse) error {
				return nil
			},
			wantErr: interactor.ErrFirstArgHasInvalidType,
		},
		{
			name: "second input param must be a struct implementing Request",
			runner: func(_ context.Context, _ int, _ *TestResponse) error {
				return nil
			},
			wantErr: interactor.ErrSecondArgHasInvalidType,
		},
		{
			name: "third input param must be a pointer type implementing Response",
			runner: func(_ context.Context, _ TestRequest, _ struct{}) error {
				return nil
			},
			wantErr: interactor.ErrThirdArgHasInvalidType,
		},
		{
			name: "provided response type must match expected response type",
			runner: func(_ context.Context, _ TestRequest, _ *TestResponse) error {
				return nil
			},
			request: TestRequest{},
			response: &struct {
				interactor.Response
			}{},
			wantRunnerErr: interactor.ErrResultTypeMismatch,
		},
		{
			name: "a use case returns an error",
			runner: func(_ context.Context, _ TestRequest, _ *TestResponse) error {
				return errSomeErr
			},
			request:       TestRequest{id: 123},
			response:      &TestResponse{},
			wantRunnerErr: errSomeErr,
		},
		{
			name: "second input param maybe be a pointer to a struct implementing Request",
			runner: func(_ context.Context, _ *TestRequest, _ *TestResponse) error {
				return nil
			},
			request:    &TestRequest{},
			response:   &TestResponse{},
			wantResult: &TestResponse{},
		},
		{
			name: "provided use case successfully adapted to comply with UseCaseRunner interface",
			runner: func(_ context.Context, req TestRequest, resp *TestResponse) error {
				resp.result = req.id

				return nil
			},
			request:    TestRequest{id: 123},
			response:   &TestResponse{},
			wantResult: &TestResponse{result: 123},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			runFuncTest(t, tc)
		})
	}
}

func TestAdapt(t *testing.T) {
	t.Parallel()

	testCases := []testCase{
		{
			name:    "a use case must have a Run method",
			runner:  struct{}{},
			wantErr: interactor.ErrUseCaseRunnerHasNoRunMethod,
		},
		{
			name:    "a use case must not be nil",
			runner:  nil,
			wantErr: interactor.ErrUseCaseRunnerHasNoRunMethod,
		},
		{
			name:    "a use case must not be a nil pointer",
			runner:  (*ConcreteUseCase)(nil),
			wantErr: interactor.ErrUseCaseRunnerHasNoRunMethod,
		},
		{
			name:    "method must have 3 arguments",
			runner:  InvalidUseCaseWrongSignature{},
			wantErr: interactor.ErrInvalidUseCaseRunnerSignature,
		},
		{
			name:    "first input param must be context.Context",
			runner:  InvalidUseCaseWrongContext{},
			wantErr: interactor.ErrFirstArgHasInvalidType,
		},
		{
			name:    "second input param must be a struct implementing Request",
			runner:  InvalidUseCaseWrongRequest{},
			wantErr: interactor.ErrSecondArgHasInvalidType,
		},
		{
			name:    "third input param must be a pointer type implementing Response",
			runner:  InvalidUseCaseWrongResponse{},
			wantErr: interactor.ErrThirdArgHasInvalidType,
		},
		{
			name:          "response types must match",
			runner:        ConcreteUseCase{},
			request:       TestRequest{},
			response:      &AnotherResponse{},
			wantRunnerErr: interactor.ErrResultTypeMismatch,
		},
		{
			name:          "a use case returns an error",
			runner:        ConcreteUseCase{err: errSomeErr},
			request:       TestRequest{id: 123},
			response:      &TestResponse{},
			wantRunnerErr: errSomeErr,
		},
		{
			name:       "second input param maybe be a pointer type implementing Request",
			runner:     ValidUseCasePointerRequest{},
			request:    &TestRequest{},
			response:   &TestResponse{},
			wantResult: &TestResponse{},
		},
		{
			name:       "provided use case successfully adapted to comply with UseCaseRunner interface",
			runner:     ConcreteUseCase{},
			request:    TestRequest{id: 123},
			response:   &TestResponse{},
			wantResult: &TestResponse{result: 123},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			runAdaptTest(t, tc)
		})
	}
}

func TestFuncRejectsNilContext(t *testing.T) {
	t.Parallel()

	runner := interactor.Must(interactor.Func(ConcreteUseCase{}.Run))

	err := runner(nil, TestRequest{id: 7}, &TestResponse{})

	require.ErrorIs(t, err, interactor.ErrNilContext)
}

func TestMust(t *testing.T) {
	t.Parallel()

	t.Run("it panics if it cannot adapt a use case runner", func(t *testing.T) {
		t.Parallel()

		assert.Panics(t, func() {
			interactor.Must(interactor.Func(struct{}{}))
		})
	})

	t.Run("it adapts a use case runner function", func(t *testing.T) {
		t.Parallel()

		// act
		runner := interactor.Must(interactor.Func(ConcreteUseCase{}.Run))

		// assert
		assert.NotNil(t, runner)
	})
}

type testCase struct {
	name          string
	runner        any
	request       interactor.Request
	response      interactor.Response
	wantErr       error
	wantRunnerErr error
	wantResult    interactor.Response
}

func runFuncTest(t *testing.T, tc testCase) {
	t.Helper()

	runner, err := interactor.Func(tc.runner)
	if tc.wantErr != nil {
		require.ErrorIs(t, err, tc.wantErr)

		return
	}

	require.NoError(t, err)

	err = runner(context.Background(), tc.request, tc.response)
	if tc.wantRunnerErr != nil {
		require.ErrorIs(t, err, tc.wantRunnerErr)

		return
	}

	require.NoError(t, err)
	assert.Equal(t, tc.wantResult, tc.response)
}

func runAdaptTest(t *testing.T, tc testCase) {
	t.Helper()

	runner, err := interactor.Adapt(tc.runner)
	if tc.wantErr != nil {
		require.ErrorIs(t, err, tc.wantErr)

		return
	}

	require.NoError(t, err)

	err = runner(context.Background(), tc.request, tc.response)
	if tc.wantRunnerErr != nil {
		require.ErrorIs(t, err, tc.wantRunnerErr)

		return
	}

	require.NoError(t, err)
	assert.Equal(t, tc.wantResult, tc.response)
}
