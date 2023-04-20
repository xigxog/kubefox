// +kubebuilder:object:generate=true
package common

import metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

type Route struct {
	RouteTypeProp       `json:",inline"`
	Priority            int                      `json:"priority,omitempty"`
	Match               string                   `json:"match,omitempty"`
	Schedule            string                   `json:"schedule,omitempty"`
	CompositeController *CompositeControllerSpec `json:"compositeController,omitempty"`
	DecoratorController *DecoratorControllerSpec `json:"decoratorController,omitempty"`
}

type CompositeControllerSpec struct {
	ParentResource CompositeControllerParentResourceRule  `json:"parentResource"`
	ChildResources []CompositeControllerChildResourceRule `json:"childResources,omitempty" validate:"dive"`

	ResyncPeriodSeconds *int32 `json:"resyncPeriodSeconds,omitempty"`
	GenerateSelector    *bool  `json:"generateSelector,omitempty"`
}

type ResourceRule struct {
	APIVersion string `json:"apiVersion"`
	Resource   string `json:"resource"`
}

type CompositeControllerParentResourceRule struct {
	ResourceRule    `json:",inline"`
	RevisionHistory *CompositeControllerRevisionHistory `json:"revisionHistory,omitempty"`
}

type CompositeControllerRevisionHistory struct {
	FieldPaths []string `json:"fieldPaths,omitempty"`
}

type ChildUpdateMethod string

const (
	ChildUpdateOnDelete        ChildUpdateMethod = "OnDelete"
	ChildUpdateRecreate        ChildUpdateMethod = "Recreate"
	ChildUpdateInPlace         ChildUpdateMethod = "InPlace"
	ChildUpdateRollingRecreate ChildUpdateMethod = "RollingRecreate"
	ChildUpdateRollingInPlace  ChildUpdateMethod = "RollingInPlace"
)

type CompositeControllerChildResourceRule struct {
	ResourceRule   `json:",inline"`
	UpdateStrategy *CompositeControllerChildUpdateStrategy `json:"updateStrategy,omitempty"`
}

type CompositeControllerChildUpdateStrategy struct {
	Method       ChildUpdateMethod       `json:"method,omitempty"`
	StatusChecks ChildUpdateStatusChecks `json:"statusChecks,omitempty"`
}

type ChildUpdateStatusChecks struct {
	Conditions []StatusConditionCheck `json:"conditions,omitempty" validate:"dive"`
}

type StatusConditionCheck struct {
	Type   string  `json:"type"`
	Status *string `json:"status,omitempty"`
	Reason *string `json:"reason,omitempty"`
}

type ControllerRevisionChildren struct {
	APIGroup string   `json:"apiGroup"`
	Kind     string   `json:"kind"`
	Names    []string `json:"names"`
}

type DecoratorControllerSpec struct {
	Resources   []DecoratorControllerResourceRule   `json:"resources" validate:"dive"`
	Attachments []DecoratorControllerAttachmentRule `json:"attachments,omitempty" validate:"dive"`

	ResyncPeriodSeconds *int32 `json:"resyncPeriodSeconds,omitempty"`
}

type DecoratorControllerResourceRule struct {
	ResourceRule       `json:",inline"`
	LabelSelector      *metav1.LabelSelector `json:"labelSelector,omitempty"`
	AnnotationSelector *AnnotationSelector   `json:"annotationSelector,omitempty"`
}

type AnnotationSelector struct {
	MatchAnnotations map[string]string                 `json:"matchAnnotations,omitempty"`
	MatchExpressions []metav1.LabelSelectorRequirement `json:"matchExpressions,omitempty" validate:"dive"`
}

type DecoratorControllerAttachmentRule struct {
	ResourceRule   `json:",inline"`
	UpdateStrategy *DecoratorControllerAttachmentUpdateStrategy `json:"updateStrategy,omitempty"`
}

type DecoratorControllerAttachmentUpdateStrategy struct {
	Method ChildUpdateMethod `json:"method,omitempty"`
}

type RelatedResourceRule struct {
	ResourceRule          `json:",inline"`
	*metav1.LabelSelector `json:"labelSelector"`
	Namespace             string   `json:"namespace,omitempty"`
	Names                 []string `json:"names"`
}
