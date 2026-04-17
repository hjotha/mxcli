// SPDX-License-Identifier: Apache-2.0

// Package mdlerrors provides structured error types for the MDL executor.
//
// Typed errors support errors.As for programmatic classification.
// Sentinel or wrapped errors may also support errors.Is where applicable
// (for example, ErrExit and BackendError via Unwrap).
// Every error preserves the original message via Error() for backward-compatible
// string output — callers that only use %v or .Error() see no change.
//
// Only BackendError supports Unwrap — it wraps an underlying storage/IO error.
// All other error types are leaf errors with no wrapped cause.
package mdlerrors

import (
	"errors"
	"fmt"
)

// ErrExit is a sentinel error indicating clean script/session termination.
// Use errors.Is(err, ErrExit) to detect exit requests.
var ErrExit = errors.New("exit")

// NotConnectedError indicates an operation was attempted without an active project connection.
type NotConnectedError struct {
	// WriteMode is true when write access was required but not available.
	WriteMode bool
	msg       string
}

// NewNotConnected creates a NotConnectedError for read access.
func NewNotConnected() *NotConnectedError {
	return &NotConnectedError{msg: "not connected to a project"}
}

// NewNotConnectedMsg creates a NotConnectedError with a custom message.
func NewNotConnectedMsg(msg string) *NotConnectedError {
	return &NotConnectedError{msg: msg}
}

// NewNotConnectedWrite creates a NotConnectedError for write access.
func NewNotConnectedWrite() *NotConnectedError {
	return &NotConnectedError{WriteMode: true, msg: "not connected to a project in write mode"}
}

func (e *NotConnectedError) Error() string { return e.msg }

// NotFoundError indicates a named element was not found.
type NotFoundError struct {
	// Kind is the element type (e.g. "entity", "module", "microflow").
	Kind string
	// Name is the qualified or simple name of the element.
	Name string
	msg  string
}

// NewNotFound creates a NotFoundError.
func NewNotFound(kind, name string) *NotFoundError {
	return &NotFoundError{
		Kind: kind,
		Name: name,
		msg:  fmt.Sprintf("%s not found: %s", kind, name),
	}
}

// NewNotFoundMsg creates a NotFoundError with a custom message.
func NewNotFoundMsg(kind, name, msg string) *NotFoundError {
	return &NotFoundError{Kind: kind, Name: name, msg: msg}
}

func (e *NotFoundError) Error() string { return e.msg }

// AlreadyExistsError indicates an element already exists when creating.
type AlreadyExistsError struct {
	// Kind is the element type.
	Kind string
	// Name is the qualified or simple name.
	Name string
	msg  string
}

// NewAlreadyExists creates an AlreadyExistsError.
func NewAlreadyExists(kind, name string) *AlreadyExistsError {
	return &AlreadyExistsError{
		Kind: kind,
		Name: name,
		msg:  fmt.Sprintf("%s already exists: %s", kind, name),
	}
}

// NewAlreadyExistsMsg creates an AlreadyExistsError with a custom message.
func NewAlreadyExistsMsg(kind, name, msg string) *AlreadyExistsError {
	return &AlreadyExistsError{Kind: kind, Name: name, msg: msg}
}

func (e *AlreadyExistsError) Error() string { return e.msg }

// UnsupportedError indicates an unsupported operation, feature, or property.
type UnsupportedError struct {
	// What holds the full error message describing what is unsupported
	// (e.g. "unsupported attribute type: Binary").
	What string
	msg  string
}

// NewUnsupported creates an UnsupportedError.
func NewUnsupported(msg string) *UnsupportedError {
	return &UnsupportedError{What: msg, msg: msg}
}

func (e *UnsupportedError) Error() string { return e.msg }

// ValidationError indicates invalid input or configuration.
type ValidationError struct {
	msg string
}

// NewValidation creates a ValidationError.
func NewValidation(msg string) *ValidationError {
	return &ValidationError{msg: msg}
}

// NewValidationf creates a ValidationError with formatted message.
func NewValidationf(format string, args ...any) *ValidationError {
	return &ValidationError{msg: fmt.Sprintf(format, args...)}
}

func (e *ValidationError) Error() string { return e.msg }

// BackendError wraps an error from the underlying storage layer (mpr/SDK).
type BackendError struct {
	// Op describes the operation that failed (e.g. "get domain model", "write entity").
	Op  string
	Err error
}

// NewBackend creates a BackendError wrapping a cause.
func NewBackend(op string, err error) *BackendError {
	return &BackendError{Op: op, Err: err}
}

func (e *BackendError) Error() string {
	if e.Err == nil {
		return fmt.Sprintf("failed to %s", e.Op)
	}
	return fmt.Sprintf("failed to %s: %v", e.Op, e.Err)
}

func (e *BackendError) Unwrap() error { return e.Err }
