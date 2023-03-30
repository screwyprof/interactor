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
