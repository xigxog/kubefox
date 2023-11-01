package core

import (
	"encoding/json"
	"fmt"
)

type SoSType int64

const (
	SoSTypeString SoSType = iota
	SoSTypeSecret
)

type StringOrSecret struct {
	Type      SoSType
	StringVal string
	SecretVal Secret
}

type Secret struct {
	SecretRef string `json:"secretRef"`
}

// UnmarshalJSON implements the json.Unmarshaller interface.
func (sos *StringOrSecret) UnmarshalJSON(value []byte) error {
	if value[0] == '"' {
		sos.Type = SoSTypeString
		return json.Unmarshal(value, &sos.StringVal)
	}
	sos.Type = SoSTypeSecret
	return json.Unmarshal(value, &sos.SecretVal)
}

// MarshalJSON implements the json.Marshaller interface.
func (sos StringOrSecret) MarshalJSON() ([]byte, error) {
	switch sos.Type {
	case SoSTypeString:
		return json.Marshal(sos.StringVal)
	case SoSTypeSecret:
		return json.Marshal(sos.SecretVal)
	default:
		return []byte{}, fmt.Errorf("impossible SoSType")
	}
}
