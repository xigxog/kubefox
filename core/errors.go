package core

import (
	"encoding/json"
	"fmt"
	"net/http"
	"runtime"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type Err struct {
	// Keeps members private but allows for easy JSON marshalling.
	err
}

type err struct {
	*Stack `json:"-"`

	GRPCCode codes.Code `json:"grpcCode,omitempty"`
	HTTPCode int        `json:"httpCode,omitempty"`
	Msg      string     `json:"msg,omitempty"`
	cause    error
}

var (
	ErrBrokerMismatch        = NewErr("broker mismatch", codes.FailedPrecondition, http.StatusBadGateway)
	ErrBrokerUnavailable     = NewErr("broker unavailable", codes.Unavailable, http.StatusBadGateway)
	ErrComponentGone         = NewErr("component gone", codes.FailedPrecondition, http.StatusBadGateway)
	ErrComponentInvalid      = NewErr("component invalid", codes.InvalidArgument, http.StatusBadRequest)
	ErrComponentMismatch     = NewErr("component mismatch", codes.FailedPrecondition, http.StatusBadGateway)
	ErrComponentUnauthorized = NewErr("component unauthorized", codes.PermissionDenied, http.StatusForbidden)
	ErrComponentUnknown      = NewErr("component unknown", codes.Unauthenticated, http.StatusUnauthorized)
	ErrInvalid               = NewErr("invalid", codes.InvalidArgument, http.StatusBadRequest)
	ErrNotFound              = NewErr("not found", codes.Unimplemented, http.StatusNotFound)
	ErrRequestGone           = NewErr("request gone", codes.DeadlineExceeded, http.StatusGatewayTimeout)
	ErrResponseInvalid       = NewErr("response invalid", codes.FailedPrecondition, http.StatusBadGateway)
	ErrRouteInvalid          = NewErr("route invalid", codes.InvalidArgument, http.StatusBadRequest)
	ErrTimeout               = NewErr("time out", codes.DeadlineExceeded, http.StatusGatewayTimeout)
	ErrPortUnavailable       = NewErr("port unavailable", codes.Unavailable, http.StatusConflict)
	ErrRouteNotFound         = NewErr("route not found", codes.Unimplemented, http.StatusNotFound)
	ErrUnexpected            = NewErr("unexpected error", codes.Unknown, http.StatusInternalServerError)
	ErrUnsupportedAdapter    = NewErr("unsupported adapter", codes.Unimplemented, http.StatusBadRequest)
	ErrUnknownContentType    = NewErr("unknown content type", codes.InvalidArgument, http.StatusBadRequest)
)

func NewErr(msg string, grpcCode codes.Code, httpCode int) *Err {
	return &Err{err: err{
		Msg:      msg,
		GRPCCode: grpcCode,
		HTTPCode: httpCode,
	}}
}

func (e *Err) GRPCStatus() *status.Status {
	return status.New(e.err.GRPCCode, e.Msg)
}

func (e *Err) GRPCCode() codes.Code {
	return e.err.GRPCCode
}

func (e *Err) HTTPCode() int {
	return e.err.HTTPCode
}

// Details creates a copy of err and updates the copy's msg by appending
// the provided details. A stack trace is then recorded.
func (e *Err) Details(details string) *Err {
	copy := *e
	if details != "" {
		copy.Msg = fmt.Sprintf("%s: %s", copy.Msg, details)
	}
	copy.Stack = callers()
	return &copy
}

// Wrap creates a copy of err, sets the copy's cause to the provided error and
// updates its msg by appending details and the cause msg if they are not empty.
// A stack trace is then recorded.
func (e *Err) Wrap(cause error, details string) *Err {
	copy := *e
	copy.cause = cause
	if details != "" {
		copy.Msg = fmt.Sprintf("%s: %s", copy.Msg, details)
	}
	if cause != nil && cause.Error() != "" {
		copy.Msg = fmt.Sprintf("%s: %s", copy.Msg, cause.Error())
	}
	copy.Stack = callers()
	return &copy
}

// RecordStack creates a copy of err then records a stack trace and attaches it
// to the err.
func (e *Err) RecordStack() *Err {
	copy := *e
	copy.Stack = callers()
	return &copy
}

func (e *Err) Cause() error {
	return e.cause
}

func (e *Err) Error() string {
	return e.String()
}

func (e *Err) String() string {
	return e.Msg
}

func (e *Err) Format(s fmt.State, verb rune) {
	if e.Stack == nil {
		fmt.Fprint(s, e.String())
	} else {
		fmt.Fprintf(s, "%s%+v\n---", e.String(), e.StackTrace())
	}
}

// UnmarshalJSON implements the json.Unmarshaller interface.
func (e *Err) UnmarshalJSON(value []byte) error {
	e.err = err{}
	return json.Unmarshal(value, &e.err)
}

// MarshalJSON implements the json.Marshaller interface.
func (err *Err) MarshalJSON() ([]byte, error) {
	return json.Marshal(err.err)
}

func callers() *Stack {
	const depth = 32
	var pcs [depth]uintptr
	n := runtime.Callers(3, pcs[:])
	var st Stack = pcs[0:n]
	return &st
}
