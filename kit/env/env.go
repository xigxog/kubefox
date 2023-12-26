package env

import "github.com/xigxog/kubefox/api"

type VarOption func(*api.EnvVarDefinition)

type Var struct {
	name string
	typ  api.EnvVarType
}

func NewVar(name string, typ api.EnvVarType) *Var {
	return &Var{name: name, typ: typ}
}

func (v *Var) Name() string {
	return v.name
}

func (v *Var) Type() api.EnvVarType {
	return v.typ
}

var (
	Array = func(evs *api.EnvVarDefinition) {
		evs.Type = api.EnvVarTypeArray
	}

	Bool = func(evs *api.EnvVarDefinition) {
		evs.Type = api.EnvVarTypeBoolean
	}

	Number = func(evs *api.EnvVarDefinition) {
		evs.Type = api.EnvVarTypeNumber
	}

	String = func(evs *api.EnvVarDefinition) {
		evs.Type = api.EnvVarTypeString
	}

	Required = func(evs *api.EnvVarDefinition) {
		evs.Required = true
	}
)
