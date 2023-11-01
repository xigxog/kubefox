package env

import kubefox "github.com/xigxog/kubefox/core"

type VarOption func(*kubefox.EnvVarSchema)

type Var struct {
	Name string
	Type kubefox.EnvVarType
}

func (v *Var) GetName() string {
	return v.Name
}

func (v *Var) GetType() kubefox.EnvVarType {
	return v.Type
}

func Array() VarOption {
	return func(evs *kubefox.EnvVarSchema) {
		evs.Type = kubefox.EnvVarTypeArray
	}
}

func Bool() VarOption {
	return func(evs *kubefox.EnvVarSchema) {
		evs.Type = kubefox.EnvVarTypeBoolean
	}
}

func Number() VarOption {
	return func(evs *kubefox.EnvVarSchema) {
		evs.Type = kubefox.EnvVarTypeNumber
	}
}

func String() VarOption {
	return func(evs *kubefox.EnvVarSchema) {
		evs.Type = kubefox.EnvVarTypeString
	}
}

func Required() VarOption {
	return func(evs *kubefox.EnvVarSchema) {
		evs.Required = true
	}
}

func Unique() VarOption {
	return func(evs *kubefox.EnvVarSchema) {
		evs.Unique = true
	}
}

func Title(title string) VarOption {
	return func(evs *kubefox.EnvVarSchema) {
		evs.Title = title
	}
}

func Description(description string) VarOption {
	return func(evs *kubefox.EnvVarSchema) {
		evs.Description = description
	}
}
