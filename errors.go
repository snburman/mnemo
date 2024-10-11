package mnemo

import (
	"fmt"
)

type (
	// Error is a generic error type for the Mnemo package.
	Error[T any] struct {
		Err    error
		Status int
		Logger Logger
		level  LogLevel
	}
)

// NewError returns a new Error instance with a logger
func NewError[T any](msg string, opts ...Opt[Error[T]]) Error[T] {
	e := Error[T]{
		Err:    fmt.Errorf(msg),
		Logger: logger,
	}
	return e
}

func (e Error[T]) Type() string {
	return fmt.Sprintf("%T", e)
}

// Error implements the error interface.
func (e Error[T]) Error() string {
	return fmt.Sprintf("%v error: %v", new(T), e.Err.Error())
}

// IsStatusError returns true if the error has a status code.
func (e Error[T]) IsStatusError() bool {
	return e.Status != 0
}

// Log logs the error by level.
func (e Error[T]) Log() {
	err := fmt.Sprintf("%T error: %v", e, e.Err.Error())
	switch e.level {
	case Err:
		e.Logger.Error(err)
	case Debug:
		e.Logger.Debug(err)
	case Info:
		e.Logger.Info(err)
	case Warn:
		e.Logger.Warn(err)
	case Fatal:
		e.Logger.Fatal(err)
	case Panic:
		panic(err)
	default:
		e.Logger.Error(err)
	}
}

// WithStatus sets the status code for the error.
func (e Error[T]) WithStatus(status int) Error[T] {
	e.Status = status
	return e
}

// WithLogLevel sets the log level for the error.
func (e Error[T]) WithLogLevel(level LogLevel) Error[T] {
	e.level = level
	return e
}

// IsErrorType reflects Error[T] from an error.
func IsErrorType[T any](err error) (Error[T], bool) {
	t, ok := err.(Error[T])
	return t, ok
}
