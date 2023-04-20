package v1alpha1

import (
	"github.com/xigxog/kubefox/libs/core/api/admin"
	"github.com/xigxog/kubefox/libs/core/api/common"
)

type Environment struct {
	admin.ObjectBase `json:",inline"`

	common.InheritsProp `json:",inline"`
	// TODO uncomment when adapters are added
	// common.ConfigProp   `json:",inline"`

	// Vars is a map of variables that can be used generally throughout the
	// KubeFox Platform.
	//
	// Nested objects are not supported.
	common.VarsProp `json:",inline"`

	App   *AppEnvironment   `json:"app,omitempty"`
	Route *RouteEnvironment `json:"route,omitempty"`

	Status EnvironmentStatus `json:"status,omitempty"`
}

type AppEnvironment struct {
	// Vars is a map of variables available to Components during runtime. They
	// must be explicitly requested in the Component's App definition.
	//
	// Nested objects are not supported.
	common.VarsProp `json:",inline"`
}

type RouteEnvironment struct {
	// Vars is a map of variables that can be used in Route templates. It is
	// important not to put secrets here as they may be accidentally leaked in
	// the routes.
	//
	// Nested objects are not supported.
	common.VarsProp `json:",inline"`
}

type EnvironmentStatus struct {
	Releases []*Release `json:"releases,omitempty" validate:"dive"`
}
