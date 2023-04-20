package kubernetes

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
)

const KubeFoxGroup = "k8s.kubefox.io"

type Object interface {
	runtime.Object
	metav1.Object
}

type KubeFoxObject interface {
	Object

	GetSpec() any
}
