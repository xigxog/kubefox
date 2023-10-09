package kubefox

import (
	"path"
	"regexp"
)

type EventType string

// Component event types
const (
	ComponentRequestType   EventType = "io.kubefox.component_request"
	ComponentResponseType  EventType = "io.kubefox.component_response"
	CronRequestType        EventType = "io.kubefox.cron_request"
	CronResponseType       EventType = "io.kubefox.cron_response"
	DaprRequestType        EventType = "io.kubefox.dapr_request"
	DaprResponseType       EventType = "io.kubefox.dapr_response"
	HTTPRequestType        EventType = "io.kubefox.http_request"
	HTTPResponseType       EventType = "io.kubefox.http_response"
	KubernetesRequestType  EventType = "io.kubefox.kubernetes_request"
	KubernetesResponseType EventType = "io.kubefox.kubernetes_response"
)

// Platform event types
const (
	BootstrapRequestType  EventType = "io.kubefox.bootstrap_request"
	BootstrapResponseType EventType = "io.kubefox.bootstrap_response"
	HealthRequestType     EventType = "io.kubefox.health_request"
	HealthResponseType    EventType = "io.kubefox.health_response"
	MetricsRequestType    EventType = "io.kubefox.metrics_request"
	MetricsResponseType   EventType = "io.kubefox.metrics_response"
	TelemetryRequestType  EventType = "io.kubefox.telemetry_request"
	TelemetryResponseType EventType = "io.kubefox.telemetry_response"

	AckEventType      EventType = "io.kubefox.ack"
	ErrorEventType    EventType = "io.kubefox.error"
	NackEventType     EventType = "io.kubefox.nack"
	RegisterType      EventType = "io.kubefox.register"
	RejectedEventType EventType = "io.kubefox.rejected"
	UnknownEventType  EventType = "io.kubefox.unknown"
)

// Keys for well known values.
const (
	HeaderValKey     = "header"
	HostValKey       = "host"
	MethodValKey     = "method"
	PathValKey       = "path"
	QueryValKey      = "queryParam"
	StatusCodeValKey = "statusCode"
	StatusValKey     = "status"
	URLValKey        = "url"
	// ValKey = ""
)

// Headers and query params.
const (
	EnvHeader            = "kf-environment"
	EnvHeaderShort       = "kf-env"
	EnvHeaderAbbrv       = "kfe"
	DepHeader            = "kf-deployment"
	DepHeaderShort       = "kf-dep"
	DepHeaderAbbrv       = "kfd"
	EventTypeHeader      = "kf-type"
	EventTypeHeaderAbbrv = "kft"
)

var (
	CommitRegexp = regexp.MustCompile(`^[0-9a-f]{40}$`)
	ImageRegexp  = regexp.MustCompile(`^.*:[a-z0-9-]{7}$`)
	// TODO use SHA256, switch pattern to ^.*@sha256:[a-z0-9]{64}$
	NameRegexp        = regexp.MustCompile(`^[a-z0-9][a-z0-9-]{0,28}[a-z0-9]$`)
	TagOrBranchRegexp = regexp.MustCompile(`^[a-z0-9][a-z0-9-\\.]{0,28}[a-z0-9]$`)
	UUIDRegexp        = regexp.MustCompile(`^[0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12}$`)
)

// Certificate paths.
const (
	CACertFile  = "ca.crt"
	TLSCertFile = "tls.crt"
	TLSKeyFile  = "tls.key"

	SvcAccTokenFile = "/var/run/secrets/kubernetes.io/serviceaccount/token"
)

var (
	KubeFoxHome = path.Join("/", "tmp", "kubefox")

	CACertPath  = path.Join(KubeFoxHome, CACertFile)
	TLSCertPath = path.Join(KubeFoxHome, TLSCertFile)
	TLSKeyPath  = path.Join(KubeFoxHome, TLSKeyFile)
)
