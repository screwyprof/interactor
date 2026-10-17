package testdsl_test

import (
	"context"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"runtime"
	"runtime/debug"
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
	skipped  bool            // SkipNow was called
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

func (f *fakeTB) SkipNow() {
	f.skipped = true

	runtime.Goexit()
}

func (f *fakeTB) Skipped() bool { return f.skipped }

func (f *fakeTB) Cleanup(fn func()) { f.cleanups = append(f.cleanups, fn) }

func (f *fakeTB) Context() context.Context { return f.ctx }

// run runs scenario against a fakeTB on its own goroutine, as testing.T would, then its cleanups.
// Waiting for the goroutine orders its writes to f before the reads here, so f needs no lock. A
// panic is kept apart from the failures, so a test expecting the DSL to fail cannot pass because
// it crashed instead.
func run(t *testing.T, scenario func(tb testing.TB)) *fakeTB {
	t.Helper()

	f := &fakeTB{ctx: t.Context()}

	var (
		wg       sync.WaitGroup
		panicked string
	)

	wg.Go(func() {
		defer func() {
			if r := recover(); r != nil {
				panicked = fmt.Sprintf("%v\n%s", r, debug.Stack())
			}
		}()

		scenario(f)
	})
	wg.Wait()

	require.Empty(t, panicked, "the scenario panicked instead of failing")

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

	t.Run("is not reported when the test was skipped", func(t *testing.T) {
		t.Parallel()

		f := run(t, func(tb testing.TB) {
			tb.Helper()

			testdsl.Test(tb).Given(Pinger{}).When(Ping{})
			tb.SkipNow()
		})

		assert.True(t, f.skipped)
		assert.Empty(t, f.errors)
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

			f := run(t, func(tb testing.TB) {
				tb.Helper()

				testdsl.Test(tb).Given(useCase).When(Ping{}).Then(Pong{OK: true})
			})

			assert.True(t, f.stopped)
			require.Len(t, f.errors, 1)
			assert.Contains(t, f.errors[0], "testdsl: Given needs a use case, got nil")
		})
	}
}

func TestWhenContext(t *testing.T) {
	t.Parallel()

	derived := context.WithoutCancel(t.Context())

	var nilCtx context.Context // a nil context is one of the cases under test

	cases := map[string]struct {
		given func(g *testdsl.Given[Ping, Pong]) *testdsl.Given[Ping, Pong]
		want  func(f *fakeTB) context.Context
	}{
		"is the test's context by default": {
			given: func(g *testdsl.Given[Ping, Pong]) *testdsl.Given[Ping, Pong] { return g },
			want:  func(f *fakeTB) context.Context { return f.ctx },
		},
		"is the one given WithContext": {
			given: func(g *testdsl.Given[Ping, Pong]) *testdsl.Given[Ping, Pong] { return g.WithContext(derived) },
			want:  func(*fakeTB) context.Context { return derived },
		},
		"is nil when WithContext was given nil": {
			given: func(g *testdsl.Given[Ping, Pong]) *testdsl.Given[Ping, Pong] { return g.WithContext(nilCtx) },
			want:  func(*fakeTB) context.Context { return nil },
		},
	}

	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			var gotCtx context.Context

			f := run(t, func(tb testing.TB) {
				tb.Helper()

				tc.given(testdsl.Test(tb).Given(Pinger{gotCtx: &gotCtx})).When(Ping{}).Then(Pong{OK: true})
			})

			assert.False(t, f.Failed())
			assert.Equal(t, tc.want(f), gotCtx)
		})
	}
}

func TestWithContextLeavesTheGivenUnchanged(t *testing.T) {
	t.Parallel()

	var gotCtx context.Context

	other := context.WithoutCancel(t.Context())

	f := run(t, func(tb testing.TB) {
		tb.Helper()

		given := testdsl.Test(tb).Given(Pinger{gotCtx: &gotCtx})
		given.WithContext(other).When(Ping{}).Then(Pong{OK: true})
		given.When(Ping{}).Then(Pong{OK: true})
	})

	assert.False(t, f.Failed())
	assert.Equal(t, f.ctx, gotCtx, "the second When must run under the test's context again")
}

// TestMissingThenIsReportedAtTheCallersLine re-runs this test binary to run
// TestMissingThenInRealTest, because only a real *testing.T shows which line the report lands on.
func TestMissingThenIsReportedAtTheCallersLine(t *testing.T) {
	t.Parallel()

	source, err := os.ReadFile("dsl_test.go")
	require.NoError(t, err)

	// The marker appears twice: in this lookup, and on the scenario line it marks.
	const marker = "// the scenario the report must point at"
	require.Equal(t, 2, strings.Count(string(source), marker), "marker comment missing or duplicated")

	at := strings.LastIndex(string(source), marker)

	line := strings.Count(string(source[:at]), "\n") + 1

	// Only -test.run: the parent's own flags, such as -test.coverprofile, must not reach the child.
	child := exec.CommandContext(t.Context(), os.Args[0], "-test.run=^TestMissingThenInRealTest$", "-test.count=1")

	child.Env = append(os.Environ(), "INTERACTOR_TESTDSL_MISSING_THEN_CHILD=1")

	out, err := child.CombinedOutput()

	require.Error(t, err, "the child test must fail")
	assert.Contains(t, string(out), fmt.Sprintf("dsl_test.go:%d: testdsl: scenario has no Then", line), string(out))
}

func TestMissingThenInRealTest(t *testing.T) {
	t.Parallel()

	if os.Getenv("INTERACTOR_TESTDSL_MISSING_THEN_CHILD") != "1" {
		t.Skip("run by TestMissingThenIsReportedAtTheCallersLine")
	}

	testdsl.Test(t).Given(Pinger{}).When(Ping{}) // the scenario the report must point at
}
