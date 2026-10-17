package testdsl_test

import (
	"context"
	"errors"
	"fmt"
	"runtime"
	"slices"
	"strings"
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
	testing.TB // methods the DSL must not call panic through the nil interface

	ctx      context.Context //nolint:containedctx // a fake TB has to return a context from Context().
	errors   []string        // every failure message, in order
	stopped  bool            // FailNow was called
	cleanups []func()
}

func (f *fakeTB) Helper() {}

func (f *fakeTB) Name() string { return "fake" }

func (f *fakeTB) Errorf(format string, args ...any) {
	f.errors = append(f.errors, fmt.Sprintf(format, args...))
}

func (f *fakeTB) FailNow() {
	f.stopped = true

	runtime.Goexit()
}

func (f *fakeTB) Failed() bool { return len(f.errors) > 0 || f.stopped }

func (f *fakeTB) Cleanup(fn func()) { f.cleanups = append(f.cleanups, fn) }

func (f *fakeTB) Context() context.Context { return f.ctx }

// run runs scenario against a fakeTB on its own goroutine, as testing.T would, then its cleanups.
// Waiting for the goroutine orders its writes to f before the reads here, so f needs no lock.
func run(t *testing.T, scenario func(tb testing.TB)) *fakeTB {
	t.Helper()

	f := &fakeTB{ctx: t.Context()}

	var wg sync.WaitGroup

	wg.Go(func() { scenario(f) })
	wg.Wait()

	for _, cleanup := range slices.Backward(f.cleanups) {
		cleanup()
	}

	return f
}

// failed runs scenario and reports whether it failed.
func failed(t *testing.T, scenario func(tb testing.TB)) bool {
	t.Helper()

	return run(t, scenario).Failed()
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
}

func TestThenCheck(t *testing.T) {
	t.Parallel()

	t.Run("hands over the scenario's TB and the typed result", func(t *testing.T) {
		t.Parallel()

		var (
			gotTB  testing.TB
			gotRes Pong
		)

		f := run(t, func(tb testing.TB) {
			tb.Helper()

			testdsl.Test(tb).Given(Pinger{}).When(Ping{}).ThenCheck(func(tb testing.TB, res Pong, _ error) {
				tb.Helper()

				gotTB, gotRes = tb, res
			})
		})

		assert.False(t, f.Failed())
		assert.Same(t, f, gotTB)
		assert.Equal(t, Pong{OK: true}, gotRes)
	})
}

func TestThenRejectsANilArgument(t *testing.T) {
	t.Parallel()

	cases := map[string]struct {
		then func(w *testdsl.When[Pong])
		want string
	}{
		"ThenFailWith": {func(w *testdsl.When[Pong]) { w.ThenFailWith(nil) }, "ThenFailWith needs the error to expect"},
		"ThenCheck":    {func(w *testdsl.When[Pong]) { w.ThenCheck(nil) }, "ThenCheck needs a check"},
	}

	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			f := run(t, func(tb testing.TB) {
				tb.Helper()

				tc.then(testdsl.Test(tb).Given(Pinger{}).When(Ping{Fail: true}))
			})

			assert.True(t, f.stopped)
			assert.Contains(t, strings.Join(f.errors, "\n"), tc.want)
		})
	}
}

func TestFailingThenReportsOnce(t *testing.T) {
	t.Parallel()

	f := run(t, func(tb testing.TB) {
		tb.Helper()

		testdsl.Test(tb).Given(Pinger{}).When(Ping{}).Then(Pong{OK: false})
	})

	require.Len(t, f.errors, 1, "a failing Then must not also report a missing Then")
}

func TestMissingThen(t *testing.T) {
	t.Parallel()

	const missing = "testdsl: scenario has no Then, ThenFailWith or ThenCheck"

	t.Run("is reported", func(t *testing.T) {
		t.Parallel()

		f := run(t, func(tb testing.TB) {
			tb.Helper()

			testdsl.Test(tb).Given(Pinger{}).When(Ping{})
		})

		assert.Equal(t, []string{missing}, f.errors)
	})

	t.Run("is reported even after an earlier failure", func(t *testing.T) {
		t.Parallel()

		f := run(t, func(tb testing.TB) {
			tb.Helper()

			tb.Errorf("earlier failure")
			testdsl.Test(tb).Given(Pinger{}).When(Ping{})
		})

		assert.Equal(t, []string{"earlier failure", missing}, f.errors)
	})

	t.Run("is reported when the test stopped before its Then", func(t *testing.T) {
		t.Parallel()

		f := run(t, func(tb testing.TB) {
			tb.Helper()

			testdsl.Test(tb).Given(Pinger{}).When(Ping{})
			tb.FailNow()
		})

		assert.Equal(t, []string{missing}, f.errors)
	})
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

func TestWhenContext(t *testing.T) {
	t.Parallel()

	t.Run("is the test's context by default", func(t *testing.T) {
		t.Parallel()

		var gotCtx context.Context

		f := run(t, func(tb testing.TB) {
			tb.Helper()

			testdsl.Test(tb).
				Given(Pinger{gotCtx: &gotCtx}).
				When(Ping{}). //nolint:contextcheck // which context reaches the use case is under test.
				Then(Pong{OK: true})
		})

		assert.False(t, f.Failed())
		assert.Same(t, f.ctx, gotCtx)
	})

	t.Run("is the one given WithContext", func(t *testing.T) {
		t.Parallel()

		var gotCtx context.Context

		ctx := context.WithValue(t.Context(), ctxKey{}, "marker")

		f := run(t, func(tb testing.TB) {
			tb.Helper()

			testdsl.Test(tb).
				Given(Pinger{gotCtx: &gotCtx}).
				WithContext(ctx).
				When(Ping{}). //nolint:contextcheck // which context reaches the use case is under test.
				Then(Pong{OK: true})
		})

		assert.False(t, f.Failed())
		require.NotNil(t, gotCtx)
		assert.Equal(t, "marker", gotCtx.Value(ctxKey{}))
	})
}

type ctxKey struct{}
