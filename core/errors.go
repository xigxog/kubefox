package core

import (
	"errors"
	"fmt"
)

var (
	ErrKubeFox = errors.New("")

	ErrBrokerMismatch    = fmt.Errorf("%wbroker mismatch", ErrKubeFox)
	ErrComponentGone     = fmt.Errorf("%wcomponent gone", ErrKubeFox)
	ErrComponentMismatch = fmt.Errorf("%wcomponent mismatch", ErrKubeFox)
	ErrEventInvalid      = fmt.Errorf("%wevent invalid", ErrKubeFox)
	ErrEventRequestGone  = fmt.Errorf("%wevent request gone", ErrKubeFox)
	ErrEventTimeout      = fmt.Errorf("%wevent time out", ErrKubeFox)
	ErrRouteInvalid      = fmt.Errorf("%wroute invalid", ErrKubeFox)
	ErrRouteNotFound     = fmt.Errorf("%wroute not found", ErrKubeFox)
	ErrSubCanceled       = fmt.Errorf("%wsubscription canceled", ErrKubeFox)
	ErrUnexpected        = fmt.Errorf("%wunexpected error", ErrKubeFox)
)
