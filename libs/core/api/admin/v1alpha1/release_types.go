package v1alpha1

import (
	"github.com/xigxog/kubefox/libs/core/api/admin"
	"github.com/xigxog/kubefox/libs/core/api/common"
	"github.com/xigxog/kubefox/libs/core/api/uri"
)

type Release struct {
	admin.SubObjectBase `json:",inline"`

	common.SystemProp      `json:",inline"`
	common.EnvironmentProp `json:",inline"`

	Status common.ReleaseStatus `json:"status,omitempty"`
}

func (r *Release) GetURI(org, platform string) (uri.URI, error) {
	return uri.New(org, uri.Platform, platform, uri.Release, r.GetSystem(), r.GetEnvironment())
}
