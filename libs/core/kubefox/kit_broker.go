package kubefox

import (
	"fmt"

	"github.com/xigxog/kubefox/libs/core/grpc"
)

type KitBroker interface {
	InvokeTarget(DataEvent) (DataEvent, error)
}

type kitBroker struct {
	kit    *kit
	broker grpc.ComponentServiceClient
}

func (brk *kitBroker) InvokeTarget(req DataEvent) (DataEvent, error) {
	if req.GetTarget() == nil {
		return nil, fmt.Errorf("target missing")
	}

	req.SetParent(brk.kit.req)
	req.GetTarget().SetApp(brk.kit.req.GetSource().GetApp())

	respData, err := brk.broker.InvokeTarget(brk.kit.Ctx(), req.GetData())
	if err != nil {
		return nil, err
	}

	return newEvent(respData), nil
}

func (brk *kitBroker) SendResponse(resp DataEvent) error {
	return nil
}
