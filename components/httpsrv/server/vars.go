package server

import (
	"time"

	"github.com/xigxog/kubefox/api"
	kubefox "github.com/xigxog/kubefox/core"
)

var (
	HTTPAddr, HTTPSAddr       string
	BrokerAddr, HealthSrvAddr string
	EventTimeout              time.Duration
	MaxEventSize              int64
)

var (
	Component = new(kubefox.Component)
	Spec      = new(api.ComponentDefinition)
)
