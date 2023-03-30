package interactor

import (
	"context"
	"fmt"
	"reflect"
)

// Func is a helper function that converts a function with the appropriate signature into a UseCaseRunnerFn.
//
// The function must have the following signature:
//
//  1. Have 3 arguments:
//
//     - ctx context.Context,
//
//     - req a struct which implements Request interface,
//
//     - res a pointer to a struct which implements Response interface.
//
//  2. Return an error
//
// An example signature may look like as follows:
//
//	func(ctx context.Context, req TestRequest, res *TestResponse) error
func Func(fn interface{}) (UseCaseRunnerFn, error) {
	useCaseRunnerType := reflect.TypeOf(fn)
	if err := ensureSignatureIsValid(useCaseRunnerType); err != nil {
		return nil, err
	}

	return func(ctx context.Context, req Request, resp Response) error {
		if err := ensureResultHasValidType(useCaseRunnerType, resp); err != nil {
			return err
		}

		return invokeUseCaseRunner(fn, ctx, req, resp)
	}, nil
}

// Must is a wrapper around Func which panics if an error occurs.
//
// It is useful for tests and for cases where you are sure that the signature is valid.
// You can use it like this:
//
//	interactor.Must(interactor.Func(userCaseRunner.run))
func Must(fn UseCaseRunnerFn, err error) UseCaseRunnerFn {
	if err != nil {
		panic(err)
	}

	return fn
}

// Adapt is a helper function that converts a struct with a Run method into a UseCaseRunnerFn.
//
// The function `Run` must have the following signature:
//
//  1. Have 3 arguments:
//
//     - ctx context.Context,
//
//     - req a struct which implements Request interface,
//
//     - res a pointer to a struct which implements Response interface.
//
//  2. Return an error
//
// An example signature may look like as follows:
//
//	func (uc *UseCase) Run(ctx context.Context, req TestRequest, res *TestResponse) error
func Adapt(runner interface{}) (UseCaseRunnerFn, error) {
	fnValue := reflect.ValueOf(runner).MethodByName("Run")
	if !fnValue.IsValid() {
		return nil, fmt.Errorf("%w", ErrUseCaseRunnerHasNoRunMethod)
	}

	return Func(fnValue.Interface())
}

// MustAdapt is a wrapper around Adapt which panics if an error occurs.
//
// It is useful for tests and for cases where you are sure that the runner has a `Ru`n method with a valid signature.
// You can use it like this:
//
//	interactor.MustAdapt(userCaseRunner)
func MustAdapt(fn interface{}) UseCaseRunnerFn {
	return Must(Adapt(fn))
}

func ensureSignatureIsValid(useCaseRunnerType reflect.Type) error {
	if useCaseRunnerType.Kind() != reflect.Func {
		return fmt.Errorf("%w: %s", ErrUseCaseRunnerIsNotAFunction, useCaseRunnerType.String())
	}

	if num := useCaseRunnerType.NumIn(); num != 3 {
		return fmt.Errorf("%w: %d given", ErrInvalidUseCaseRunnerSignature, num)
	}

	if err := ensureParamsHaveValidTypes(useCaseRunnerType); err != nil {
		return fmt.Errorf("%w", err)
	}

	return nil
}

func ensureResultHasValidType(runnerType reflect.Type, res interface{}) error {
	want := runnerType.In(2)
	got := reflect.TypeOf(res)

	if got != want {
		return fmt.Errorf("%w: want %v, got %v", ErrResultTypeMismatch, want, got)
	}

	return nil
}

func ensureParamsHaveValidTypes(useCaseRunnerType reflect.Type) error {
	if !firstArgIsContext(useCaseRunnerType) {
		return fmt.Errorf("%w: %s given", ErrFirstArgHasInvalidType, useCaseRunnerType.In(0).String())
	}

	if !secondArgIsRequest(useCaseRunnerType) {
		return fmt.Errorf("%w: %s given", ErrSecondArgHasInvalidType, useCaseRunnerType.In(1).String())
	}

	if !thirdArgIsResponse(useCaseRunnerType) {
		return fmt.Errorf("%w: %s given", ErrThirdArgHasInvalidType, useCaseRunnerType.In(2).String())
	}

	return nil
}

func firstArgIsContext(useCaseRunnerType reflect.Type) bool {
	ctxtInterface := reflect.TypeOf((*context.Context)(nil)).Elem()
	ctx := useCaseRunnerType.In(0)

	return ctx.Implements(ctxtInterface)
}

func secondArgIsRequest(useCaseRunnerType reflect.Type) bool {
	requestInterface := reflect.TypeOf((*Request)(nil)).Elem()

	secondArg := useCaseRunnerType.In(1)
	if secondArg.Kind() == reflect.Ptr {
		secondArg = secondArg.Elem()
	}

	return secondArg.Kind() == reflect.Struct && secondArg.Implements(requestInterface)
}

func thirdArgIsResponse(useCaseRunnerType reflect.Type) bool {
	responseInterface := reflect.TypeOf((*Response)(nil)).Elem()
	thirdArg := useCaseRunnerType.In(2)

	return thirdArg.Kind() == reflect.Ptr && thirdArg.Implements(responseInterface)
}

func invokeUseCaseRunner(useCaseRunner interface{}, args ...interface{}) error {
	result := invoke(useCaseRunner, args...)

	if err, ok := result[0].Interface().(error); ok && err != nil {
		return err
	}

	return nil
}

func invoke(fn interface{}, args ...interface{}) []reflect.Value {
	v := reflect.ValueOf(fn)

	in := make([]reflect.Value, len(args))
	for i, arg := range args {
		in[i] = reflect.ValueOf(arg)
	}

	return v.Call(in)
}
