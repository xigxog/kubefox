package engine

import "errors"

var (
	ErrEventInvalid  = errors.New("event invalid")
	ErrEventTimeout  = errors.New("event timed out")
	ErrRouteInvalid  = errors.New("route invalid")
	ErrRouteNotFound = errors.New("route to component not found")
	ErrSubCanceled   = errors.New("subscription canceled")
	ErrUnexpected    = errors.New("unexpected error")
	ErrComponentGone = errors.New("component gone")
)
