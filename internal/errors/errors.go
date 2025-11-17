// Package errors provides custom error types and handling utilities
package errors

import (
	"errors"
	"fmt"
)

// ErrorType identifies the category of error
type ErrorType int

const (
	// Error type constants
	Unknown ErrorType = iota
	Configuration
	Network
	Browser
	UserInput
	External
)

// AppError represents an application-specific error with context
type AppError struct {
	Type    ErrorType
	Message string
	Err     error
}

// Error implements the error interface
func (e *AppError) Error() string {
	if e.Err != nil {
		return fmt.Sprintf("%s: %v", e.Message, e.Err)
	}
	return e.Message
}

// Unwrap returns the wrapped error
func (e *AppError) Unwrap() error {
	return e.Err
}

// Is implements error comparison for errors.Is
func (e *AppError) Is(target error) bool {
	t, ok := target.(*AppError)
	if !ok {
		return false
	}
	return e.Type == t.Type
}

// NewError creates a new application error
func NewError(errType ErrorType, message string, err error) error {
	return &AppError{
		Type:    errType,
		Message: message,
		Err:     err,
	}
}

// Configuration errors
func NewConfigError(message string, err error) error {
	return NewError(Configuration, message, err)
}

// Network errors
func NewNetworkError(message string, err error) error {
	return NewError(Network, message, err)
}

// Browser errors
func NewBrowserError(message string, err error) error {
	return NewError(Browser, message, err)
}

// User input errors
func NewUserInputError(message string, err error) error {
	return NewError(UserInput, message, err)
}

// External system errors
func NewExternalError(message string, err error) error {
	return NewError(External, message, err)
}

// IsErrorType checks if an error is of a specific type
func IsErrorType(err error, errorType ErrorType) bool {
	var appErr *AppError
	if err == nil {
		return false
	}
	return errors.As(err, &appErr) && appErr.Type == errorType
}