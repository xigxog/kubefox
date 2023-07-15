package vault

import (
	"fmt"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

const (
	SystemKVPrefix = "system"

	APISrvRole   = "kubefox-api-server"
	BrokerRole   = "kubefox-broker"
	OperatorRole = "kubefox-operator"
)

const (
	Create UpdateOp = iota
	Put
	Patch
)

type UpdateOp uint8

func (o UpdateOp) String() string {
	switch o {
	case Create:
		return "creating"
	case Put:
		return "putting"
	case Patch:
		return "patching"
	default:
		return "shrugging"
	}
}

type Index map[string]*IndexEntry

type IndexEntry struct {
	*metav1.TypeMeta   `json:",inline"`
	*metav1.ObjectMeta `json:"metadata,omitempty"`
}

type Secret struct {
	Data *Data `json:"data"`
}

type Data struct {
	Data any `json:"data"`
}

type Object struct {
	Object any `json:"object"`
}

type VaultKeys struct {
	Keys []string `json:"keys"`
}

func MountPath(platform string) string {
	if platform == "" {
		return ""
	}

	return fmt.Sprintf(SystemKVPrefix+"/%s", platform)
}
