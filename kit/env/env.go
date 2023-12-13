package env

import "github.com/xigxog/kubefox/api"

type VarOption func(*api.VirtualEnvVarDefinition)

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
	Array = func(evs *api.VirtualEnvVarDefinition) {
		evs.Type = api.EnvVarTypeArray
	}

	Bool = func(evs *api.VirtualEnvVarDefinition) {
		evs.Type = api.EnvVarTypeBoolean
	}

	Number = func(evs *api.VirtualEnvVarDefinition) {
		evs.Type = api.EnvVarTypeNumber
	}

	String = func(evs *api.VirtualEnvVarDefinition) {
		evs.Type = api.EnvVarTypeString
	}

	Required = func(evs *api.VirtualEnvVarDefinition) {
		evs.Required = true
	}

	Unique = func(evs *api.VirtualEnvVarDefinition) {
		evs.Required = true
		evs.Unique = true
	}
)
