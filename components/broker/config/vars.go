package config

import "time"

var (
	Instance  string
	Platform  string
	Namespace string

	NumWorkers        int
	TelemetryInterval time.Duration

	LogFormat string
	LogLevel  string

	GRPCSrvAddr   string
	HealthSrvAddr string

	NATSAddr      string
	TelemetryAddr string
)
