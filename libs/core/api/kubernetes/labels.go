package kubernetes

// Kubernetes labels
const (
	InstanceLabel = "app.kubernetes.io/instance"
	NameLabel     = "app.kubernetes.io/name"
)

// KubeFox labels
const (
	CompLabel = "app.kubernetes.io/component"
	// OrganizationLabel = "k8s.kubefox.io/organization"
	PlatformLabel = "k8s.kubefox.io/platform"

	EnvironmentLabel = "k8s.kubefox.io/environment"
	EnvRefLabel      = "k8s.kubefox.io/environment-ref"
	EnvIdLabel       = "k8s.kubefox.io/environment-id"

	ConfigLabel    = "k8s.kubefox.io/config"
	ConfigRefLabel = "k8s.kubefox.io/config-ref"
	ConfigIdLabel  = "k8s.kubefox.io/config-id"

	SystemLabel    = "k8s.kubefox.io/system"
	SystemRefLabel = "k8s.kubefox.io/system-ref"
	SystemIdLabel  = "k8s.kubefox.io/system-id"

	AppLabel       = "k8s.kubefox.io/app"
	ComponentLabel = "k8s.kubefox.io/component"
	CompHashLabel  = "k8s.kubefox.io/component-hash"
)
