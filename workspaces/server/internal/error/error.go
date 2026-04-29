package apperror

import (
	"errors"
	"fmt"

	"github.com/NishLy/go-fiber-boilerplate/internal/response"
)

// Code represents an application-level error code.
type Code int

const (
	NotFound         Code = 404
	Duplicate        Code = 409
	Internal         Code = 500
	Invalid          Code = 400
	PermissionDenied Code = 403
	Unauthorized     Code = 401
)

// String makes Code implement fmt.Stringer for readable logging.
func (c Code) String() string {
	switch c {
	case NotFound:
		return "NOT_FOUND"
	case Duplicate:
		return "DUPLICATE"
	case Internal:
		return "INTERNAL"
	case Invalid:
		return "INVALID"
	case PermissionDenied:
		return "PERMISSION_DENIED"
	case Unauthorized:
		return "UNAUTHORIZED"
	default:
		return fmt.Sprintf("UNKNOWN(%d)", int(c))
	}
}

// Error is a structured application error carrying a code, human-readable
// message, and an optional underlying cause.
type Error struct {
	Code    Code
	Message string
	Err     error
	Data    *interface{} // Optional field for additional error context
}

func (e *Error) Error() string {
	msg := e.Message
	if e.Err != nil {
		msg = e.Err.Error()
	}
	return fmt.Sprintf("[%s] %s", e.Code, msg)
}

func (e *Error) Unwrap() error { return e.Err }

// Is allows errors.Is(err, &Error{Code: NotFound}) matching by code.
func (e *Error) Is(target error) bool {
	t, ok := target.(*Error)
	if !ok {
		return false
	}
	return e.Code == t.Code
}

// New constructs an Error. Pass nil for err if there is no underlying cause.
func New(code Code, msg string, err error, data *interface{}) *Error {
	return &Error{Code: code, Message: msg, Err: err, Data: data}
}

// Sentinel errors for use with errors.Is.
var (
	ErrNotFound         = &Error{Code: NotFound}
	ErrDuplicate        = &Error{Code: Duplicate}
	ErrInternal         = &Error{Code: Internal}
	ErrInvalid          = &Error{Code: Invalid}
	ErrPermissionDenied = &Error{Code: PermissionDenied}
	ErrUnauthorized     = &Error{Code: Unauthorized}
)

// Convenience constructors — accept an optional custom message.
// If msg is empty, a default is used.

func NotFoundErr(err error, msg ...string) *Error {
	return New(NotFound, firstOr(msg, "resource not found"), err, nil)
}

func DuplicateErr(err error, msg ...string) *Error {
	return New(Duplicate, firstOr(msg, "duplicate data"), err, nil)
}

func InternalErr(err error, msg ...string) *Error {
	return New(Internal, firstOr(msg, "internal server error"), err, nil)
}

func BadRequestErr(err error, msg ...string) *Error {
	return New(Invalid, firstOr(msg, "bad request"), err, nil)
}

func ValidationErr(response []response.ValidationError, msg ...string) *Error {
	var data interface{} = response
	return New(Invalid, firstOr(msg, "validation failed"), fmt.Errorf("%v", response), &data)
}

func PermissionDeniedErr(err error, msg ...string) *Error {
	return New(PermissionDenied, firstOr(msg, "permission denied"), err, nil)
}

func UnauthorizedErr(err error, msg ...string) *Error {
	return New(Unauthorized, firstOr(msg, "unauthorized"), err, nil)
}

// IsCode reports whether any error in err's chain has the given code.
func IsCode(err error, code Code) bool {
	var e *Error
	if errors.As(err, &e) {
		return e.Code == code
	}
	return false
}

func firstOr(vals []string, fallback string) string {
	if len(vals) > 0 && vals[0] != "" {
		return vals[0]
	}
	return fallback
}
