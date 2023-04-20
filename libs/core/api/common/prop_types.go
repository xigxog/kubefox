// +kubebuilder:object:generate=true
package common

import (
	"fmt"
)

// TODO temp removing org concept for simplicity, will need to be added back.
// type OrganizationProp struct {
// 	// +kubebuilder:validation:Pattern="^[a-z0-9][a-z0-9-]{0,28}[a-z0-9]$"
// 	Organization string `json:"organization" validate:"required,objName"`
// }

// func (p *OrganizationProp) GetOrganization() string {
// 	return p.Organization
// }

// func (p *OrganizationProp) SetOrganization(val string) {
// 	p.Organization = val
// }

type EnvironmentProp struct {
	Environment string `json:"environment" validate:"required,environmentRef"`
}

func (p *EnvironmentProp) GetEnvironment() string {
	return p.Environment
}

func (p *EnvironmentProp) SetEnvironment(val string) {
	p.Environment = val
}

type EnvironmentIdProp struct {
	EnvironmentId string `json:"environmentId" validate:"required,environmentIdRef"`
}

func (p *EnvironmentIdProp) GetEnvironmentId() string {
	return p.EnvironmentId
}

func (p *EnvironmentIdProp) SetEnvironmentId(val string) {
	p.EnvironmentId = val
}

type ConfigProp struct {
	Config string `json:"config" validate:"required,configRef"`
}

func (p *ConfigProp) GetConfig() string {
	return p.Config
}

func (p *ConfigProp) SetConfig(val string) {
	p.Config = val
}

type SystemProp struct {
	System string `json:"system" validate:"required,systemRef"`
}

func (p *SystemProp) GetSystem() string {
	return p.System
}

func (p *SystemProp) SetSystem(val string) {
	p.System = val
}

type SystemIdProp struct {
	SystemId string `json:"systemId" validate:"required,systemIdRef"`
}

func (p *SystemIdProp) GetSystemId() string {
	return p.SystemId
}

func (p *SystemIdProp) SetSystemId(val string) {
	p.SystemId = val
}

type InheritsProp struct {
	Inherits string `json:"inherits,omitempty" validate:"omitempty,environmentRef"`
}

func (p *InheritsProp) GetInherits() string {
	return p.Inherits
}

func (p *InheritsProp) SetInherits(val string) {
	p.Inherits = val
}

type GitRepoProp struct {
	// +kubebuilder:validation:Format=uri
	GitRepo string `json:"gitRepo" validate:"required,uri"`
}

type GitHashProp struct {
	// +kubebuilder:validation:Pattern="^[a-z0-9]{7}$"
	GitHash string `json:"gitHash" validate:"required,gitHash"`
}

// TODO add regexp
type GitRefProp struct {
	// +kubebuilder:validation:MinLength=1
	GitRef string `json:"gitRef" validate:"required"`
}

// TODO use SHA256, switch pattern to ^.*@sha256:[a-z0-9]{64}$
type ImageProp struct {
	// +kubebuilder:validation:Pattern="^.*:[a-z0-9-]{7}$"
	Image string `json:"image" validate:"required,componentImage"`
}

// TODO create enum with custom json marshalling
type ComponentTypeProp struct {
	// +kubebuilder:validation:Enum=graphql;http;kubefox;k8s;kv;object
	Type string `json:"type" validate:"required,oneof=graphql http kubefox k8s kv object"`
}

type VarsProp struct {
	// +kubebuilder:validation:Schemaless
	// +kubebuilder:validation:Type=object
	Vars map[string]*Var `json:"vars,omitempty" validate:"dive"`
}

type SecretsProp struct {
	// +kubebuilder:validation:Schemaless
	// +kubebuilder:validation:Type=object
	Secrets map[string]*Var `json:"secrets,omitempty" validate:"dive"`
}

type ComponentKey string

type ComponentProps struct {
	ComponentTypeProp `json:",inline"`
	GitHashProp       `json:",inline"`
	ImageProp         `json:",inline"`

	Name string `json:"name,omitempty"`
}

func (p *ComponentProps) ShortHash() string {
	if len(p.GitHash) < 7 {
		return p.GitHash

	}
	return p.GitHash[0:7]
}

func (p *ComponentProps) Key() ComponentKey {
	return ComponentKey(fmt.Sprintf("%s-%s", p.Name, p.ShortHash()))
}

type RouteTypeProp struct {
	// +kubebuilder:validation:Enum=controller;cron;http
	Type string `json:"type" validate:"required,oneof=controller cron http"`
}

type VarTypeProp struct {
	// +kubebuilder:validation:Enum=array;boolean;number;string
	Type string `json:"type" validate:"required,oneof=array boolean number string"`
}
