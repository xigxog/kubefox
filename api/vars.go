package api

import (
	"path"
	"regexp"

	"github.com/xigxog/kubefox/utils"
)

const (
	DefaultLogFormat                     = "json"
	DefaultLogLevel                      = "info"
	DefaultMaxEventSizeBytes             = 5242880 // 5 MiB
	DefaultReleaseHistoryLimitCount      = 10
	DefaultReleasePendingDeadlineSeconds = 300 // 5 mins
	DefaultTimeoutSeconds                = 30

	MaximumMaxEventSizeBytes = 16777216 // 16 MiB
)

// Kubernetes Labels
const (
	LabelK8sAppBranch             string = "kubefox.xigxog.io/app-branch"
	LabelK8sAppCommit             string = "kubefox.xigxog.io/app-commit"
	LabelK8sAppCommitShort        string = "kubefox.xigxog.io/app-commit-short"
	LabelK8sAppComponent          string = "kubefox.xigxog.io/app-component"
	LabelK8sAppDeployment         string = "kubefox.xigxog.io/app-deployment"
	LabelK8sAppName               string = "app.kubernetes.io/name"
	LabelK8sAppTag                string = "kubefox.xigxog.io/app-tag"
	LabelK8sAppVersion            string = "kubefox.xigxog.io/app-version"
	LabelK8sComponent             string = "app.kubernetes.io/component"
	LabelK8sComponentCommit       string = "kubefox.xigxog.io/component-commit"
	LabelK8sComponentCommitShort  string = "kubefox.xigxog.io/component-commit-short"
	LabelK8sComponentType         string = "kubefox.xigxog.io/component-type"
	LabelK8sInstance              string = "app.kubernetes.io/instance"
	LabelK8sPlatform              string = "kubefox.xigxog.io/platform"
	LabelK8sPlatformComponent     string = "kubefox.xigxog.io/platform-component"
	LabelK8sReleaseStatus         string = "kubefox.xigxog.io/release-status"
	LabelK8sRuntimeVersion        string = "kubefox.xigxog.io/runtime-version"
	LabelK8sSourceResourceVersion string = "kubefox.xigxog.io/source-resource-version"
	LabelK8sVirtualEnv            string = "kubefox.xigxog.io/virtual-env"
	LabelK8sVirtualEnvParent      string = "kubefox.xigxog.io/virtual-env-parent"
	LabelK8sVirtualEnvSnapshot    string = "kubefox.xigxog.io/virtual-env-snapshot"
)

// Kubernetes Annotations
const (
	AnnotationTemplateData     string = "kubefox.xigxog.io/template-data"
	AnnotationTemplateDataHash string = "kubefox.xigxog.io/template-data-hash"
)

// Container Labels
const (
	LabelOCIApp       string = "com.xigxog.kubefox.app"
	LabelOCIComponent string = "com.xigxog.kubefox.component"
	LabelOCICreated   string = "org.opencontainers.image.created"
	LabelOCIRevision  string = "org.opencontainers.image.revision"
	LabelOCISource    string = "org.opencontainers.image.source"
)

const (
	FinalizerReleaseProtection string = "kubefox.xigxog.io/release-protection"
)

const (
	EnvNodeName = "KUBEFOX_NODE"
	EnvPodIP    = "KUBEFOX_POD_IP"
	EnvPodName  = "KUBEFOX_POD"
)

const (
	PlatformComponentBootstrap string = "bootstrap"
	PlatformComponentBroker    string = "broker"
	PlatformComponentHTTPSrv   string = "httpsrv"
	PlatformComponentNATS      string = "nats"
	PlatformComponentOperator  string = "operator"
)

const (
	ConditionTypeAvailable              string = "Available"
	ConditionTypeProgressing            string = "Progressing"
	ConditionTypeActiveReleaseAvailable string = "ActiveReleaseAvailable"
	ConditionTypeReleasePending         string = "ReleasePending"
)

const (
	ConditionReasonAppDeploymentAvailable         string = "AppDeploymentAvailable"
	ConditionReasonBrokerUnavailable              string = "BrokerUnavailable"
	ConditionReasonComponentDeploymentFailed      string = "ComponentDeploymentFailed"
	ConditionReasonComponentDeploymentProgressing string = "ComponentDeploymentProgressing"
	ConditionReasonComponentsAvailable            string = "ComponentsAvailable"
	ConditionReasonComponentsDeployed             string = "ComponentsDeployed"
	ConditionReasonComponentUnavailable           string = "ComponentUnavailable"
	ConditionReasonHTTPSrvUnavailable             string = "HTTPSrvUnavailable"
	ConditionReasonNATSUnavailable                string = "NATSUnavailable"
	ConditionReasonNoRelease                      string = "NoRelease"
	ConditionReasonPendingDeadlineExceeded        string = "PendingDeadlineExceeded"
	ConditionReasonPlatformComponentsAvailable    string = "PlatformComponentsAvailable"
	ConditionReasonProblemsExist                  string = "ProblemsExist"
	ConditionReasonReconcileFailed                string = "ReconcileFailed"
	ConditionReasonReleaseActive                  string = "ReleaseActive"
	ConditionReasonReleasePending                 string = "ReleasePending"
)

