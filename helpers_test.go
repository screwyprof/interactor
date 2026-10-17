package interactor_test

import (
	"context"
	"errors"
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

func (i ConcreteUseCase) Run(_ context.Context, req TestRequest, res *TestResponse) error {
	res.result = req.id

	return i.err
}

type AnotherResponse struct{}

// customContext implements context.Context but is not context.Context itself.
type customContext struct{ context.Context } //nolint:containedctx // a type implementing context.Context is the point.

// extendedContext asks for more than context.Context provides.
type extendedContext interface {
	context.Context
	Extra()
}

type InvalidUseCaseWrongSignature struct{}

func (i InvalidUseCaseWrongSignature) Run(_ context.Context, _ TestRequest) error {
	return nil
}

type InvalidUseCaseWrongContext struct{}

func (i InvalidUseCaseWrongContext) Run(_ struct{}, _ TestRequest, _ *TestResponse) error {
	return nil
}

type InvalidUseCaseWrongRequest struct{}

func (i InvalidUseCaseWrongRequest) Run(_ context.Context, _ int, _ *TestResponse) error {
	return nil
}

type ValidUseCasePointerRequest struct{}

func (i ValidUseCasePointerRequest) Run(_ context.Context, _ *TestRequest, _ *TestResponse) error {
	return nil
}

type InvalidUseCaseWrongResponse struct{}

func (i InvalidUseCaseWrongResponse) Run(_ context.Context, _ TestRequest, _ struct{}) error {
	return nil
}
