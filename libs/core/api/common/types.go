package common

// +kubebuilder:object:generate=false
type NamedHashed interface {
	GetName() string
	GetGitHash() string
}

// +kubebuilder:object:generate=false
type ConfigReferrer interface {
	GetConfig() string
	SetConfig(string)
}

// +kubebuilder:object:generate=false
type ConfigIdReferrer interface {
	GetConfigId() string
	SetConfigId(string)
}

// +kubebuilder:object:generate=false
type EnvironmentReferrer interface {
	GetEnvironment() string
	SetEnvironment(string)
}

// +kubebuilder:object:generate=false
type EnvironmentIdReferrer interface {
	GetEnvironmentId() string
	SetEnvironmentId(string)
}

// +kubebuilder:object:generate=false
type SystemReferrer interface {
	GetSystem() string
	SetSystem(string)
}

// +kubebuilder:object:generate=false
type SystemIdReferrer interface {
	GetSystemId() string
	SetSystemId(string)
}

// +kubebuilder:object:generate=false
type InheritsReferrer interface {
	GetInherits() string
	SetInherits(string)
}
