package testdsl_test

import (
	"context"
	"errors"
	"fmt"
	"runtime"
	"runtime/debug"
	"sync"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

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

	errors  []string // every failure message, in order
	stopped bool     // FailNow was called
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

// run runs scenario against a fakeTB on its own goroutine, as testing.T would. Waiting for the
// goroutine orders its writes to f before the reads here, so f needs no lock. A panic is kept
// apart from the failures, so a test expecting the DSL to fail cannot pass because it crashed.
func run(t *testing.T, scenario func(tb testing.TB)) *fakeTB {
	t.Helper()

	f := &fakeTB{}

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

	return f
}

// failed runs scenario and reports whether it failed.
func failed(t *testing.T, scenario func(tb testing.TB)) bool {
	t.Helper()

	return run(t, scenario).Failed()
}

func TestThen(t *testing.T) {
	t.Parallel()

	cases := map[string]struct {
		when Ping
		then Pong
		fail bool
	}{
		"passes on the expected response": {when: Ping{}, then: Pong{OK: true}},
		"fails on a different response":   {when: Ping{}, then: Pong{OK: false}, fail: true},
		"fails when the use case errors":  {when: Ping{Fail: true}, then: Pong{}, fail: true},
	}

	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			assert.Equal(t, tc.fail, failed(t, func(tb testing.TB) {
				tb.Helper()

				testdsl.Test(tb).Given(Pinger{}).When(t.Context(), tc.when).Then(tc.then)
			}))
		})
	}
}

func TestThenFailWith(t *testing.T) {
	t.Parallel()

	cases := map[string]struct {
		when Ping
		want error
		fail bool
	}{
		"passes on the expected error": {when: Ping{Fail: true}, want: errBoom},
		"fails on a different error": {
			when: Ping{Fail: true},
			want: errors.New("other"),
			fail: true,
		},
		"fails when the use case succeeds":              {when: Ping{}, want: errBoom, fail: true},
		"fails when the use case succeeds and want nil": {when: Ping{}, want: nil, fail: true},
		"fails when the use case errors and want nil":   {when: Ping{Fail: true}, want: nil, fail: true},
	}

	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			assert.Equal(t, tc.fail, failed(t, func(tb testing.TB) {
				tb.Helper()

				testdsl.Test(tb).Given(Pinger{}).When(t.Context(), tc.when).ThenFailWith(tc.want)
			}))
		})
	}
}

func TestThenCheck(t *testing.T) {
	t.Parallel()

	var (
		gotTB  testing.TB
		gotRes Pong
	)

	f := run(t, func(tb testing.TB) {
		tb.Helper()

		testdsl.Test(tb).Given(Pinger{}).When(t.Context(), Ping{}).ThenCheck(func(tb testing.TB, res Pong, _ error) {
			tb.Helper()

			gotTB, gotRes = tb, res
		})
	})

	assert.False(t, f.Failed())
	assert.Same(t, f, gotTB, "the check gets the scenario's TB")
	assert.Equal(t, Pong{OK: true}, gotRes)
}

func TestWhenRunsTheUseCaseUnderTheGivenContext(t *testing.T) {
	t.Parallel()

	var nilCtx context.Context // a nil context is passed through like any other

	cases := map[string]context.Context{
		"a context":     context.WithoutCancel(t.Context()),
		"a nil context": nilCtx,
	}

	for name, ctx := range cases {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			gotCtx := t.Context()

			f := run(t, func(tb testing.TB) {
				tb.Helper()

				testdsl.Test(tb).Given(Pinger{gotCtx: &gotCtx}).When(ctx, Ping{}).Then(Pong{OK: true})
			})

			assert.False(t, f.Failed())
			assert.Equal(t, ctx, gotCtx)
		})
	}
}
