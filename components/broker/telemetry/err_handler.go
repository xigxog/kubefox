package telemetry

import "github.com/xigxog/kubefox/libs/core/logger"

type OTELErrorHandler struct {
	*logger.Log
}

func (h OTELErrorHandler) Handle(err error) {
	h.Log.Warn(err)
}