// gRPC metadata keys.
const (
	GRPCKeyApp       string = "app"
	GRPCKeyCommit    string = "commit"
	GRPCKeyComponent string = "component"
	GRPCKeyId        string = "id"
	GRPCKeyName      string = "name"
	GRPCKeyPlatform  string = "platform"
	GRPCKeyToken     string = "token"
	GRPCKeyType      string = "type"
)

type ArchiveReason string

const (
	ArchiveReasonPendingDeadlineExceeded ArchiveReason = "PendingDeadlineExceeded"
	ArchiveReasonRolledBack              ArchiveReason = "RolledBack"
	ArchiveReasonSuperseded              ArchiveReason = "Superseded"
)

type ProblemType string

const (
	ProblemTypeAdapterNotFound          ProblemType = "AdapterNotFound"
	ProblemTypeAppDeploymentFailed      ProblemType = "AppDeploymentFailed"
	ProblemTypeDependencyInvalid        ProblemType = "DependencyInvalid"
	ProblemTypeDependencyNotFound       ProblemType = "DependencyNotFound"
	ProblemTypeParseError               ProblemType = "ParseError"
	ProblemTypePolicyViolation          ProblemType = "PolicyViolation"
	ProblemTypeRouteConflict            ProblemType = "RouteConflict"
	ProblemTypeVarNotFound              ProblemType = "VarNotFound"
	ProblemTypeVarWrongType             ProblemType = "VarWrongType"
	ProblemTypeVirtualEnvSnapshotFailed ProblemType = "VirtualEnvSnapshotFailed"
)

type ProblemSourceKind string

const (
	ProblemSourceKindAppDeployment      ProblemSourceKind = "AppDeployment"
	ProblemSourceKindComponent          ProblemSourceKind = "Component"
	ProblemSourceKindHTTPAdapter        ProblemSourceKind = "HTTPAdapter"
	ProblemSourceKindVirtualEnv         ProblemSourceKind = "VirtualEnv"
	ProblemSourceKindVirtualEnvSnapshot ProblemSourceKind = "VirtualEnvSnapshot"
)

type EnvVarType string

const (
	EnvVarTypeArray   EnvVarType = "Array"
	EnvVarTypeBoolean EnvVarType = "Boolean"
	EnvVarTypeNumber  EnvVarType = "Number"
	EnvVarTypeString  EnvVarType = "String"
)

type ComponentType string

const (
	ComponentTypeBroker          ComponentType = "Broker"
	ComponentTypeDatabaseAdapter ComponentType = "DBAdapter"
	ComponentTypeHTTPAdapter     ComponentType = "HTTPAdapter"
	ComponentTypeKubeFox         ComponentType = "KubeFox"
	ComponentTypeNATS            ComponentType = "NATS"
)

func (c ComponentType) IsAdapter() bool {
	switch c {
	case ComponentTypeBroker, ComponentTypeHTTPAdapter:
		return true
	default:
		return false
	}
}

type FollowRedirects string

const (
	FollowRedirectsAlways   FollowRedirects = "Always"
	FollowRedirectsNever    FollowRedirects = "Never"
	FollowRedirectsSameHost FollowRedirects = "SameHost"
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

type AppDeploymentPolicy string

const (
	AppDeploymentPolicyVersionOptional AppDeploymentPolicy = "VersionOptional"
	AppDeploymentPolicyVersionRequired AppDeploymentPolicy = "VersionRequired"
)

type VirtualEnvPolicy string

const (
	VirtualEnvPolicySnapshotOptional VirtualEnvPolicy = "SnapshotOptional"
	VirtualEnvPolicySnapshotRequired VirtualEnvPolicy = "SnapshotRequired"
)

// Keys for well known values.
const (
	ValKeyHeader       = "header"
	ValKeyHost         = "host"
	ValKeyMaxEventSize = "maxEventSize"
	ValKeyMethod       = "method"
	ValKeyPath         = "path"
	ValKeyQuery        = "queryParam"
	ValKeySpanId       = "spanId"
	ValKeyStatus       = "status"
	ValKeyStatusCode   = "statusCode"
	ValKeyTraceFlags   = "traceFlags"
	ValKeyTraceId      = "traceId"
	ValKeyURL          = "url"
	ValKeyVaultURL     = "vaultURL"
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
