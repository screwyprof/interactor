package testdsl

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/screwyprof/interactor"
)

// GivenFn is a test init function.
type GivenFn func() interactor.UseCaseRunner

// WhenFn is a use case runner function.
type WhenFn func(useCaseRunner interactor.UseCaseRunner) (interactor.Response, error)

// ThenFn prepares the Checker.
type ThenFn func(t testing.TB) Checker

// Checker asserts the given results.
type Checker func(resp interactor.Response, err error)

// InteractorTester defines a Test Suite.
type InteractorTester func(given GivenFn, when WhenFn, then ThenFn)

// Test runs the test.
//
// Example:
//
//	want := &TestResponse{result: 123}
//	useCaseRunner := &ConcreteUseCase{res: want.result}
//	adaptedUseCaseRunner := interactor.MustAdapt(useCaseRunner.Run)
//
//	Test(t)(
//		Given(adaptedUseCaseRunner.Run),
//		When(context.Background(), TestRequest{123}, &TestResponse{}),
//		Then(want))
func Test(tb testing.TB) InteractorTester {
	tb.Helper()

	return func(given GivenFn, when WhenFn, then ThenFn) {
		tb.Helper()

		then(tb)(when(given()))
	}
}

// Given prepares the given use case runner for testing.
func Given(useCaseRunner interactor.UseCaseRunnerFn) GivenFn {
	return func() interactor.UseCaseRunner {
		return useCaseRunner
	}
}

// When runs the use case.
func When(ctx context.Context, req interactor.Request, res interactor.Response) WhenFn {
	return func(useCaseRunner interactor.UseCaseRunner) (interactor.Response, error) {
		return res, useCaseRunner.Run(ctx, req, res)
	}
}

// Then asserts that the expected response is returned.
func Then(want interactor.Response) ThenFn {
	return func(tb testing.TB) Checker {
		tb.Helper()

		return func(got interactor.Response, err error) {
			tb.Helper()

			assert.NoError(tb, err)
			assert.Equal(tb, want, got)
		}
	}
}

// ThenFailWith asserts that the expected error occurred.
func ThenFailWith(want error) ThenFn {
	return func(tb testing.TB) Checker {
		tb.Helper()

		return func(res interactor.Response, err error) {
			tb.Helper()

			assert.ErrorIs(tb, err, want)
		}
	}
}
