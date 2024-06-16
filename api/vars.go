// Copyright 2023 XigXog
//
// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.
//
// SPDX-License-Identifier: MPL-2.0

package api

import (
	"path"
	"regexp"

	"github.com/xigxog/kubefox/utils"
)

// Misc
const (
	SecretMask             = "••••••"
	MaxEventSizeBytesLimit = 16777216 // 16 MiB
)

var (
	t     = true
	f     = false
	True  = &t
	False = &f
)

// Defaults
const (
	DefaultLogFormat                        = "json"
	DefaultLogLevel                         = "info"
	DefaultMaxEventSizeBytes                = 5242880 // 5 MiB
	DefaultReleaseActivationDeadlineSeconds = 300     // 5 mins
	DefaultReleaseHistoryAgeLimit           = 0
	DefaultReleaseHistoryCountLimit         = 10
	DefaultTimeoutSeconds                   = 30
	DefaultEventListenerSize                = 1
)

// Kubernetes Labels
const (
	LabelK8sAppBranch          string = "kubefox.xigxog.io/app-branch"
	LabelK8sAppCommit          string = "kubefox.xigxog.io/app-commit"
	LabelK8sAppCommitShort     string = "kubefox.xigxog.io/app-commit-short"
	LabelK8sAppName            string = "app.kubernetes.io/name"
	LabelK8sAppTag             string = "kubefox.xigxog.io/app-tag"
	LabelK8sAppVersion         string = "kubefox.xigxog.io/app-version"
	LabelK8sComponent          string = "app.kubernetes.io/component"
	LabelK8sComponentHash      string = "kubefox.xigxog.io/component-hash"
	LabelK8sComponentHashShort string = "kubefox.xigxog.io/component-hash-short"
	LabelK8sComponentType      string = "kubefox.xigxog.io/component-type"
	LabelK8sEnvironment        string = "kubefox.xigxog.io/environment"
	LabelK8sInstance           string = "app.kubernetes.io/instance"
	LabelK8sPlatform           string = "kubefox.xigxog.io/platform"
	LabelK8sRelManifest        string = "kubefox.xigxog.io/release-manifest"
	LabelK8sRuntimeVersion     string = "kubefox.xigxog.io/runtime-version"
	LabelK8sVirtualEnvironment string = "kubefox.xigxog.io/virtual-environment"
)

// Kubernetes Annotations
const (
	AnnotationLastApplied      string = "kubectl.kubernetes.io/last-applied-configuration"
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
	FinalizerEnvironmentProtection string = "kubefox.xigxog.io/environment-protection"
	FinalizerReleaseProtection     string = "kubefox.xigxog.io/release-protection"
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
	ConditionReasonBrokerUnavailable              string = "BrokerUnavailable"
	ConditionReasonComponentDeploymentFailed      string = "ComponentDeploymentFailed"
	ConditionReasonComponentDeploymentProgressing string = "ComponentDeploymentProgressing"
	ConditionReasonComponentsAvailable            string = "ComponentsAvailable"
	ConditionReasonComponentsDeployed             string = "ComponentsDeployed"
	ConditionReasonComponentUnavailable           string = "ComponentUnavailable"
	ConditionReasonContextAvailable               string = "ContextAvailable"
	ConditionReasonEnvironmentNotFound            string = "EnvironmentNotFound"
	ConditionReasonHTTPSrvUnavailable             string = "HTTPSrvUnavailable"
	ConditionReasonNATSUnavailable                string = "NATSUnavailable"
	ConditionReasonNoRelease                      string = "NoRelease"
	ConditionReasonPendingDeadlineExceeded        string = "PendingDeadlineExceeded"
	ConditionReasonPlatformComponentsAvailable    string = "PlatformComponentsAvailable"
	ConditionReasonProblemsFound                  string = "ProblemsFound"
	ConditionReasonReconcileFailed                string = "ReconcileFailed"
	ConditionReasonReleaseActivated               string = "ReleaseActivated"
	ConditionReasonReleasePending                 string = "ReleasePending"
)

