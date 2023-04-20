package kubernetes

import (
	"fmt"

	"k8s.io/apimachinery/pkg/runtime"
)

// Key returns a simple key for the provided object. The key is in the format
// `{kind}/{name}`. For example, `Platform/dev`.
func Key(obj Object) string {
	if name := obj.GetName(); name == "" {
		return Kind(obj)
	} else {
		return fmt.Sprintf("%s/%s", Kind(obj), obj.GetName())
	}
}

// FullKey returns a unique key for the provided object. The key is in the
// format `{group}/{version}/{kind}/{name}`. For example,
// `k8s.kubefox.io/v1alpha1/Platform/dev`.
func FullKey(obj Object) string {
	return fmt.Sprintf("%s/%s", KindKey(obj), obj.GetName())
}

// KindKey returns a key for the provided object kind. The key is in the
// format `{group}/{version}/{kind}`. For example,
// `k8s.kubefox.io/v1alpha1/Platform`.
func KindKey(obj runtime.Object) string {
	if Group(obj) == "" {
		return fmt.Sprintf("%s/%s", Version(obj), Kind(obj))
	} else {
		return fmt.Sprintf("%s/%s/%s", Group(obj), Version(obj), Kind(obj))
	}
}

func Group(obj runtime.Object) string {
	return obj.GetObjectKind().GroupVersionKind().Group
}

func Version(obj runtime.Object) string {
	return obj.GetObjectKind().GroupVersionKind().Version
}

// Kind returns the provided object kind.
func Kind(obj runtime.Object) string {
	return obj.GetObjectKind().GroupVersionKind().Kind
}
