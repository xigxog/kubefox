package uri

import (
	"encoding/json"
	"fmt"
	"strings"
)

type Key string
type Kind int
type SubKind int

const (
	Unknown Kind = iota
	Config
	Environment
	System
	Platform
)

const (
	None SubKind = iota
	Metadata
	Id
	Tag
	Branch
	Deployment
	Release
)

func (r Kind) String() string {
	switch r {
	case Config:
		return "Config"
	case Environment:
		return "Environment"
	case System:
		return "System"
	case Platform:
		return "Platform"
	default:
		return ""
	}
}

func (r Kind) string() string {
	return strings.ToLower(r.String())
}

func KindFromString(str string) Kind {
	switch strings.ToLower(str) {
	case "config", "configs", "conf":
		return Config
	case "environment", "environments", "env":
		return Environment
	case "system", "systems", "sys":
		return System
	case "platform", "platforms", "plat":
		return Platform
	default:
		return Unknown
	}
}

func (r SubKind) String() string {
	switch r {
	case Metadata:
		return "Metadata"
	case Id:
		return "Id"
	case Tag:
		return "Tag"
	case Branch:
		return "Branch"
	case Deployment:
		return "Deployment"
	case Release:
		return "Release"
	default:
		return ""
	}
}

func (r SubKind) string() string {
	return strings.ToLower(r.String())
}

func SubKindFromString(str string) SubKind {
	switch strings.ToLower(str) {
	case "metadata":
		return Metadata
	case "id", "ids":
		return Id
	case "tag", "tags":
		return Tag
	case "branch", "branches":
		return Branch
	case "deployment", "deployments":
		return Deployment
	case "release", "releases":
		return Release
	default:
		return None
	}
}

func (k *Kind) UnmarshalJSON(data []byte) error {
	kStr := ""
	if err := json.Unmarshal(data, &kStr); err != nil {
		return err
	}

	kind := KindFromString(kStr)
	if kind == Unknown {
		return fmt.Errorf("unknown resource kind %s", kStr)
	}
	*k = kind

	return nil
}

func (k *Kind) MarshalJSON() ([]byte, error) {
	return json.Marshal(k.String())
}

func (k *SubKind) UnmarshalJSON(data []byte) error {
	kStr := ""
	if err := json.Unmarshal(data, &kStr); err != nil {
		return err
	}

	kind := SubKindFromString(kStr)
	if kind == None {
		return fmt.Errorf("unknown subresource kind %s", kStr)
	}
	*k = kind

	return nil
}

func (k *SubKind) MarshalJSON() ([]byte, error) {
	return json.Marshal(k.String())
}

func (l Key) Equals(r Key) bool {
	return l == r
}
