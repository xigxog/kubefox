package server

import (
	"time"

	"github.com/xigxog/kubefox/api"
	"github.com/xigxog/kubefox/core"
)

var (
	Platform                  string
	HTTPAddr, HTTPSAddr       string
	BrokerAddr, HealthSrvAddr string
	EventTimeout              time.Duration
	MaxEventSize              int64
)

var (
	Component    = &core.Component{Id: core.GenerateId(), Type: string(api.ComponentTypeHTTPAdapter)}
	ComponentDef = &api.ComponentDefinition{Type: api.ComponentTypeHTTPAdapter}
)
