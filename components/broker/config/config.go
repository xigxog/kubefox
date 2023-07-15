package config

import (
	"github.com/xigxog/kubefox/libs/core/component"
	"github.com/xigxog/kubefox/libs/core/grpc"
)

// Injected at build time
var (
	GitRef  string
	GitHash string
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
	Namespace   string
	System      string
	CompName    string
	CompGitHash string

	CACertPath       string
	GRPCCertsDir     string
	OperatorCertsDir string

	EventTimeout    uint8
	MetricsInterval uint8

	IsOperator    bool
	SkipBootstrap bool

	DevEnv    string
	DevApp    string
	IsDevMode bool

	LogLevel string
}

type Addrs struct {
	OperatorAddr       string
	GRPCSrvAddr        string
	DevHTTPSrvAddr     string
	HTTPSrvAddr        string
	NATSAddr           string
	TelemetryAgentAddr string
	HealthSrvAddr      string
}

type DevContext struct {
	*grpc.EventContext

	Target component.Component
}
