package core

import (
	"path"
	"regexp"

	"github.com/xigxog/kubefox/utils"
)

const (
	// Large events reduce performance. Do not exceed 32 MiB.
	MaxContentSizeBytes = 3145728 // 3 MiB
)

type EventType string

// Component event types
const (
	EventTypeCron       EventType = "io.kubefox.cron"
	EventTypeDapr       EventType = "io.kubefox.dapr"
	EventTypeHTTP       EventType = "io.kubefox.http"
	EventTypeKubeFox    EventType = "io.kubefox.kubefox"
	EventTypeKubernetes EventType = "io.kubefox.kubernetes"
)

// Platform event types
const (
	EventTypeAck       EventType = "io.kubefox.ack"
	EventTypeBootstrap EventType = "io.kubefox.bootstrap"
	EventTypeError     EventType = "io.kubefox.error"
	EventTypeHealth    EventType = "io.kubefox.health"
	EventTypeMetrics   EventType = "io.kubefox.metrics"
	EventTypeNack      EventType = "io.kubefox.nack"
	EventTypeRegister  EventType = "io.kubefox.register"
	EventTypeRejected  EventType = "io.kubefox.rejected"
	EventTypeTelemetry EventType = "io.kubefox.telemetry"
	EventTypeUnknown   EventType = "io.kubefox.unknown"
)

// Keys for well known values.
const (
	ValKeyHeader     = "header"
	ValKeyHost       = "host"
	ValKeyMethod     = "method"
	ValKeyPath       = "path"
	ValKeyQuery      = "queryParam"
	ValKeyStatusCode = "statusCode"
	ValKeyStatus     = "status"
	ValKeyURL        = "url"
	ValKeyTraceId    = "traceId"
	ValKeySpanId     = "spanId"
	ValKeyTraceFlags = "traceFlags"
)

// Headers and query params.
const (
	HeaderAbbrvDep       = "kf-dep"
	HeaderAbbrvEnv       = "kf-env"
	HeaderAbbrvEventType = "kf-type"
	HeaderAdapter        = "kubefox-adapter"
	HeaderContentLength  = "Content-Length"
	HeaderDep            = "kubefox-deployment"
	HeaderEnv            = "kubefox-environment"
	HeaderEventType      = "kubefox-type"
	HeaderHost           = "Host"
	HeaderShortDep       = "kfd"
	HeaderShortEnv       = "kfe"
	HeaderShortEventType = "kft"
	HeaderTraceId        = "kubefox-trace-id"
)

const (
	CharSetUTF8 = "charset=UTF-8"

	DataSchemaEvent = "kubefox.proto.v1.Event"

	ContentTypeHTML     = "text/html"
	ContentTypeJSON     = "application/json"
	ContentTypePlain    = "text/plain"
	ContentTypeProtobuf = "application/protobuf"
)

var (
	RegexpCommit = regexp.MustCompile(`^[0-9a-f]{40}$`)
	RegexpGitRef = regexp.MustCompile(`^[a-z0-9][a-z0-9-\\.]{0,28}[a-z0-9]$`)
	RegexpImage  = regexp.MustCompile(`^.*:[a-z0-9-]{40}$`)
	RegexpName   = regexp.MustCompile(`^[a-z0-9][a-z0-9-]{0,28}[a-z0-9]$`)
	RegexpUUID   = regexp.MustCompile(`^[0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12}$`)
)

var (
	KubeFoxHome = utils.EnvDef("KUBEFOX_HOME", path.Join("/", "tmp", "kubefox"))

	FileCACert  = "ca.crt"
	FileTLSCert = "tls.crt"
	FileTLSKey  = "tls.key"

	PathCACert      = path.Join(KubeFoxHome, FileCACert)
	PathSvcAccToken = "/var/run/secrets/kubernetes.io/serviceaccount/token"
	PathTLSCert     = path.Join(KubeFoxHome, FileTLSCert)
	PathTLSKey      = path.Join(KubeFoxHome, FileTLSKey)
)

const (
	DefaultRouteId = -1
)

const (
	CloudEventId = "ce_id"
)
