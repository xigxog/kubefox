package server

import (
	"time"

	kubefox "github.com/xigxog/kubefox/core"
)

var (
	HTTPAddr, HTTPSAddr       string
	BrokerAddr, HealthSrvAddr string
	EventTTL                  time.Duration
)

var (
	Component = new(kubefox.Component)
	Spec      = new(kubefox.ComponentSpec)
)
