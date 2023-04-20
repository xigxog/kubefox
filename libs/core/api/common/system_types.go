package common

type App struct {
	GitHashProp `json:",inline"`

	Title       string                   `json:"title,omitempty"`
	Description string                   `json:"description,omitempty"`
	Vars        *JSONSchema              `json:"vars,omitempty"`
	Components  map[string]*AppComponent `json:"components" validate:"dive"`
}

type JSONSchema struct {
	Properties map[string]*JSONSchemaProp `json:"properties,omitempty" validate:"dive"`
	Required   []string                   `json:"required,omitempty"`
}

type JSONSchemaProp struct {
	VarTypeProp `json:",inline"`
	Title       string `json:"title,omitempty"`
	Description string `json:"description,omitempty"`
}

type AppComponent struct {
	ComponentProps `json:",inline"`

	Routes []*Route `json:"routes,omitempty" validate:"dive"`
}
