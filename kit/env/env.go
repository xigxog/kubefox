package env

import (
	common "github.com/xigxog/kubefox/api/kubernetes"
)

type VarOption func(*common.EnvVarSchema)

type Var struct {
	name string
	typ  common.EnvVarType
}

func NewVar(name string, typ common.EnvVarType) *Var {
	return &Var{name: name, typ: typ}
}

func (v *Var) Name() string {
	return v.name
}

func (v *Var) Type() common.EnvVarType {
	return v.typ
}

func Array() VarOption {
	return func(evs *common.EnvVarSchema) {
		evs.Type = common.EnvVarTypeArray
	}
}

func Bool() VarOption {
	return func(evs *common.EnvVarSchema) {
		evs.Type = common.EnvVarTypeBoolean
	}
}

func Number() VarOption {
	return func(evs *common.EnvVarSchema) {
		evs.Type = common.EnvVarTypeNumber
	}
}

func String() VarOption {
	return func(evs *common.EnvVarSchema) {
		evs.Type = common.EnvVarTypeString
	}
}

func Required() VarOption {
	return func(evs *common.EnvVarSchema) {
		evs.Required = true
	}
}

func Unique() VarOption {
	return func(evs *common.EnvVarSchema) {
		evs.Unique = true
	}
}
