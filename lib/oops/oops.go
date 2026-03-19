// Package oops provides an unified error type for gRPC and HTTP transports.
// Internal errors are always sanitized before being sent to clients.
package oops

import (
	"errors"
	"fmt"
	"net/http"
	"path/filepath"
	"runtime"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// httpToGRPCCodes maps HTTP status codes to gRPC codes.
var httpToGRPCCodes = map[int]codes.Code{
	http.StatusBadRequest:          codes.InvalidArgument,
	http.StatusUnauthorized:        codes.Unauthenticated,
	http.StatusForbidden:           codes.PermissionDenied,
	http.StatusNotFound:            codes.NotFound,
	http.StatusConflict:            codes.AlreadyExists,
	http.StatusInternalServerError: codes.Internal,
	http.StatusNotImplemented:      codes.Unimplemented,
	http.StatusServiceUnavailable:  codes.Unavailable,
}

// grpcToHTTPCodes maps gRPC codes to HTTP status codes.
var grpcToHTTPCodes = map[codes.Code]int{
	codes.InvalidArgument:  http.StatusBadRequest,
	codes.Unauthenticated:  http.StatusUnauthorized,
	codes.PermissionDenied: http.StatusForbidden,
	codes.NotFound:         http.StatusNotFound,
	codes.AlreadyExists:    http.StatusConflict,
	codes.Internal:         http.StatusInternalServerError,
	codes.Unimplemented:    http.StatusNotImplemented,
	codes.Unavailable:      http.StatusServiceUnavailable,
}

// HTTPToGRPC returns the gRPC code for an HTTP status code.
// Returns codes.Unknown for unmapped codes.
func HTTPToGRPC(httpCode int) codes.Code {
	if c, ok := httpToGRPCCodes[httpCode]; ok {
		return c
	}
	return codes.Unknown
}

// GRPCToHTTP returns the HTTP status code for a gRPC code.
// Returns http.StatusInternalServerError for unmapped codes.
func GRPCToHTTP(code codes.Code) int {
	if h, ok := grpcToHTTPCodes[code]; ok {
		return h
	}
	return http.StatusInternalServerError
}

const internalClientMsg = "internal server error"

// callerLoc returns "file.go:line" for the frame skip+1 levels above callerLoc.
// skip=0 returns the caller of callerLoc, skip=1 returns its caller, and so on.
func callerLoc(skip int) string {
	_, file, line, ok := runtime.Caller(skip + 1)
	if !ok {
		return ""
	}
	return fmt.Sprintf("%s:%d", filepath.Base(file), line)
}

// Error is a domain error carrying a gRPC/HTTP status code.
// For internal errors, the raw cause is never exposed to clients.
type Error struct {
	code  codes.Code
	msg   string // client-facing message; unused for Internal errors
	cause error  // original cause for Internal errors (server-side only)
	loc   string // "file.go:line" of the call site that created this error
}

// Unwrap returns the underlying cause for Internal errors, enabling errors.Is
// and errors.As to traverse the error chain.
func (e *Error) Unwrap() error {
	return e.cause
}

// Error implements the error interface.
// Returns the client-facing message (or cause message for Internal errors)
// followed by the call site location.
func (e *Error) Error() string {
	msg := e.msg
	if e.cause != nil {
		msg = e.cause.Error()
	}
	if e.loc != "" {
		return e.loc + ": " + msg
	}
	return msg
}

// GRPCStatus implements the interface needed by the gRPC framework for building
// response status.
// Unlike WriteHTTP, the raw cause message is included for Internal errors.
func (e *Error) GRPCStatus() *status.Status {
	return status.New(e.code, e.Error())
}

// WriteHTTP writes the associated HTTP status code and message to w.
// Internal errors are sanitized: the cause is never included in the client message.
func (e *Error) WriteHTTP(w http.ResponseWriter) {
	msg := e.msg
	if e.code == codes.Internal {
		msg = internalClientMsg
	}
	http.Error(w, msg, GRPCToHTTP(e.code))
}

// Code returns the gRPC code associated with this error.
func (e *Error) Code() codes.Code {
	return e.code
}

// Is reports whether err is an *Error with the given gRPC code.
func Is(err error, code codes.Code) bool {
	var e *Error
	return errors.As(err, &e) && e.code == code
}

// NotFound creates an Error with a NotFound code.
func NotFound(format string, v ...any) error {
	msg := fmt.Sprintf(format, v...)
	return &Error{code: codes.NotFound, msg: msg, loc: callerLoc(1)}
}

// InvalidArgument creates an Error with an InvalidArgument code.
func InvalidArgument(format string, v ...any) error {
	msg := fmt.Sprintf(format, v...)
	return &Error{code: codes.InvalidArgument, msg: msg, loc: callerLoc(1)}
}

// AlreadyExists creates an Error with an AlreadyExists code.
func AlreadyExists(format string, v ...any) error {
	msg := fmt.Sprintf(format, v...)
	return &Error{code: codes.AlreadyExists, msg: msg, loc: callerLoc(1)}
}

// Internal wraps an unexpected error. The cause is available via Error() for
// server-side logging but is never sent to clients.
func Internal(err error) error {
	if err == nil {
		return nil
	}
	return &Error{code: codes.Internal, cause: err, loc: callerLoc(1)}
}

// Unauthenticated creates an Error with an Unauthenticated code.
func Unauthenticated(format string, v ...any) error {
	msg := fmt.Sprintf(format, v...)
	return &Error{code: codes.Unauthenticated, msg: msg, loc: callerLoc(1)}
}

// PermissionDenied creates an Error with a PermissionDenied code.
func PermissionDenied(format string, v ...any) error {
	msg := fmt.Sprintf(format, v...)
	return &Error{code: codes.PermissionDenied, msg: msg, loc: callerLoc(1)}
}

// Unimplemented creates an Error with an Unimplemented code.
func Unimplemented(format string, v ...any) error {
	msg := fmt.Sprintf(format, v...)
	return &Error{code: codes.Unimplemented, msg: msg, loc: callerLoc(1)}
}

// Unavailable creates an Error with an Unavailable code.
func Unavailable(format string, v ...any) error {
	msg := fmt.Sprintf(format, v...)
	return &Error{code: codes.Unavailable, msg: msg, loc: callerLoc(1)}
}
