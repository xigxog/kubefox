package kubernetes

// Kubernetes labels
const (
	// Name of the Component.
	NameLabel = "app.kubernetes.io/name"
	// Name of Component appended with Platform name.
	InstanceLabel = "app.kubernetes.io/instance"
	// Name of System that the Component is part of.
	PartOfLabel = "app.kubernetes.io/part-of"
	// Git hash of Component.
	VersionLabel = "app.kubernetes.io/version"
)

// KubeFox labels
const (
	// OrganizationLabel = "k8s.kubefox.io/organization"

	EnvLabel    = "k8s.kubefox.io/environment"
	EnvIdLabel  = "k8s.kubefox.io/environment-id"
	EnvRefLabel = "k8s.kubefox.io/environment-ref"

	ConfigLabel    = "k8s.kubefox.io/config"
	ConfigIdLabel  = "k8s.kubefox.io/config-id"
	ConfigRefLabel = "k8s.kubefox.io/config-ref"

	SystemLabel        = "k8s.kubefox.io/system"
	SystemIdLabel      = "k8s.kubefox.io/system-id"
	SystemRefLabel     = "k8s.kubefox.io/system-ref"
	SystemGitHashLabel = "k8s.kubefox.io/system-git-hash"
	SystemGitRefLabel  = "k8s.kubefox.io/system-git-ref"

	ComponentLabel        = "k8s.kubefox.io/component"
	ComponentGitHashLabel = "k8s.kubefox.io/component-git-hash"
)
