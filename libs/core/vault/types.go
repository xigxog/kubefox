package vault

import (
	"fmt"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

const (
	SysKVPrefix = "kfs"

	BrkPolicy      = "kfp-broker-policy"
	BrkRole        = "kfp-broker-role"
	PlatformPolicy = "kfp-platform-policy"
	PlatformRole   = "kfp-platform-role"
)

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

	return fmt.Sprintf(SysKVPrefix+"/%s", platform)
}
