// Package testdsl is a Given/When/Then DSL for testing your own use cases.
//
// Import it from your tests when you want each test to read as a scenario:
//
//	func TestGreet(t *testing.T) {
//		testdsl.Test(t).
//			Given(Greet{}).
//			When(Hello{Name: "Gopher"}).
//			Then(Greeting{Text: "Hello, Gopher!"})
//	}
//
// Given fixes the request and response types for the rest of the chain, so When only accepts the
// use case's request and Then only its response: a mismatched expectation does not compile.
// [When.ThenFailWith] expects an error instead, and [When.ThenCheck] runs a check of your own.
//
// It uses testify, which your tests then depend on too.
package testdsl

import (
	"reflect"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/screwyprof/interactor/v3"
)

// Scenario starts a test. Create one with [Test].
type Scenario struct {
	tb testing.TB
}

// Test starts a scenario reported to tb.
func Test(tb testing.TB) *Scenario {
	tb.Helper()

	return &Scenario{tb: tb}
}

// Given is a scenario with the use case under test.
type Given[Req, Res any] struct {
	tb      testing.TB
	useCase interactor.UseCase[Req, Res]
}

// Given sets the use case under test. A nil use case fails the test.
func (s *Scenario) Given[Req, Res any](useCase interactor.UseCase[Req, Res]) *Given[Req, Res] {
	s.tb.Helper()

	if isNil(useCase) {
		require.FailNow(s.tb, "testdsl: Given needs a use case, got nil")
	}

	return &Given[Req, Res]{tb: s.tb, useCase: useCase}
}

// When is a scenario whose use case has run, holding what it returned.
type When[Res any] struct {
	tb      testing.TB
	got     Res
	err     error
	checked bool
}

// When runs the use case with req under the test's context. The scenario must end with one of the
// Then methods; the test fails if it does not.
func (g *Given[Req, Res]) When(req Req) *When[Res] {
	g.tb.Helper()

	got, err := g.useCase.Run(g.tb.Context(), req)
	w := &When[Res]{tb: g.tb, got: got, err: err}

	g.tb.Cleanup(func() {
		if !w.checked {
			g.tb.Errorf("testdsl: scenario has no Then, ThenFailWith or ThenCheck")
		}
	})

	return w
}

// Then expects the use case to have succeeded and returned want.
func (w *When[Res]) Then(want Res) {
	w.tb.Helper()
	w.checked = true

	require.NoError(w.tb, w.err)
	assert.Equal(w.tb, want, w.got)
}

// ThenFailWith expects the use case to have failed with an error matching want, as by [errors.Is].
func (w *When[Res]) ThenFailWith(want error) {
	w.tb.Helper()
	w.checked = true

	require.Error(w.tb, w.err)
	require.ErrorIs(w.tb, w.err, want)
}

// ThenCheck runs check on what the use case returned, for expectations Then and ThenFailWith do
// not cover.
func (w *When[Res]) ThenCheck(check func(tb testing.TB, got Res, err error)) {
	w.tb.Helper()
	w.checked = true

	check(w.tb, w.got, w.err)
}

// isNil mirrors the dispatcher's check: a nil interface, nil pointer or nil func.
func isNil(useCase any) bool {
	if useCase == nil {
		return true
	}

	v := reflect.ValueOf(useCase)

	return (v.Kind() == reflect.Pointer || v.Kind() == reflect.Func) && v.IsNil()
}
