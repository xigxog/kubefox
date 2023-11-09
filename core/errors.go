package core

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"runtime"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

var RecordStackTraces bool

type Code int

const (
	CodeUnexpected Code = iota

	CodeBrokerMismatch
	CodeBrokerUnavailable
	CodeComponentGone
	CodeComponentMismatch
	CodeContentTooLarge
	CodeInvalid
	CodeNotFound
	CodePortUnavailable
	CodeRouteInvalid
	CodeRouteNotFound
	CodeTimeout
	CodeUnauthorized
	CodeUnknownContentType
	CodeUnsupportedAdapter
)

type Err struct {
	// Keeps members private but allows for easy JSON marshalling.
	err err
}

type err struct {
	*Stack `json:"-"`

	Code     Code       `json:"code"`
	GRPCCode codes.Code `json:"grpcCode"`
	HTTPCode int        `json:"httpCode"`
	Msg      string     `json:"msg,omitempty"`
	Cause    string     `json:"cause,omitempty"`

	cause error
}

func ErrBrokerMismatch(cause ...error) *Err {
	return NewKubeFoxErr("broker mismatch", CodeBrokerMismatch, codes.FailedPrecondition, http.StatusBadGateway, cause...)
}

func ErrBrokerUnavailable(cause ...error) *Err {
	return NewKubeFoxErr("broker unavailable", CodeBrokerUnavailable, codes.Unavailable, http.StatusBadGateway, cause...)
}

func ErrComponentGone(cause ...error) *Err {
	return NewKubeFoxErr("component gone", CodeComponentGone, codes.FailedPrecondition, http.StatusBadGateway, cause...)
}

func ErrComponentMismatch(cause ...error) *Err {
	return NewKubeFoxErr("component mismatch", CodeComponentMismatch, codes.FailedPrecondition, http.StatusBadGateway, cause...)
}

func ErrContentTooLarge(cause ...error) *Err {
	return NewKubeFoxErr("content too large", CodeContentTooLarge, codes.ResourceExhausted, http.StatusRequestEntityTooLarge, cause...)
}

func ErrInvalid(cause ...error) *Err {
	return NewKubeFoxErr("invalid", CodeInvalid, codes.InvalidArgument, http.StatusBadRequest, cause...)
}

func ErrNotFound(cause ...error) *Err {
	return NewKubeFoxErr("not found", CodeNotFound, codes.Unimplemented, http.StatusNotFound, cause...)
}

func ErrPortUnavailable(cause ...error) *Err {
	return NewKubeFoxErr("port unavailable", CodePortUnavailable, codes.Unavailable, http.StatusConflict, cause...)
}

func ErrRouteInvalid(cause ...error) *Err {
	return NewKubeFoxErr("route invalid", CodeRouteInvalid, codes.InvalidArgument, http.StatusBadRequest, cause...)
}

func ErrRouteNotFound(cause ...error) *Err {
	return NewKubeFoxErr("route not found", CodeRouteNotFound, codes.Unimplemented, http.StatusNotFound, cause...)
}

func ErrTimeout(cause ...error) *Err {
	return NewKubeFoxErr("time out", CodeTimeout, codes.DeadlineExceeded, http.StatusGatewayTimeout, cause...)
}

func ErrUnauthorized(cause ...error) *Err {
	return NewKubeFoxErr("component unauthorized", CodeUnauthorized, codes.PermissionDenied, http.StatusForbidden, cause...)
}

func ErrUnexpected(cause ...error) *Err {
	return NewKubeFoxErr("unexpected error", CodeUnexpected, codes.Unknown, http.StatusInternalServerError, cause...)
}

func ErrUnknownContentType(cause ...error) *Err {
	return NewKubeFoxErr("unknown content type", CodeUnknownContentType, codes.InvalidArgument, http.StatusBadRequest, cause...)
}

func ErrUnsupportedAdapter(cause ...error) *Err {
	return NewKubeFoxErr("unsupported adapter", CodeUnsupportedAdapter, codes.Unimplemented, http.StatusBadRequest, cause...)
}

func NewKubeFoxErr(msg string, code Code, grpcCode codes.Code, httpCode int, cause ...error) *Err {
	var c error
	if len(cause) > 0 {
		c = cause[0]
	}

	var s *Stack
	if RecordStackTraces || code == CodeUnexpected {
		s = callers()
	}

	return &Err{
		err: err{
			Stack:    s,
			Msg:      msg,
			Code:     code,
			GRPCCode: grpcCode,
			HTTPCode: httpCode,
			cause:    c,
		},
	}
}

func (e *Err) Code() Code {
	return e.err.Code
}

func (e *Err) GRPCCode() codes.Code {
	return e.err.GRPCCode
}

func (e *Err) GRPCStatus() *status.Status {
	return status.New(e.err.GRPCCode, e.err.Msg)
}

func (e *Err) HTTPCode() int {
	return e.err.HTTPCode
}

func (e *Err) Unwrap() error {
	return e.err.cause
}

func (e *Err) Is(err error) bool {
	_, ok := err.(*Err)
	return ok
}

func (e *Err) Error() string {
	return e.String()
}

func (e *Err) String() string {
	if e.err.cause != nil {
		return e.err.Msg + ": " + e.err.cause.Error()
	}
	return e.err.Msg
}

func (e *Err) Format(s fmt.State, verb rune) {
	if e.err.Stack == nil {
		fmt.Fprint(s, e.String())
	} else {
		fmt.Fprintf(s, "%s%+v\n---", e.String(), e.err.StackTrace())
	}
}

// UnmarshalJSON implements the json.Unmarshaller interface.
func (e *Err) UnmarshalJSON(value []byte) error {
	e.err = err{}
	if err := json.Unmarshal(value, &e.err); err != nil {
		return err
	}
	if e.err.Cause != "" {
		e.err.cause = errors.New(e.err.Cause)
	}
	return nil
}

// MarshalJSON implements the json.Marshaller interface.
func (e *Err) MarshalJSON() ([]byte, error) {
	if e.err.cause != nil {
		e.err.Cause = e.err.cause.Error()
	}
	return json.Marshal(e.err)
}

func callers() *Stack {
	const depth = 32
	var pcs [depth]uintptr
	n := runtime.Callers(4, pcs[:])
	var st Stack = pcs[0:n]
	return &st
}
