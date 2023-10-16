package kubefox

import (
	"path"
	"regexp"

	"github.com/xigxog/kubefox/libs/core/utils"
)

// Injected at build time
var (
	ComponentName string
	GitCommit     string
	GitRef        string
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
	LabelOCIComponent       string = "com.xigxog.kubefox.component"
	LabelOCICreated         string = "org.opencontainers.image.created"
	LabelOCIRevision        string = "org.opencontainers.image.revision"
	LabelOCISource          string = "org.opencontainers.image.source"
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
	HeaderAbbrvDep       = "kfd"
	HeaderAbbrvEnv       = "kfe"
	HeaderAbbrvEventType = "kft"
	HeaderDep            = "kf-deployment"
	HeaderEnv            = "kf-environment"
	HeaderEventType      = "kf-type"
	HeaderShortDep       = "kf-dep"
	HeaderShortEnv       = "kf-env"
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
	RegexpImage  = regexp.MustCompile(`^.*:[a-z0-9-]{7}$`)
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
