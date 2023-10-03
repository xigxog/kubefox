package v1alpha1

type EnvSchema struct {
	Vars     map[string]*EnvVarSchema `json:"vars,omitempty"`
	Required []string                 `json:"required,omitempty"`
}

type EnvVarSchema struct {
	// +kubebuilder:validation:Enum=array;boolean;number;string
	Type        string `json:"type"`
	Title       string `json:"title,omitempty"`
	Description string `json:"description,omitempty"`
}
