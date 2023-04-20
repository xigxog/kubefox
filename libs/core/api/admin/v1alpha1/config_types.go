package v1alpha1

import (
	"github.com/xigxog/kubefox/libs/core/api/admin"
	"github.com/xigxog/kubefox/libs/core/api/common"
)

type Config struct {
	admin.ObjectBase `json:",inline"`

	common.SecretsProp `json:",inline"`

	Components map[string]*ComponentConfig `json:"components,omitempty" validate:"dive"`

	Status ConfigStatus `json:"status,omitempty"`
}

type ComponentConfig struct {
	common.ComponentTypeProp `json:",inline"`
	common.VarsProp          `json:",inline"`
}

type ConfigStatus struct {
	Releases []*Release `json:"releases,omitempty" validate:"dive"`
}
