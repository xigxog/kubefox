package kubefox

import "errors"

var (
	ErrChannelNotFound       = errors.New("a channel with that id was not found")
	ErrComponentNameNotFound = errors.New("component name not found, environment variable `UNIT_NAME` should be set")
	ErrDuplicateChannel      = errors.New("a channel with that id already exists")
	ErrEntrypointNotFound    = errors.New("a matching entrypoint was not found")
	ErrInvalidEnvironment    = errors.New("invalid environment")
	ErrInvalidSystem         = errors.New("invalid environment")
	ErrInvalidTarget         = errors.New("provided KubeFox target component is invalid")
	ErrMissingEnvironment    = errors.New("request is missing KubeFox Environment")
	ErrMissingId             = errors.New("id is missing")
	ErrMissingMetadata       = errors.New("request is missing metadata")
	ErrMissingRequestId      = errors.New("request id is missing")
	ErrMissingRequestor      = errors.New("requestor is missing")
	ErrMissingResponder      = errors.New("responder is missing")
	ErrMissingSystem         = errors.New("request is missing KubeFox System")
	ErrMissingSysRef         = errors.New("request is missing KubeFox SysRef")
	ErrMissingTarget         = errors.New("request is missing KubeFox target Component")
	ErrNotFound              = errors.New("not found")
	ErrSubscriptionClosed    = errors.New("subscription closed, unable to fetch")
	ErrSubscriptionExists    = errors.New("a component as already subscribed")
	ErrSubscriptionNotFound  = errors.New("a subscription with that id was not found")
	ErrTypeMismatch          = errors.New("event type mismatch")
	ErrUnsupportedRequest    = errors.New("request type is not supported")
)
