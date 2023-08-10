// +kubebuilder:object:generate=true
package common

import (
	"fmt"

	"github.com/xigxog/kubefox/libs/core/api/uri"
)

type PlatformSpec struct {
	Systems map[uri.Key]*PlatformSystem `json:"systems" validate:"dive"`
}

type PlatformSystem struct {
	ContainerRegistry string `json:"containerRegistry,omitempty"`
	ImagePullSecret   string `json:"imagePullSecret,omitempty"`
}

type PlatformStatus struct {
	Healthy bool                              `json:"healthy"`
	Systems map[uri.Key]*PlatformSystemStatus `json:"systems,omitempty" validate:"dive"`
}

type PlatformSystemStatus struct {
	Healthy bool `json:"healthy"`
}

type ReleaseStatus struct {
	Ready bool `json:"ready"`
}

type DeploymentStatus struct {
	Ready bool `json:"ready"`
}

type ComponentStatus struct {
	Deployments []uri.Key `json:"deployments,omitempty"`
	Ready       bool      `json:"ready"`
}

type Fabric struct {
	System *FabricSystem `json:"system" validate:"required,dive"`
	Env    *FabricEnv    `json:"environment" validate:"required,dive"`
}

type FabricSystem struct {
	SystemIdProp `json:",inline"`
	SystemProp   `json:",inline"`

	App FabricApp `json:"app" validate:"required,dive"`
}

type FabricEnv struct {
	EnvironmentIdProp `json:",inline"`
	EnvironmentProp   `json:",inline"`

	Config  map[string]*Var `json:"config" validate:"required,dive"`
	Secrets map[string]*Var `json:"secrets" validate:"required,dive"`
	EnvVars map[string]*Var `json:"envVars" validate:"required,dive"`
}

type FabricApp struct {
	App

	Name string `json:"name" validate:"required"`
}

func (f *Fabric) CheckComponent(name, gitHash string) error {
	ac := f.GetAppComponent(name)
	if ac == nil || (gitHash != "" && gitHash != ac.GitHash) {
		return fmt.Errorf("component %s:%s not found in fabric", name, gitHash)
	}

	return nil
}

func (f *Fabric) GetAppComponent(name string) *AppComponent {
	return f.System.App.Components[name]
}
