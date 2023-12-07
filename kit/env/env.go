package env

import "github.com/xigxog/kubefox/api"

type VarOption func(*api.EnvVarSchema)

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
	Array = func(evs *api.EnvVarSchema) {
		evs.Type = api.EnvVarTypeArray
	}

	Bool = func(evs *api.EnvVarSchema) {
		evs.Type = api.EnvVarTypeBoolean
	}

	Number = func(evs *api.EnvVarSchema) {
		evs.Type = api.EnvVarTypeNumber
	}

	String = func(evs *api.EnvVarSchema) {
		evs.Type = api.EnvVarTypeString
	}

	Required = func(evs *api.EnvVarSchema) {
		evs.Required = true
	}

	Unique = func(evs *api.EnvVarSchema) {
		evs.Required = true
		evs.Unique = true
	}
)
