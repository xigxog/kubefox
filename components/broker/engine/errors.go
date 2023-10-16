package engine

import "errors"

var (
	ErrComponentGone     = errors.New("component gone")
	ErrComponentMismatch = errors.New("component mismatch")
	ErrEventInvalid      = errors.New("event invalid")
	ErrEventTimeout      = errors.New("event time out")
	ErrRouteInvalid      = errors.New("route invalid")
	ErrRouteNotFound     = errors.New("route not found")
	ErrSubCanceled       = errors.New("subscription canceled")
	ErrUnexpected        = errors.New("unexpected error")
)
