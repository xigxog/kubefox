// Copyright 2024 XigXog
//
// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.
//
// SPDX-License-Identifier: MPL-2.0

package telemetry

const (
	// KubeFox Attribute Keys
	AttrKeyComponentApp       = "kubefox.component.app"
	AttrKeyComponentHash      = "kubefox.component.hash"
	AttrKeyComponentId        = "kubefox.component.id"
	AttrKeyComponentName      = "kubefox.component.name"
	AttrKeyComponentType      = "kubefox.component.type"
	AttrKeyEventAppDeployment = "kubefox.event.context.app_deployment"
	AttrKeyEventCategory      = "kubefox.event.category"
	AttrKeyEventId            = "kubefox.event.id"
	AttrKeyEventParentId      = "kubefox.event.parent_id"
	AttrKeyEventRelManifest   = "kubefox.event.context.release_manifest"
	AttrKeyEventSourceHash    = "kubefox.event.source.hash"
	AttrKeyEventSourceId      = "kubefox.event.source.id"
	AttrKeyEventSourceName    = "kubefox.event.source.name"
	AttrKeyEventSourceType    = "kubefox.event.source.type"
	AttrKeyEventTargetHash    = "kubefox.event.target.hash"
	AttrKeyEventTargetId      = "kubefox.event.target.id"
	AttrKeyEventTargetName    = "kubefox.event.target.name"
	AttrKeyEventTargetType    = "kubefox.event.target.type"
	AttrKeyEventTTL           = "kubefox.event.ttl"
	AttrKeyEventType          = "kubefox.event.type"
	AttrKeyEventVirtualEnv    = "kubefox.event.context.virtual_environment"
	AttrKeyInstance           = "kubefox.instance"
	AttrKeyPlatform           = "kubefox.platform"
	AttrKeyRouteId            = "kubefox.route.id"

	// OTEL Attribute Keys
	AttrKeySDKLang    = "telemetry.sdk.language" // Required
	AttrKeySDKName    = "telemetry.sdk.name"     // Required
	AttrKeySDKVersion = "telemetry.sdk.version"  // Required
	AttrKeySvcName    = "service.name"           // Required

	AttrKeyCloudAccountId        = "cloud.account.id"
	AttrKeyCloudAZ               = "cloud.availability_zone"
	AttrKeyCloudPlatform         = "cloud.platform"
	AttrKeyCloudProvider         = "cloud.provider"
	AttrKeyCloudRegion           = "cloud.region"
	AttrKeyCloudResourceId       = "cloud.resource_id"
	AttrKeyCodeColumn            = "code.column"
	AttrKeyCodeFilepath          = "code.filepath"
	AttrKeyCodeFunction          = "code.function"
	AttrKeyCodeLineNo            = "code.lineno"
	AttrKeyCodeNamespace         = "code.namespace"
	AttrKeyCodeStacktrace        = "code.stacktrace"
	AttrKeyContainerArgs         = "container.command_args"
	AttrKeyContainerCommand      = "container.command"
	AttrKeyContainerId           = "container.id"
	AttrKeyContainerImageDigest  = "container.image.repo_digests"
	AttrKeyContainerImageId      = "container.image.id"
	AttrKeyContainerImageName    = "container.image.name"
	AttrKeyContainerName         = "container.name"
	AttrKeyErrType               = "error.type"
	AttrKeyExceptionMsg          = "exception.message"
	AttrKeyExceptionStacktrace   = "exception.stacktrace"
	AttrKeyExceptionType         = "exception.type"
	AttrKeyGraphQLDocument       = "graphql.document"
	AttrKeyGraphQLOpName         = "graphql.operation.name"
	AttrKeyGraphQLOpType         = "graphql.operation.type" // query, mutation, subscription
	AttrKeyGRPCStatusCode        = "rpc.grpc.status_code"
	AttrKeyHTTPReqBodySize       = "http.request.body.size"
	AttrKeyHTTPReqMethod         = "http.request.method"
	AttrKeyHTTPRespBodySize      = "http.response.body.size"
	AttrKeyHTTPRespStatusCode    = "http.response.status_code"
	AttrKeyHTTPRoute             = "http.route"
	AttrKeyK8sClusterId          = "k8s.cluster.uid"
	AttrKeyK8sClusterName        = "k8s.cluster.name"
	AttrKeyK8sContainerName      = "k8s.container.name"
	AttrKeyK8sContainerRestart   = "k8s.container.restart_count"
	AttrKeyK8sNamespace          = "k8s.namespace.name"
	AttrKeyK8sNodeId             = "k8s.node.uid"
	AttrKeyK8sNodeName           = "k8s.node.name"
	AttrKeyK8sPodId              = "k8s.pod.uid"
	AttrKeyK8sPodName            = "k8s.pod.name"
	AttrKeyMsgId                 = "message.id"
	AttrKeyMsgSize               = "message.uncompressed_size"
	AttrKeyMsgType               = "message.type" // SENT, RECEIVED
	AttrKeyNetworkProtocol       = "network.protocol.name"
	AttrKeyOCIManifestDigest     = "oci.manifest.digest"
	AttrKeyOTELStatusCode        = "otel.status_code" // OK, ERROR
	AttrKeyOTELStatusDescription = "otel.status_description"
	AttrKeySvcInstanceId         = "service.instance.id"
	AttrKeySvcNamespace          = "service.namespace"
	AttrKeySvcVersion            = "service.version"
	AttrKeyThreadId              = "thread.id"
	AttrKeyThreadName            = "thread.name"
	AttrKeyURLFull               = "url.full"
	AttrKeyURLPath               = "url.path"
	AttrKeyURLQuery              = "url.query"
	AttrKeyURLScheme             = "url.scheme"
	AttrKeyUserAgent             = "user_agent.original"

	// Event names
	EventNameException = "exception"
)
