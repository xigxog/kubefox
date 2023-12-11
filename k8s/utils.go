package k8s

import (
	"errors"
	"fmt"

	"github.com/xigxog/kubefox/api"
	"github.com/xigxog/kubefox/core"
	"github.com/xigxog/kubefox/utils"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/equality"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
)

func init() {
	equality.Semantic.AddFunc(func(lhs, rhs *api.Val) bool {
		return lhs.Equals(rhs)
	})
}

func Key(namespace, name string) types.NamespacedName {
	return types.NamespacedName{
		Namespace: namespace,
		Name:      name,
	}
}

func HasLabel(obj client.Object, key, value string) bool {
	if obj == nil {
		return false
	}

	value = utils.CleanLabel(value)
	if obj.GetLabels() == nil {
		obj.SetLabels(make(map[string]string))
	}

	if curVal, found := obj.GetLabels()[key]; curVal == value {
		return true
	} else if !found && value == "" {
		return true
	}

	return false
}

func RemoveLabel(obj client.Object, key string) bool {
	return UpdateLabel(obj, key, "")
}

func UpdateLabel(obj client.Object, key, value string) bool {
	if obj == nil {
		return false
	}

	value = utils.CleanLabel(value)
	if HasLabel(obj, key, value) {
		return false
	}

	if value != "" {
		obj.GetLabels()[key] = value
	} else {
		delete(obj.GetLabels(), key)
	}

	return true
}

func AddFinalizer(o client.Object, finalizer string) bool {
	return controllerutil.AddFinalizer(o, finalizer)
}

func RemoveFinalizer(o client.Object, finalizer string) bool {
	return controllerutil.RemoveFinalizer(o, finalizer)
}

func DeepEqual(lhs interface{}, rhs interface{}) bool {
	return equality.Semantic.DeepEqual(lhs, rhs)
}

func IgnoreNotFound(err error) error {
	if IsNotFound(err) {
		return nil
	}
	return err
}

func IsNotFound(err error) bool {
	if err == nil {
		return false
	}
	return apierrors.IsNotFound(err) || errors.Is(err, core.ErrNotFound())
}

func IsAlreadyExists(err error) bool {
	if err == nil {
		return false
	}
	return apierrors.IsAlreadyExists(err)
}

func IsConflict(err error) bool {
	if err == nil {
		return false
	}
	return apierrors.IsConflict(err)
}

func ToString(obj client.Object) string {
	gvk := obj.GetObjectKind().GroupVersionKind()
	grp := gvk.Group
	if grp == "" {
		grp = "core"
	}
	return fmt.Sprintf("%s/%s/%s/%s/%s", obj.GetNamespace(), grp, gvk.Version, gvk.Kind, obj.GetName())
}

func PodCondition(pod *v1.Pod, typ v1.PodConditionType) v1.PodCondition {
	if pod == nil {
		return v1.PodCondition{
			Type:   typ,
			Status: v1.ConditionUnknown,
		}
	}

	for _, cond := range pod.Status.Conditions {
		if cond.Type == typ {
			return cond
		}
	}

	return v1.PodCondition{
		Type:   typ,
		Status: v1.ConditionUnknown,
	}
}

func UpdateCondition(now metav1.Time, conds []metav1.Condition, cond *metav1.Condition) ([]metav1.Condition, bool) {
	for i, c := range conds {
		if c.Type == cond.Type {
			if c.Status != cond.Status {
				cond.LastTransitionTime = now
			} else {
				cond.LastTransitionTime = c.LastTransitionTime
			}
			conds[i] = *cond
			return conds, true
		}
	}

	cond.LastTransitionTime = now
	return append(conds, *cond), true
}

func Condition(conds []metav1.Condition, typ string) *metav1.Condition {
	for _, c := range conds {
		if c.Type == typ {
			return &c
		}
	}

	return &metav1.Condition{Type: typ, Status: metav1.ConditionUnknown}
}

func IsAvailable(conds []metav1.Condition) bool {
	return Condition(conds, api.ConditionTypeAvailable).Status == metav1.ConditionTrue
}
