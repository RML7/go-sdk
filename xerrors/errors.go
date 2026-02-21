package xerrors

import (
	stderrors "errors"

	"github.com/pkg/errors"
)

var (
	// ErrNotFound is the error returned when a resource is not found.
	ErrNotFound = New("not found")
	// ErrAlreadyExists is the error returned when a resource is already exists.
	ErrAlreadyExists = New("already exists")
	// ErrInvalidArgument is the error returned when a request is invalid.
	ErrInvalidArgument = New("invalid argument")
	// ErrFailedPrecondition is the error returned when a request is failed precondition.
	ErrFailedPrecondition = New("failed precondition")
	// ErrPermissionDenied is the error returned when a request is denied.
	ErrPermissionDenied = New("permission denied")
	// ErrUnauthenticated is the error returned when a request is not authenticated.
	ErrUnauthenticated = New("unauthenticated")
	// ErrResourceExhausted is the error returned when a request cannot be completed due to resource exhaustion.
	ErrResourceExhausted = New("resource exhausted")
	// ErrCancelled is the error returned when a request was canceled by client.
	ErrCancelled = New("canceled")
	// ErrAborted is the error returned when a request was aborted by the server.
	ErrAborted = New("aborted")
	// ErrDeadlineExceeded is the error returned when a request was canceled due to timeout.
	ErrDeadlineExceeded = New("deadline exceeded")
	// ErrUnknown is being used when an error is lost, nevertheless there must be an error.
	// Must be treated as internal server error.
	ErrUnknown = New("unknown")
)

func New(message string) error {
	return errors.New(message)
}

func Errorf(format string, args ...interface{}) error {
	return errors.Errorf(format, args...)
}

func WithMessage(err error, message string) error {
	if err == nil {
		err = ErrUnknown
	}

	err = errors.WithMessage(err, message)

	return err
}

func WithMessagef(err error, format string, args ...interface{}) error {
	if err == nil {
		err = ErrUnknown
	}

	err = errors.WithMessagef(err, format, args...)

	return err
}

func Is(err, target error) bool {
	return errors.Is(err, target)
}

func As(err error, target interface{}) bool {
	return errors.As(err, target)
}

func Join(errs ...error) error {
	return stderrors.Join(errs...)
}
