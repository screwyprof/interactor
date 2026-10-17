// Package testdsl is a small Given/When/Then DSL for testing your own use cases.
//
// Import it from your tests when you want each test to read as a scenario:
//
//	func TestGreet(t *testing.T) {
//		testdsl.Test(t,
//			Greet{},                      // given this use case
//			Hello{Name: "Gopher"},        // when it runs with this request
//			testdsl.ThenReturns(Greeting{Text: "Hello, Gopher!"}), // then
//		)
//	}
//
// [ThenFails] checks the error instead. Write your own [Then] for anything else; it is a plain
// func(tb, got, err).
//
// It uses testify, which your tests then depend on too.
package testdsl

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/screwyprof/interactor/v3"
)

// Then checks what a use case returned.
type Then[Res any] func(tb testing.TB, got Res, err error)

// Test runs the given use case with the when request and checks the result with then.
//
//	testdsl.Test(t, Greet{}, Hello{Name: "Gopher"}, testdsl.ThenReturns(Greeting{Text: "Hello, Gopher!"}))
func Test[Req, Res any](tb testing.TB, given interactor.UseCase[Req, Res], when Req, then Then[Res]) {
	tb.Helper()

	got, err := given.Run(tb.Context(), when)
	then(tb, got, err)
}

// ThenReturns expects the use case to succeed with want.
func ThenReturns[Res any](want Res) Then[Res] {
	return func(tb testing.TB, got Res, err error) {
		tb.Helper()

		require.NoError(tb, err)
		assert.Equal(tb, want, got)
	}
}

// ThenFails expects the use case to fail with an error matching want, as by [errors.Is].
func ThenFails[Res any](want error) Then[Res] {
	return func(tb testing.TB, _ Res, err error) {
		tb.Helper()

		require.ErrorIs(tb, err, want)
	}
}
