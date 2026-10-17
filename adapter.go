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
func Func(fn any) (UseCaseRunnerFn, error) {
	return newRunner(reflect.ValueOf(fn))
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
//     - ctx context.Context,
//     - req a struct which implements Request interface,
//     - res a pointer to a struct which implements Response interface.
//
//  2. Return an error
//
// An example signature may look like as follows:
//
//	func (uc *UseCase) Run(ctx context.Context, req TestRequest, res *TestResponse) error
func Adapt(runner any) (UseCaseRunnerFn, error) {
	if runner == nil {
		return nil, fmt.Errorf("%w: nil given", ErrUseCaseRunnerHasNoRunMethod)
	}

	fnValue := reflect.ValueOf(runner).MethodByName("Run")
	if !fnValue.IsValid() {
		return nil, fmt.Errorf("%w", ErrUseCaseRunnerHasNoRunMethod)
	}

	return newRunner(fnValue)
}

// MustAdapt is a wrapper around Adapt which panics if an error occurs.
//
// It is useful for tests and for cases where you are sure that the runner has a `Ru`n method with a valid signature.
// You can use it like this:
//
//	interactor.MustAdapt(userCaseRunner)
func MustAdapt(fn any) UseCaseRunnerFn {
	return Must(Adapt(fn))
}

func ensureSignatureIsValid(useCaseRunnerType reflect.Type) error {
	if useCaseRunnerType == nil {
		return fmt.Errorf("%w: nil given", ErrUseCaseRunnerIsNotAFunction)
	}

	if useCaseRunnerType.Kind() != reflect.Func {
		return fmt.Errorf("%w: %s", ErrUseCaseRunnerIsNotAFunction, useCaseRunnerType.String())
	}

	if num := useCaseRunnerType.NumIn(); num != 3 {
		return fmt.Errorf("%w: %d given", ErrInvalidUseCaseRunnerSignature, num)
	}

	if err := ensureParamsHaveValidTypes(useCaseRunnerType); err != nil {
		return fmt.Errorf("%w", err)
	}

	if !returnsSingleError(useCaseRunnerType) {
		return fmt.Errorf("%w: %s given", ErrInvalidReturnType, useCaseRunnerType.String())
	}

	return nil
}

func returnsSingleError(useCaseRunnerType reflect.Type) bool {
	return useCaseRunnerType.NumOut() == 1 && useCaseRunnerType.Out(0) == reflect.TypeFor[error]()
}

// newRunner validates the signature once and captures everything the call needs, so the
// returned runner only checks the argument types and calls the function.
func newRunner(fnValue reflect.Value) (UseCaseRunnerFn, error) {
	var fnType reflect.Type
	if fnValue.IsValid() {
		fnType = fnValue.Type()
	}

	if err := ensureSignatureIsValid(fnType); err != nil {
		return nil, err
	}

	ctxType, reqType, respType := fnType.In(0), fnType.In(1), fnType.In(2)

	return func(ctx context.Context, req Request, resp Response) error {
		if err := ensureArgHasType(reqType, req, ErrRequestTypeMismatch); err != nil {
			return err
		}

		if err := ensureArgHasType(respType, resp, ErrResultTypeMismatch); err != nil {
			return err
		}

		in := [3]reflect.Value{contextValue(ctx, ctxType), reflect.ValueOf(req), reflect.ValueOf(resp)}
		if err, ok := reflect.TypeAssert[error](fnValue.Call(in[:])[0]); ok && err != nil {
			return err
		}

		return nil
	}, nil
}

func ensureArgHasType(want reflect.Type, arg any, errMismatch error) error {
	if got := reflect.TypeOf(arg); got != want {
		return fmt.Errorf("%w: want %v, got %v", errMismatch, want, got)
	}

	return nil
}

// contextValue passes a nil context as the zero value of the parameter's type:
// reflect.ValueOf(nil) is the zero Value, which Call rejects.
func contextValue(ctx context.Context, ctxType reflect.Type) reflect.Value {
	if ctx == nil {
		return reflect.Zero(ctxType)
	}

	return reflect.ValueOf(ctx)
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
	ctxtInterface := reflect.TypeFor[context.Context]()
	ctx := useCaseRunnerType.In(0)

	return ctx.Implements(ctxtInterface)
}

func secondArgIsRequest(useCaseRunnerType reflect.Type) bool {
	requestInterface := reflect.TypeFor[Request]()

	secondArg := useCaseRunnerType.In(1)
	if secondArg.Kind() == reflect.Pointer {
		secondArg = secondArg.Elem()
	}

	return secondArg.Kind() == reflect.Struct && secondArg.Implements(requestInterface)
}

func thirdArgIsResponse(useCaseRunnerType reflect.Type) bool {
	responseInterface := reflect.TypeFor[Response]()
	thirdArg := useCaseRunnerType.In(2)

	return thirdArg.Kind() == reflect.Pointer && thirdArg.Implements(responseInterface)
}
