package core

import (
	"path"
	"regexp"

	"github.com/xigxog/kubefox/utils"
)

type EventType string

// Component event types
const (
	EventTypeComponent  EventType = "io.kubefox.component"
	EventTypeCron       EventType = "io.kubefox.cron"
	EventTypeDapr       EventType = "io.kubefox.dapr"
	EventTypeHTTP       EventType = "io.kubefox.http"
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

// Labels
const (
	LabelK8sComponent       string = "kubefox.xigxog.io/component"
	LabelK8sComponentCommit string = "kubefox.xigxog.io/component-commit"
	LabelK8sPlatform        string = "kubefox.xigxog.io/platform"
	LabelK8sKubeFoxVersion  string = "kubefox.xigxog.io/version"
	LabelOCIApp             string = "com.xigxog.kubefox.app"
	LabelOCIComponent       string = "com.xigxog.kubefox.component"
	LabelOCICreated         string = "org.opencontainers.image.created"
	LabelOCIRevision        string = "org.opencontainers.image.revision"
	LabelOCISource          string = "org.opencontainers.image.source"
)

const (
	EnvNodeName = "KUBEFOX_NODE"
	EnvPodIP    = "KUBEFOX_POD_IP"
	EnvPodName  = "KUBEFOX_POD"
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
	HeaderDep            = "kubefox-deployment"
	HeaderEnv            = "kubefox-environment"
	HeaderEventType      = "kubefox-type"
	HeaderShortDep       = "kfd"
	HeaderShortEnv       = "kfe"
	HeaderShortEventType = "kft"
	HeaderTraceId        = "kubefox-trace-id"
)

const (
	CharSetUTF8 = "charset=UTF-8"

	DataSchemaKubefox = "xigxog.proto.v1.KubeFoxData"

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
