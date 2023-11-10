package kubernetes

// Kubernetes Labels
const (
	LabelK8sAppName         string = "app.kubernetes.io/name"
	LabelK8sComponent       string = "app.kubernetes.io/component"
	LabelK8sComponentCommit string = "kubefox.xigxog.io/component-commit"
	LabelK8sInstance        string = "app.kubernetes.io/instance"
	LabelK8sPlatform        string = "kubefox.xigxog.io/platform"
	LabelK8sRuntimeVersion  string = "app.kubernetes.io/runtime-version"
)

// Kubernetes Annotations
const (
	AnnotationTemplateData string = "kubefox.xigxog.io/template-data"
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
	EnvNodeName = "KUBEFOX_NODE"
	EnvPodIP    = "KUBEFOX_POD_IP"
	EnvPodName  = "KUBEFOX_POD"
)

type EnvVarType string

const (
	EnvVarTypeArray   EnvVarType = "array"
	EnvVarTypeBoolean EnvVarType = "boolean"
	EnvVarTypeNumber  EnvVarType = "number"
	EnvVarTypeString  EnvVarType = "string"
)

type ComponentType string

const (
	ComponentTypeDatabase ComponentType = "db"
	ComponentTypeGenesis  ComponentType = "genesis"
	ComponentTypeHTTP     ComponentType = "http"
	ComponentTypeKubeFox  ComponentType = "kubefox"
)

type FollowRedirects string

const (
	FollowRedirectsAlways   FollowRedirects = "Always"
	FollowRedirectsNever    FollowRedirects = "Never"
	FollowRedirectsSameHost FollowRedirects = "SameHost"
)
