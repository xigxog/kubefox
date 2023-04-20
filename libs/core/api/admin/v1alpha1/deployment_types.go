package v1alpha1

import (
	"github.com/xigxog/kubefox/libs/core/api/admin"
	"github.com/xigxog/kubefox/libs/core/api/common"
	"github.com/xigxog/kubefox/libs/core/api/uri"
)

type Deployment struct {
	admin.SubObjectBase `json:",inline"`

	common.SystemProp `json:",inline"`

	Status DeploymentStatus `json:"status,omitempty"`
}

type DeploymentStatus struct {
	Ready      bool                                            `json:"ready"`
	Components map[common.ComponentKey]*common.ComponentStatus `json:"components,omitempty" validate:"dive"`
}

func (d *Deployment) GetURI(org, platform string) (uri.URI, error) {
	return uri.New(org, uri.Platform, platform, uri.Deployment, d.GetSystem())
}
