package interactor_test

import (
	"context"
	"reflect"
	"sync"
	"sync/atomic"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/screwyprof/interactor/v3"
)

func TestDispatch(t *testing.T) {
	t.Parallel()

	t.Run("returns the registered use case's response", func(t *testing.T) {
		t.Parallel()

		var d interactor.Dispatcher
		d.MustRegister(AnswerQuestion{})

		got, err := d.Dispatch[Answer](t.Context(), Question{ID: 21})

		require.NoError(t, err)
		assert.Equal(t, Answer{Value: 42}, got)
	})

	t.Run("returns the use case's error unchanged", func(t *testing.T) {
		t.Parallel()

		var d interactor.Dispatcher
		d.MustRegister(AnswerQuestion{err: errBoom})

		_, err := d.Dispatch[Answer](t.Context(), Question{})

		require.ErrorIs(t, err, errBoom)
	})

	t.Run("fails when nothing is registered for the request type", func(t *testing.T) {
		t.Parallel()

		var d interactor.Dispatcher

		_, err := d.Dispatch[Answer](t.Context(), Question{})

		require.ErrorIs(t, err, interactor.ErrUseCaseNotFound)
		assert.ErrorContains(t, err, "interactor_test.Question")
	})

	t.Run("matches request types exactly", func(t *testing.T) {
		t.Parallel()

		var d interactor.Dispatcher
		d.MustRegister(AnswerQuestion{})

		_, err := d.Dispatch[Answer](t.Context(), &Question{})

		require.ErrorIs(t, err, interactor.ErrUseCaseNotFound)
	})

	t.Run("routes by the request's static type, not one held in an interface", func(t *testing.T) {
		t.Parallel()

		var d interactor.Dispatcher
		d.MustRegister(AnswerQuestion{})

		var req any = Question{ID: 1}

		_, err := d.Dispatch[Answer](t.Context(), req)

		require.ErrorIs(t, err, interactor.ErrUseCaseNotFound)
	})

	t.Run("fails when asked for a different response type", func(t *testing.T) {
		t.Parallel()

		var d interactor.Dispatcher
		d.MustRegister(AnswerQuestion{})

		_, err := d.Dispatch[string](t.Context(), Question{})

		var typeErr *interactor.ResponseTypeError
		require.ErrorAs(t, err, &typeErr)
		assert.Equal(t, reflect.TypeFor[Question](), typeErr.Request)
		assert.Equal(t, reflect.TypeFor[Answer](), typeErr.Registered)
		assert.Equal(t, reflect.TypeFor[string](), typeErr.Requested)
		assert.EqualError(t, err,
			"interactor: use case for interactor_test.Question returns interactor_test.Answer, not string")
	})

	t.Run("rejects a nil context", func(t *testing.T) {
		t.Parallel()

		var d interactor.Dispatcher
		d.MustRegister(AnswerQuestion{})

		var nilCtx context.Context // a nil context is what is under test

		_, err := d.Dispatch[Answer](nilCtx, Question{})

		require.ErrorIs(t, err, interactor.ErrNilContext)
	})
}

func TestRegister(t *testing.T) {
	t.Parallel()

	t.Run("rejects a second use case for the same request type", func(t *testing.T) {
		t.Parallel()

		var d interactor.Dispatcher
		d.MustRegister(AnswerQuestion{})

		err := d.Register(AnswerQuestion{err: errBoom})

		require.ErrorIs(t, err, interactor.ErrAlreadyRegistered)

		got, err := d.Dispatch[Answer](t.Context(), Question{ID: 1})
		require.NoError(t, err, "the first registration must still serve")
		assert.Equal(t, Answer{Value: 2}, got)
	})

	nilUseCases := []struct {
		name    string
		useCase interactor.UseCase[Question, Answer]
	}{
		{name: "nil interface", useCase: nil},
		{name: "nil pointer", useCase: (*AnswerQuestion)(nil)},
		{name: "nil func", useCase: interactor.UseCaseFunc[Question, Answer](nil)},
		{name: "nil out-param func", useCase: interactor.FromOutParam[Question, Answer](nil)},
	}

	for _, tc := range nilUseCases {
		t.Run("rejects a "+tc.name, func(t *testing.T) {
			t.Parallel()

			var d interactor.Dispatcher

			err := d.Register(tc.useCase)

			require.ErrorIs(t, err, interactor.ErrNilUseCase)
			assert.ErrorContains(t, err, "interactor_test.Question", "the error names the registration")
		})
	}
}

func TestDispatcherIsSafeForConcurrentUse(t *testing.T) {
	t.Parallel()

	var (
		d          interactor.Dispatcher
		registered atomic.Int32
		wg         sync.WaitGroup
	)

	for range 10 {
		wg.Go(func() {
			if d.Register(AnswerQuestion{}) == nil {
				registered.Add(1)
			}
		})
		wg.Go(func() {
			got, err := d.Dispatch[Answer](t.Context(), Question{ID: 1})
			if err != nil {
				assert.ErrorIs(t, err, interactor.ErrUseCaseNotFound)

				return
			}

			assert.Equal(t, Answer{Value: 2}, got)
		})
	}

	wg.Wait()

	assert.Equal(t, int32(1), registered.Load(), "exactly one concurrent registration must win")
}

func BenchmarkDispatch(b *testing.B) {
	var d interactor.Dispatcher
	d.MustRegister(AnswerQuestion{})

	ctx := b.Context()

	if _, err := d.Dispatch[Answer](ctx, Question{ID: 1}); err != nil {
		b.Fatal(err)
	}

	b.ReportAllocs()

	for b.Loop() {
		_, _ = d.Dispatch[Answer](ctx, Question{ID: 1})
	}
}

// BenchmarkDispatchParallel is Dispatch under contention, as when every request goes through one
// Dispatcher.
func BenchmarkDispatchParallel(b *testing.B) {
	var d interactor.Dispatcher
	d.MustRegister(AnswerQuestion{})

	ctx := b.Context()

	b.ReportAllocs()

	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			_, _ = d.Dispatch[Answer](ctx, Question{ID: 1})
		}
	})
}

// BenchmarkDirect is the floor Dispatch is measured against: the same use case, called directly.
func BenchmarkDirect(b *testing.B) {
	uc := AnswerQuestion{}
	ctx := b.Context()

	b.ReportAllocs()

	for b.Loop() {
		_, _ = uc.Run(ctx, Question{ID: 1})
	}
}

func TestMustRegister(t *testing.T) {
	t.Parallel()

	var d interactor.Dispatcher
	d.MustRegister(AnswerQuestion{})

	assertPanicsWith(t, interactor.ErrAlreadyRegistered, func() { d.MustRegister(AnswerQuestion{}) })
	assertPanicsWith(
		t,
		interactor.ErrNilUseCase,
		func() { d.MustRegister(interactor.UseCase[Question, Answer](nil)) },
	)
}

// assertPanicsWith checks fn panics with an error matching want, as by errors.Is.
func assertPanicsWith(t *testing.T, want error, fn func()) {
	t.Helper()

	defer func() {
		t.Helper()

		err, ok := recover().(error)
		require.True(t, ok, "expected a panic with an error value")
		require.ErrorIs(t, err, want)
	}()

	fn()
}
