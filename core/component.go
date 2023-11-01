// +kubebuilder:object:generate=false
package core

import (
	"fmt"
	"os"
	"strings"

	"github.com/google/uuid"
	"github.com/xigxog/kubefox/utils"
)

// +kubebuilder:object:generate=true
type App struct {
	Name string `json:"name"`
	// +kubebuilder:validation:Format=uri
	ContainerRegistry string `json:"containerRegistry,omitempty"`
	Title             string `json:"title,omitempty"`
	Description       string `json:"description,omitempty"`
}

// +kubebuilder:object:generate=true
type ComponentSpec struct {
	ComponentTypeVar `json:",inline"`

	Title          string                       `json:"title,omitempty"`
	Description    string                       `json:"description,omitempty"`
	Routes         []RouteSpec                  `json:"routes,omitempty"`
	DefaultHandler bool                         `json:"defaultHandler,omitempty"`
	EnvSchema      map[string]*EnvVarSchema     `json:"envSchema,omitempty"`
	Dependencies   map[string]*ComponentTypeVar `json:"dependencies,omitempty"`
}

// +kubebuilder:object:generate=true
type EnvVarSchema struct {
	// +kubebuilder:validation:Enum=array;boolean;number;string
	Type        EnvVarType `json:"type"`
	Required    bool       `json:"required"`
	Unique      bool       `json:"unique"`
	Title       string     `json:"title,omitempty"`
	Description string     `json:"description,omitempty"`
}

// +kubebuilder:object:generate=true
type ComponentTypeVar struct {
	// +kubebuilder:validation:Enum=kubefox;http
	Type ComponentType `json:"type"`
}

func GenerateId() string {
	_, id := GenerateNameAndId()
	return id
}

func GenerateNameAndId() (string, string) {
	id := uuid.NewString()
	name := id
	if p, _ := os.LookupEnv(EnvPodName); p != "" {
		name = p
		s := strings.Split(p, "-")
		if len(s) > 1 {
			id = s[len(s)-1]
		}
	} else if h, _ := os.Hostname(); h != "" {
		name = h
	}

	return utils.CleanName(name), id
}

func (c *Component) IsFull() bool {
	return c.Name != "" && c.Commit != "" && c.Id != "" && c.BrokerId != ""
}

func (c *Component) IsNameOnly() bool {
	return c.Name != "" && c.Commit == "" && c.Id == "" && c.BrokerId == ""
}

func (lhs *Component) Equal(rhs *Component) bool {
	if rhs == nil {
		return false
	}
	return lhs.Name == rhs.Name &&
		lhs.Commit == rhs.Commit &&
		lhs.Id == rhs.Id &&
		lhs.BrokerId == rhs.BrokerId
}

func (c *Component) Key() string {
	return fmt.Sprintf("%s-%s-%s", c.Name, c.ShortCommit(), c.Id)
}

func (c *Component) GroupKey() string {
	return fmt.Sprintf("%s-%s", c.Name, c.ShortCommit())
}

func (c *Component) Subject() string {
	if c.BrokerId != "" {
		return c.BrokerSubject()
	}
	if c.Id == "" {
		return c.GroupSubject()
	}
	return fmt.Sprintf("evt.js.%s.%s.%s", c.Name, c.ShortCommit(), c.Id)
}

func (c *Component) GroupSubject() string {
	return fmt.Sprintf("evt.js.%s.%s", c.Name, c.ShortCommit())
}

func (c *Component) BrokerSubject() string {
	return fmt.Sprintf("evt.brk.%s", c.BrokerId)
}

func (c *Component) ShortCommit() string {
	if len(c.Commit) >= 7 {
		return c.Commit[0:7]
	}

	return ""
}
