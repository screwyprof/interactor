package testdsl_test

import (
	"context"
	"errors"
	"runtime"
	"slices"
	"sync"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/screwyprof/interactor/v3"
	"github.com/screwyprof/interactor/v3/testdsl"
)

var errBoom = errors.New("boom")

type Ping struct{ Fail bool }

type Pong struct{ OK bool }

// Pinger answers a Ping, or fails with errBoom when asked to; it records the context it ran with.
type Pinger struct{ gotCtx *context.Context }

func (p Pinger) Run(ctx context.Context, req Ping) (Pong, error) {
	if p.gotCtx != nil {
		*p.gotCtx = ctx
	}

	if req.Fail {
		return Pong{}, errBoom
	}

	return Pong{OK: true}, nil
}

// fakeTB records failures instead of failing the real test, so the DSL's own failures can be
// asserted. FailNow ends the scenario's goroutine as testing.T does.
type fakeTB struct {
	testing.TB

	ctx      context.Context //nolint:containedctx // a fake TB has to return a context from Context().
	mu       sync.Mutex
	failed   bool
	cleanups []func()
}

func (f *fakeTB) Helper() {}

func (f *fakeTB) Name() string { return "fake" }

func (f *fakeTB) Errorf(string, ...any) {
	f.mu.Lock()
	defer f.mu.Unlock()

	f.failed = true
}

func (f *fakeTB) FailNow() {
	f.Errorf("")
	runtime.Goexit()
}

func (f *fakeTB) Cleanup(fn func()) { f.cleanups = append(f.cleanups, fn) }

func (f *fakeTB) Context() context.Context { return f.ctx }

// failed runs scenario against a fakeTB, then its cleanups, and reports whether it failed.
func failed(t *testing.T, scenario func(tb testing.TB)) bool {
	t.Helper()

	f := &fakeTB{ctx: t.Context()}
	done := make(chan struct{})

	go func() {
		defer close(done)

		scenario(f)
	}()

	<-done

	for _, cleanup := range slices.Backward(f.cleanups) {
		cleanup()
	}

	return f.failed
}

func TestThen(t *testing.T) {
	t.Parallel()

	t.Run("passes on the expected response", func(t *testing.T) {
		t.Parallel()

		assert.False(t, failed(t, func(tb testing.TB) {
			tb.Helper()

			testdsl.Test(tb).Given(Pinger{}).When(Ping{}).Then(Pong{OK: true})
		}))
	})

	t.Run("fails on a different response", func(t *testing.T) {
		t.Parallel()

		assert.True(t, failed(t, func(tb testing.TB) {
			tb.Helper()

			testdsl.Test(tb).Given(Pinger{}).When(Ping{}).Then(Pong{OK: false})
		}))
	})

	t.Run("fails when the use case errors", func(t *testing.T) {
		t.Parallel()

		assert.True(t, failed(t, func(tb testing.TB) {
			tb.Helper()

			testdsl.Test(tb).Given(Pinger{}).When(Ping{Fail: true}).Then(Pong{})
		}))
	})
}

func TestThenFailWith(t *testing.T) {
	t.Parallel()

	t.Run("passes on the expected error", func(t *testing.T) {
		t.Parallel()

		assert.False(t, failed(t, func(tb testing.TB) {
			tb.Helper()

			testdsl.Test(tb).Given(Pinger{}).When(Ping{Fail: true}).ThenFailWith(errBoom)
		}))
	})

	t.Run("fails on a different error", func(t *testing.T) {
		t.Parallel()

		assert.True(t, failed(t, func(tb testing.TB) {
			tb.Helper()

			testdsl.Test(tb).Given(Pinger{}).When(Ping{Fail: true}).ThenFailWith(errors.New("other"))
		}))
	})

	t.Run("fails when the use case succeeds", func(t *testing.T) {
		t.Parallel()

		assert.True(t, failed(t, func(tb testing.TB) {
			tb.Helper()

			testdsl.Test(tb).Given(Pinger{}).When(Ping{}).ThenFailWith(errBoom)
		}))
	})

	t.Run("fails when the use case succeeds, even for a nil expected error", func(t *testing.T) {
		t.Parallel()

		assert.True(t, failed(t, func(tb testing.TB) {
			tb.Helper()

			testdsl.Test(tb).Given(Pinger{}).When(Ping{}).ThenFailWith(nil)
		}))
	})
}

func TestThenCheck(t *testing.T) {
	t.Parallel()

	var got Pong

	assert.False(t, failed(t, func(tb testing.TB) {
		tb.Helper()

		testdsl.Test(tb).Given(Pinger{}).When(Ping{}).ThenCheck(func(_ testing.TB, res Pong, err error) {
			got = res

			require.NoError(tb, err)
		})
	}))
	assert.Equal(t, Pong{OK: true}, got)
}

func TestGivenRejectsANilUseCase(t *testing.T) {
	t.Parallel()

	nilUseCases := map[string]interactor.UseCase[Ping, Pong]{
		"nil interface": nil,
		"nil pointer":   (*Pinger)(nil),
		"nil func":      interactor.UseCaseFunc[Ping, Pong](nil),
	}

	for name, useCase := range nilUseCases {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			assert.True(t, failed(t, func(tb testing.TB) {
				tb.Helper()

				testdsl.Test(tb).Given(useCase).When(Ping{}).Then(Pong{OK: true})
			}))
		})
	}
}

func TestScenarioWithoutThenFails(t *testing.T) {
	t.Parallel()

	assert.True(t, failed(t, func(tb testing.TB) {
		tb.Helper()

		testdsl.Test(tb).Given(Pinger{}).When(Ping{})
	}))
}

func TestWhenRunsUnderTheTestContext(t *testing.T) {
	t.Parallel()

	var gotCtx context.Context

	tb := &fakeTB{ctx: context.WithValue(t.Context(), ctxKey{}, "marker")}

	testdsl.Test(tb).Given(Pinger{gotCtx: &gotCtx}).When(Ping{}).Then(Pong{OK: true})

	require.NotNil(t, gotCtx)
	assert.Equal(t, "marker", gotCtx.Value(ctxKey{}))
}

type ctxKey struct{}
