package interactor

import "errors"

// Guard errors.
var (
	ErrUseCaseRunnerNotFound         = errors.New("use case runner not registered for the given request type")
	ErrInvalidUseCaseRunnerSignature = errors.New("useCaseRunner must have 3 input params")
	ErrUseCaseRunnerIsNotAFunction   = errors.New("useCaseRunner is not a function")
	ErrUseCaseRunnerHasNoRunMethod   = errors.New("useCaseRunner has no valid Run method")
	ErrFirstArgHasInvalidType        = errors.New("first input argument must have context.Context type")
	ErrSecondArgHasInvalidType       = errors.New("second input argument must implement Request interface")
	ErrThirdArgHasInvalidType        = errors.New("third input argument must implement Response interface")
	ErrResultTypeMismatch            = errors.New("result type mismatch")
)
