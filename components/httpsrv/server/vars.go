package server

import (
	"time"

	common "github.com/xigxog/kubefox/api/kubernetes"
	kubefox "github.com/xigxog/kubefox/core"
)

var (
	HTTPAddr, HTTPSAddr       string
	BrokerAddr, HealthSrvAddr string
	EventTimeout              time.Duration
)

var (
	Component = new(kubefox.Component)
	Spec      = new(common.ComponentDefinition)
)
