package operator

import (
	kubev1a1 "github.com/xigxog/kubefox/libs/core/api/kubernetes/v1alpha1"
	"github.com/xigxog/kubefox/libs/core/vault"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	k8sclient "sigs.k8s.io/controller-runtime/pkg/client"
)

type operator struct {
	k8sCli   k8sclient.Client
	vaultCli *vault.Client
}

type Request[T any] struct {
	Object      *T               `json:"object"` // for sync requests
	Parent      *T               `json:"parent"` // for customize requests
	Attachments *RequestChildren `json:"attachments"`
	Related     *RequestChildren `json:"related"`
}

type RequestChildren struct {
	ConfigMaps      map[string]*corev1.ConfigMap      `json:"ConfigMap.v1"`
	Deployments     map[string]*appsv1.Deployment     `json:"Deployment.apps/v1"`
	Namespaces      map[string]*corev1.Namespace      `json:"Namespace.v1"`
	Pods            map[string]*corev1.Pod            `json:"Pod.v1"`
	Secrets         map[string]*corev1.Secret         `json:"Secret.v1"`
	Services        map[string]*corev1.Service        `json:"Service.v1"`
	ServiceAccounts map[string]*corev1.ServiceAccount `json:"ServiceAccount.v1"`

	ComponentSets map[string]*kubev1a1.ComponentSet `json:"ComponentSet.k8s.kubefox.io/v1alpha1"`
}

type KubeResponse struct {
	Attachments []runtime.Object `json:"attachments"`
	Status      any              `json:"status,omitempty"`
}

type CustomizeResponse struct {
	RelatedResourceRules []*RelatedResourceRule `json:"relatedResources,omitempty"`
}

type RelatedResourceRule struct {
	APIVersion    string                `json:"apiVersion"`
	Resource      string                `json:"resource"`
	Namespace     string                `json:"namespace,omitempty"`
	Names         []string              `json:"names,omitempty"`
	LabelSelector *metav1.LabelSelector `json:"labelSelector,omitempty"`
}

func (req *Request[T]) GetObject() *T {
	if req.Object != nil {
		return req.Object
	} else {
		return req.Parent
	}
}
