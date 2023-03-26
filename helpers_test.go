package interactor_test

import (
	"context"
	"errors"

	"github.com/screwyprof/interactor"
)

var errSomeErr = errors.New("some error")

type TestRequest struct {
	id int
}

type TestResponse struct {
	result int
}

type ConcreteUseCase struct {
	err error
}

func (i ConcreteUseCase) RunUseCase(_ context.Context, req TestRequest, res *TestResponse) error {
	res.result = req.id

	return i.err
}

type GeneralUseCaseSpy struct {
	wasCalled bool
}

func (s *GeneralUseCaseSpy) Run(_ context.Context, _ interactor.Request, _ interactor.Response) error {
	s.wasCalled = true

	return nil
}
