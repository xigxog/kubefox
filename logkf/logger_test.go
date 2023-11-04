package logkf

import (
	"encoding/json"
	"errors"
	"testing"

	"github.com/xigxog/kubefox/core"
)

func TestLogger_Error(t *testing.T) {
	err1 := core.ErrBrokerMismatch
	err2 := core.ErrBrokerMismatch.RecordStack()
	err3 := core.ErrBrokerMismatch.Details("these are details")
	err4 := core.ErrBrokerMismatch.Wrap(errors.New("wrapped err"), "details")

	l := Global
	l.Debug(err1)
	l.Debug(err2)
	l.Debug(err3)
	l.Debug(err4)
	l.Error(err1)
	l.Error(err2)
	l.Error(err3)
	l.Error(err4)

	b, _ := json.Marshal(err1)
	l.Debug(string(b))
	b, _ = json.Marshal(err2)
	l.Debug(string(b))
	b, _ = json.Marshal(err3)
	l.Debug(string(b))
	b, _ = json.Marshal(err4)
	l.Debug(string(b))

	e := &core.Err{}
	if err := json.Unmarshal(b, e); err != nil {
		l.Debug(err)
	} else {
		l.Debugf("%s, grpcCode: %d, httpCode: %d", e, e.GRPCCode, e.HTTPCode)
	}
}