// gRPC metadata keys.
const (
	GRPCKeyApp       string = "app"
	GRPCKeyHash      string = "hash"
	GRPCKeyComponent string = "component"
	GRPCKeyId        string = "id"
	GRPCKeyPlatform  string = "platform"
	GRPCKeyPod       string = "pod"
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
	ProblemTypeAdapterNotFound        ProblemType = "AdapterNotFound"
	ProblemTypeAppDeploymentFailed    ProblemType = "AppDeploymentFailed"
	ProblemTypeAppDeploymentNotFound  ProblemType = "AppDeploymentNotFound"
	ProblemTypeDependencyInvalid      ProblemType = "DependencyInvalid"
	ProblemTypeDependencyNotFound     ProblemType = "DependencyNotFound"
	ProblemTypeDeploymentFailed       ProblemType = "DeploymentFailed"
	ProblemTypeDeploymentNotFound     ProblemType = "DeploymentNotFound"
	ProblemTypeDeploymentUnavailable  ProblemType = "DeploymentUnavailable"
	ProblemTypeParseError             ProblemType = "ParseError"
	ProblemTypePolicyViolation        ProblemType = "PolicyViolation"
	ProblemTypeRelManifestFailed      ProblemType = "ReleaseManifestFailed"
	ProblemTypeRelManifestNotFound    ProblemType = "ReleaseManifestNotFound"
	ProblemTypeRelManifestUnavailable ProblemType = "ReleaseManifestUnavailable"
	ProblemTypeRouteConflict          ProblemType = "RouteConflict"
	ProblemTypeVarNotFound            ProblemType = "VarNotFound"
	ProblemTypeVarWrongType           ProblemType = "VarWrongType"
	ProblemTypeVersionConflict        ProblemType = "VersionConflict"
)

type DataSourceKind string

const (
	DataSourceKindVirtualEnvironment DataSourceKind = "VirtualEnvironment"
)

type ProblemSourceKind string

const (
	ProblemSourceKindAppDeployment ProblemSourceKind = "AppDeployment"
	ProblemSourceKindComponent     ProblemSourceKind = "Component"
	ProblemSourceKindDeployment    ProblemSourceKind = "Deployment"
	ProblemSourceKindHTTPAdapter   ProblemSourceKind = "HTTPAdapter"
	ProblemSourceKindRelManifest   ProblemSourceKind = "ReleaseManifest"
	ProblemSourceKindVirtualEnv    ProblemSourceKind = "VirtualEnvironment"
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
	ComponentTypeBroker ComponentType = "Broker"
	// ComponentTypeDatabaseAdapter ComponentType = "DBAdapter"
	ComponentTypeHTTPAdapter ComponentType = "HTTPAdapter"
	ComponentTypeKubeFox     ComponentType = "KubeFox"
	ComponentTypeNATS        ComponentType = "NATS"
)

func (c ComponentType) IsAdapter() bool {
	switch c {
	case ComponentTypeHTTPAdapter:
		return true
	default:
		return false
	}
}

type ReleaseType string

const (
	ReleaseTypeStable  ReleaseType = "Stable"
	ReleaseTypeTesting ReleaseType = "Testing"
)

type FollowRedirects string

const (
	FollowRedirectsAlways   FollowRedirects = "Always"
	FollowRedirectsNever    FollowRedirects = "Never"
	FollowRedirectsSameHost FollowRedirects = "SameHost"
)

type EventType string

const (
	// Component event types
	EventTypeCron       EventType = "io.kubefox.cron"
	EventTypeDapr       EventType = "io.kubefox.dapr"
	EventTypeHTTP       EventType = "io.kubefox.http"
	EventTypeKubeFox    EventType = "io.kubefox.kubefox"
	EventTypeKubernetes EventType = "io.kubefox.kubernetes"

	// Platform event types
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
	ValKeyHeader       = "header"
	ValKeyHost         = "host"
	ValKeyMaxEventSize = "maxEventSize"
	ValKeyMethod       = "method"
	ValKeyPath         = "path"
	ValKeyPathSuffix   = "pathSuffix"
	ValKeyQuery        = "queryParam"
	ValKeyStatus       = "status"
	ValKeyStatusCode   = "statusCode"
	ValKeyURL          = "url"
	ValKeyVaultURL     = "vaultURL"
	ValKeySpec         = "spec"
)

// Headers and query params.
const (
	HeaderAdapter              = "kubefox-adapter"
	HeaderAppDeployment        = "kubefox-app-deployment"
	HeaderAppDeploymentAbbrv   = "kf-dep"
	HeaderContentLength        = "Content-Length"
	HeaderContentType          = "Content-Type"
	HeaderEventId              = "kubefox-event-id"
	HeaderEventType            = "kubefox-event-type"
	HeaderEventTypeAbbrv       = "kf-type"
	HeaderHost                 = "Host"
	HeaderPlatform             = "kubefox-platform"
	HeaderRelManifest          = "kubefox-release-manifest"
	HeaderTelemetrySample      = "kubefox-telemetry-sample"
	HeaderTelemetrySampleAbbrv = "kf-sample"
	HeaderTraceId              = "kubefox-trace-id"
	HeaderVirtualEnv           = "kubefox-virtual-environment"
	HeaderVirtualEnvAbbrv      = "kf-ve"
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
	RegexpHash   = regexp.MustCompile(`^[0-9a-f]{32}$`)
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
