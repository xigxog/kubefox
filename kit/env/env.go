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

func Array() VarOption {
	return func(evs *api.EnvVarSchema) {
		evs.Type = api.EnvVarTypeArray
	}
}

func Bool() VarOption {
	return func(evs *api.EnvVarSchema) {
		evs.Type = api.EnvVarTypeBoolean
	}
}

func Number() VarOption {
	return func(evs *api.EnvVarSchema) {
		evs.Type = api.EnvVarTypeNumber
	}
}

func String() VarOption {
	return func(evs *api.EnvVarSchema) {
		evs.Type = api.EnvVarTypeString
	}
}

func Required() VarOption {
	return func(evs *api.EnvVarSchema) {
		evs.Required = true
	}
}

func Unique() VarOption {
	return func(evs *api.EnvVarSchema) {
		evs.Unique = true
	}
}
