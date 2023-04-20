package config

import (
	"github.com/xigxog/kubefox/libs/core/component"
	"github.com/xigxog/kubefox/libs/core/grpc"
)

type Config struct {
	Flags

	Comp component.Component
	Dev  DevContext
}

type Flags struct {
	Addrs

	// Organization      string
	Platform    string
	System      string
	CompName    string
	CompGitHash string
	Namespace   string

	EventTimeout int64

	IsRuntimeSrv  bool
	SkipBootstrap bool

	DevEnv    string
	DevApp    string
	IsDevMode bool

	LogLevel string
}

type Addrs struct {
	RuntimeSrvAddr   string
	GRPCSrvAddr      string
	DevHTTPSrvAddr   string
	HTTPSrvAddr      string
	NatsAddr         string
	TelemetrySrvAddr string
	TraceAgentAddr   string
}

type DevContext struct {
	*grpc.EventContext

	Target component.Component
}
