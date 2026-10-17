package interactor_test

import (
	"context"
	"errors"

	"github.com/screwyprof/interactor/v3"
)

var errBoom = errors.New("boom")

type Question struct{ ID int }

type Answer struct{ Value int }

// AnswerQuestion doubles the question's ID, or fails with err when set.
type AnswerQuestion struct{ err error }

func (uc AnswerQuestion) Run(_ context.Context, q Question) (Answer, error) {
	if uc.err != nil {
		return Answer{}, uc.err
	}

	return Answer{Value: q.ID * 2}, nil
}

// legacyAnswerQuestion is the same use case in the v2 shape, writing through a pointer.
func legacyAnswerQuestion(_ context.Context, q Question, res *Answer) error {
	if q.ID < 0 {
		return errBoom
	}

	res.Value = q.ID * 2

	return nil
}

// mustRegister fails the test at once if wiring fails, so a test reads as what it checks.
func mustRegister[Req, Res any](tb interface {
	Helper()
	Fatalf(format string, args ...any)
}, d *interactor.Dispatcher, uc interactor.UseCase[Req, Res],
) {
	tb.Helper()

	if err := d.Register(uc); err != nil {
		tb.Fatalf("register %T: %v", uc, err)
	}
}
