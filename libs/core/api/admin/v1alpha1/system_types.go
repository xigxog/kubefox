package v1alpha1

import (
	"regexp"

	"github.com/xigxog/kubefox/libs/core/api/admin"
	"github.com/xigxog/kubefox/libs/core/api/common"
	"github.com/xigxog/kubefox/libs/core/utils"
)

type System struct {
	admin.ObjectBase `json:",inline"`

	common.GitRepoProp `json:",inline"`
	common.GitHashProp `json:",inline"`
	common.GitRefProp  `json:",inline"`

	Message string                 `json:"message,omitempty"`
	Apps    map[string]*common.App `json:"apps" validate:"dive"`

	Status SystemStatus `json:"status,omitempty"`
}

type SystemStatus struct {
	Deployments []string   `json:"deployments,omitempty"`
	Releases    []*Release `json:"releases,omitempty" validate:"dive"`
}

func (s *System) GetNameRegExp() *regexp.Regexp {
	return utils.HashRegexp
}
