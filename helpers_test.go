package interactor_test

import (
	"context"
	"errors"
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

// Ticket is a distinct request type per K, so tests can register many types.
type Ticket[K any] struct{ N int }

// Stamp answers a Ticket with itself.
type Stamp[K any] struct{}

func (Stamp[K]) Run(_ context.Context, t Ticket[K]) (Ticket[K], error) { return t, nil }
