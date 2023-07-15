package templates

import (
	"testing"

	"github.com/xigxog/kubefox/libs/core/platform"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func TestRenderCRDs(t *testing.T) {
	if s, err := renderStr("crds", nil); err != nil {
		t.Errorf("%v", err)
	} else {
		t.Log(s)
	}
}

func TestRenderVault(t *testing.T) {
	d := &Data{
		Platform: Platform{
			Name:      "dev",
			Version:   "v0.0.1",
			Namespace: "kubefox-system",
		},
		System: System{
			Name:      platform.System,
			Namespace: "kubefox-system",
			Ref:       "main",
			GitHash:   "abcdef",
		},
		Component: Component{
			Name:  "vault",
			Image: "ghcr.io/xigxog/vault:1.13.3-v0.0.1",
		},
		Owner: &metav1.OwnerReference{
			APIVersion: "k8s.kubefox.io/v1alpha1",
			Kind:       "Platform",
			UID:        "123",
			Name:       "kubefox-dev",
		},
	}
	if s, err := renderStr("vault", d); err != nil {
		t.Errorf("%v", err)
	} else {
		t.Log(s)
	}
}

func TestRenderTraefik(t *testing.T) {
	d := &Data{
		Platform: Platform{
			Name:      "dev",
			Version:   "v0.0.1",
			Namespace: "kubefox-system",
		},
		System: System{
			Name:      platform.System,
			Namespace: "kubefox-system",
			Ref:       "main",
			GitHash:   "abcdef",
		},
		Broker: Broker{
			Type: HTTPServerType,
		},
		Component: Component{
			Name:  "traefik",
			Image: "traefik:v2.10",
		},
		Owner: &metav1.OwnerReference{
			APIVersion: "k8s.kubefox.io/v1alpha1",
			Kind:       "Platform",
			UID:        "123",
			Name:       "kubefox-dev",
		},
	}
	if s, err := renderStr("traefik", d); err != nil {
		t.Errorf("%v", err)
	} else {
		t.Logf("\n%s", s)
	}
}
