package kubefox

type ContextKey string

const (
	LogContextKey ContextKey = "LogContext"
)

// OS status codes
const (
	ConfigurationErrorCode   = 10
	JetStreamErrorCode       = 11
	RpcServerErrorCode       = 12
	HTTPServerErrorCode      = 13
	TelemetryServerErrorCode = 14
	InterruptCode            = 130
)

// Component event types
const (
	ComponentRequestType   = "io.kubefox.component_request"
	ComponentResponseType  = "io.kubefox.component_response"
	CronRequestType        = "io.kubefox.cron_request"
	CronResponseType       = "io.kubefox.cron_response"
	DaprRequestType        = "io.kubefox.dapr_request"
	DaprResponseType       = "io.kubefox.dapr_response"
	HTTPRequestType        = "io.kubefox.http_request"
	HTTPResponseType       = "io.kubefox.http_response"
	KubernetesRequestType  = "io.kubefox.kubernetes_request"
	KubernetesResponseType = "io.kubefox.kubernetes_response"
)

// Platform event types
//
// Update `IsPlatformEvent()` in event.go if changes are made.
const (
	BootstrapRequestType  = "io.kubefox.bootstrap_request"
	BootstrapResponseType = "io.kubefox.bootstrap_response"
	FabricRequestType     = "io.kubefox.fabric_request"
	FabricResponseType    = "io.kubefox.fabric_response"
)

// Internal event types
const (
	HealthRequestType     = "io.kubefox.health_request"
	HealthResponseType    = "io.kubefox.health_response"
	MetricsRequestType    = "io.kubefox.metrics_request"
	MetricsResponseType   = "io.kubefox.metrics_response"
	TelemetryRequestType  = "io.kubefox.telemetry_request"
	TelemetryResponseType = "io.kubefox.telemetry_response"

	ErrorEventType    = "io.kubefox.error"
	RejectedEventType = "io.kubefox.rejected"
	UnknownEventType  = "io.kubefox.unknown"
)

// KubeFox headers/query params
const (
	EnvHeader       = "kf-environment"
	EnvHeaderShort  = "kf-env"
	SysHeader       = "kf-system"
	SysHeaderShort  = "kf-sys"
	TargetHeader    = "kf-target"
	EventTypeHeader = "kf-event-type"

	RelEnvHeader       = "kf-release-environment"
	RelSysHeader       = "kf-release-system"
	RelTargetHeader    = "kf-release-target"
	RelEventTypeHeader = "kf-release-event-type"
)
