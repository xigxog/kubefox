package admin

import (
	"github.com/xigxog/kubefox/libs/core/api/uri"
)

type Object interface {
	GetKind() uri.Kind
	SetKind(uri.Kind)
	GetAPIVersion() string
	SetAPIVersion(string)

	GetId() string
	SetId(string)
	GetName() string
	SetName(string)
	GetMetadata() *Metadata
	SetMetadata(*Metadata)
}

type SubObject interface {
	GetKind() uri.SubKind
	SetKind(uri.SubKind)
	GetAPIVersion() string
	SetAPIVersion(string)

	GetURI(string, string) (uri.URI, error)
}

type ObjectBase struct {
	Id string `json:"id,omitempty" validate:"required"`

	Kind       uri.Kind `json:"kind,omitempty" validate:"required,ne=0"`
	APIVersion string   `json:"apiVersion,omitempty" validate:"required"`
	Metadata   Metadata `json:"metadata,omitempty"`

	Message string `json:"message,omitempty"`
}

type SubObjectBase struct {
	Kind       uri.SubKind `json:"kind,omitempty" validate:"required,ne=0"`
	APIVersion string      `json:"apiVersion,omitempty"`

	Message string `json:"message,omitempty"`
}

type Metadata struct {
	Name        string `json:"name,omitempty"`
	Title       string `json:"title,omitempty"`
	Description string `json:"description,omitempty"`
}

func (obj *ObjectBase) GetKind() uri.Kind {
	return obj.Kind
}

func (obj *ObjectBase) SetKind(k uri.Kind) {
	obj.Kind = k
}

func (obj *ObjectBase) GetAPIVersion() string {
	return obj.APIVersion
}

func (obj *ObjectBase) SetAPIVersion(v string) {
	obj.APIVersion = v
}

func (obj *ObjectBase) GetId() string {
	return obj.Id
}

func (obj *ObjectBase) SetId(id string) {
	obj.Id = id
}

func (obj *ObjectBase) GetName() string {
	return obj.Metadata.Name
}

func (obj *ObjectBase) SetName(name string) {
	obj.Metadata.Name = name
}

func (obj *ObjectBase) GetMetadata() *Metadata {
	return &obj.Metadata
}

func (obj *ObjectBase) SetMetadata(meta *Metadata) {
	if meta == nil {
		meta = &Metadata{}
	}
	obj.Metadata = *meta
}

func (obj *SubObjectBase) GetKind() uri.SubKind {
	return obj.Kind
}

func (obj *SubObjectBase) SetKind(k uri.SubKind) {
	obj.Kind = k
}

func (obj *SubObjectBase) SetAPIVersion(v string) {
	obj.APIVersion = v
}

func (obj *SubObjectBase) GetAPIVersion() string {
	return obj.APIVersion
}
