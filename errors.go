package interactor

import "errors"

// Errors returned by [Dispatcher.Register] and [Dispatcher.Dispatch]. They come back wrapped with
// the types involved, so match them with [errors.Is].
var (
	ErrUseCaseNotFound      = errors.New("interactor: no use case registered for request type")
	ErrAlreadyRegistered    = errors.New("interactor: use case already registered for request type")
	ErrNilUseCase           = errors.New("interactor: use case must not be nil")
	ErrResponseTypeMismatch = errors.New("interactor: registered use case returns a different response type")
)
