package kubefox

import (
	"fmt"

	"github.com/xigxog/kubefox/libs/core/grpc"
)

type ComponentSvc struct {
	comp *grpc.Component
	kit  *kit
}

type kitContext struct {
	kit    *kit
	target *grpc.Component
}

func (svc *ComponentSvc) Invoke(req Event) (Event, error) {
	dataReq, ok := req.(DataEvent)
	if !ok {
		return nil, fmt.Errorf("event does not contain data")
	}

	dataReq.SetTarget(svc.comp)

	resp, err := svc.kit.broker.InvokeTarget(dataReq)
	if err != nil {
		return nil, err
	}

	return resp, nil
}
