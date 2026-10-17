// Package testdsl is a Given/When/Then DSL for testing your own use cases.
//
// Import it from your tests when you want each test to read as a scenario:
//
//	func TestGreet(t *testing.T) {
//		testdsl.Test(t).
//			Given(Greet{}).
//			When(t.Context(), Hello{Name: "Gopher"}).
//			Then(Greeting{Text: "Hello, Gopher!"})
//	}
//
// Given fixes the request and response types for the rest of the chain, so When only accepts the
// use case's request and Then only its response: a mismatched expectation does not compile.
// [When.ThenFailWith] expects an error instead, and [When.ThenCheck] runs a check of your own.
//
// Like any assertion library, testdsl checks only what a scenario asks for: end each one with a
// Then method. It uses testify, which your tests then depend on too.
package testdsl

import (
	"context"
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

// Given sets the use case under test.
func (s *Scenario) Given[Req, Res any](useCase interactor.UseCase[Req, Res]) *Given[Req, Res] {
	return &Given[Req, Res]{tb: s.tb, useCase: useCase}
}

// When is a scenario whose use case has run, holding what it returned.
type When[Res any] struct {
	tb  testing.TB
	got Res
	err error
}

// When runs the use case with ctx and req. Call the Then methods from the test's goroutine, as
// for any test that can fail.
func (g *Given[Req, Res]) When(ctx context.Context, req Req) *When[Res] {
	got, err := g.useCase.Run(ctx, req)

	return &When[Res]{tb: g.tb, got: got, err: err}
}

// Then expects the use case to have succeeded and returned want.
func (w *When[Res]) Then(want Res) {
	w.tb.Helper()

	require.NoError(w.tb, w.err)
	assert.Equal(w.tb, want, w.got)
}

// ThenFailWith expects the use case to have failed with an error matching want, as by [errors.Is].
// A nil want never passes: the use case must have failed.
func (w *When[Res]) ThenFailWith(want error) {
	w.tb.Helper()

	require.Error(w.tb, w.err)
	require.ErrorIs(w.tb, w.err, want)
}

// ThenCheck runs check on what the use case returned, for expectations Then and ThenFailWith do
// not cover.
func (w *When[Res]) ThenCheck(check func(tb testing.TB, got Res, err error)) {
	w.tb.Helper()

	check(w.tb, w.got, w.err)
}
