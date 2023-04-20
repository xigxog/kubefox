package v1alpha1

import (
	"github.com/xigxog/kubefox/libs/core/api/admin"
	"github.com/xigxog/kubefox/libs/core/api/common"
)

type Platform struct {
	admin.ObjectBase    `json:",inline"`
	common.PlatformSpec `json:",inline"`

	Status common.PlatformStatus `json:"status,omitempty"`
}
