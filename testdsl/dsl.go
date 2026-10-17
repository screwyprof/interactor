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
// [Given.WithContext] runs the use case under a context of your own instead of the test's.
//
// It uses testify, which your tests then depend on too.
package testdsl

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/screwyprof/interactor/v3"
	"github.com/screwyprof/interactor/v3/internal/nilcheck"
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
	ctx     context.Context //nolint:containedctx // carried from WithContext to When, within one scenario.
}

// Given sets the use case under test. A nil use case fails the test.
func (s *Scenario) Given[Req, Res any](useCase interactor.UseCase[Req, Res]) *Given[Req, Res] {
	s.tb.Helper()

	if nilcheck.IsNil(useCase) {
		require.FailNow(s.tb, "testdsl: Given needs a use case, got nil")
	}

	return &Given[Req, Res]{tb: s.tb, useCase: useCase}
}

// WithContext runs the use case under ctx instead of the test's context, for use cases that read
// values such as the current user from it.
func (g *Given[Req, Res]) WithContext(ctx context.Context) *Given[Req, Res] {
	g.ctx = ctx

	return g
}

// When is a scenario whose use case has run, holding what it returned.
type When[Res any] struct {
	tb      testing.TB
	got     Res
	err     error
	checked bool
}

// When runs the use case with req, under the test's context unless WithContext set one. The
// scenario must end with one of the Then methods, or the test fails at the line that called When.
// Call When and Then from the test's goroutine, as for any test that can fail. testdsl is for
// tests: each scenario is kept until the test ends, so in benchmarks call the use case directly.
func (g *Given[Req, Res]) When(req Req) *When[Res] {
	g.tb.Helper()

	ctx := g.ctx
	if ctx == nil {
		ctx = g.tb.Context()
	}

	got, err := g.useCase.Run(ctx, req)
	w := &When[Res]{tb: g.tb, got: got, err: err}

	// The cleanup only needs the TB and the flag, not the result, so it keeps nothing large alive.
	// Marked as a helper, the testing package reports it at the line that called When, honouring
	// the caller's own t.Helper() marks.
	tb, checked := g.tb, &w.checked

	tb.Cleanup(func() {
		tb.Helper()

		if !*checked {
			tb.Errorf("testdsl: scenario has no Then, ThenFailWith or ThenCheck")
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
// want must not be nil.
func (w *When[Res]) ThenFailWith(want error) {
	w.tb.Helper()
	w.checked = true

	if want == nil {
		require.FailNow(w.tb, "testdsl: ThenFailWith needs the error to expect, got nil")
	}

	require.Error(w.tb, w.err)
	require.ErrorIs(w.tb, w.err, want)
}

// ThenCheck runs check on what the use case returned, for expectations Then and ThenFailWith do
// not cover.
func (w *When[Res]) ThenCheck(check func(tb testing.TB, got Res, err error)) {
	w.tb.Helper()
	w.checked = true

	if check == nil {
		require.FailNow(w.tb, "testdsl: ThenCheck needs a check, got nil")
	}

	check(w.tb, w.got, w.err)
}
