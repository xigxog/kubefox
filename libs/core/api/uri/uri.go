package uri

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/url"
	"strings"

	"github.com/xigxog/kubefox/libs/core/utils"
)

const (
	// TODO replace this with org
	Authority     = "kubefox"
	KubeFoxScheme = "kubefox"
	PathSeparator = "/"
	HTTPSeparator = ":"
	HeadPath      = "head"
)

var (
	ErrInvalidURI = errors.New("invalid uri")
)

type URI interface {
	Authority() string
	Kind() Kind
	Name() string
	SubKind() SubKind
	SubPath() string
	SubPathWithKind() string
	Path() string
	HeadPath() string
	MetadataPath() string
	Key() Key
	HTTPKey() string
	URL() string

	String() string

	UnmarshalJSON([]byte) error
	MarshalJSON() ([]byte, error)
}

// +kubebuilder:validation:Type=string
// +kubebuilder:validation:Format=uri
type URIType struct {
	authority string  `json:"-" validate:"omitempty,objName"`
	kind      Kind    `json:"-" validate:"required"`
	name      string  `json:"-" validate:"omitempty,objName"`
	subKind   SubKind `json:"-"`
	subPath   string  `json:"-"`
}

// New builds a URI from the parts provided. The parts should be provided in the
// order: authority, kind, name, subKind, subPath. Only authority and kind are
// required. Each part is trimmed of leading or trailing '/' before parsing.
func New(authority string, kind Kind, parts ...any) (URI, error) {
	parts = append([]any{authority, kind}, parts...)
	trimmed := make([]string, len(parts))
	for i, p := range parts {
		trimmed[i] = strings.Trim(fmt.Sprintf("%s", p), "/")
	}

	return Parse(fmt.Sprintf("%s://%s", KubeFoxScheme, strings.Join(trimmed, PathSeparator)))
}

func Parse(val string) (URI, error) {
	u, err := url.Parse(val)
	if err != nil {
		return nil, fmt.Errorf("%w: %v", ErrInvalidURI, err)
	}

	if u.Scheme != KubeFoxScheme {
		return nil, fmt.Errorf("%w: invalid scheme '%s'", ErrInvalidURI, u.Scheme)
	}

	auth := u.Host
	if auth == "" {
		return nil, fmt.Errorf("%w: organization authority must be provided", ErrInvalidURI)
	}
	if !utils.NameRegexp.MatchString(auth) {
		return nil, fmt.Errorf("%w: invalid organization authority '%s'", ErrInvalidURI, auth)
	}

	pathParts := strings.Split(strings.Trim(u.EscapedPath(), PathSeparator), PathSeparator)
	numParts := len(pathParts)

	var resKind Kind
	var subresKind SubKind
	var resKindStr, resName, subresKindStr, subresPath string
	switch numParts {
	default:
		// /{resKind}/{resName}/{subresKind}/{subresPath...}
		subresPath = strings.Join(pathParts[3:], PathSeparator)
		fallthrough
	case 3:
		// /{resKind}/{resName}/{subresKind}
		subresKindStr = pathParts[2]
		subresKind = SubKindFromString(subresKindStr)
		fallthrough
	case 2:
		// /{resKind}/{resName}
		resName = pathParts[1]
		fallthrough
	case 1:
		// /{resKind}
		resKindStr = pathParts[0]
		resKind = KindFromString(resKindStr)
	}

	switch {
	case resKind == Unknown:
		return nil, fmt.Errorf("%w: invalid kind '%s'", ErrInvalidURI, resKindStr)

	case resName != "" && !utils.NameRegexp.MatchString(resName):
		return nil, fmt.Errorf("%w: invalid resource name '%s', should match regexp '%s'", ErrInvalidURI, resName, utils.NameRegexp.String())

	}

	if subresPath != "" {
		switch {
		case subresKind == None:
			return nil, fmt.Errorf("%w: invalid subresource kind '%s'", ErrInvalidURI, subresKindStr)

		case subresKind == Metadata:
			return nil, fmt.Errorf("%w: invalid subpath '%s' for metadata subresource, should be empty", ErrInvalidURI, subresPath)

		case (subresKind == Branch || subresKind == Tag) && !utils.TagOrBranchRegexp.MatchString(subresPath):
			return nil, fmt.Errorf("%w: invalid %s name '%s', should match regexp '%s'", ErrInvalidURI, subresKind, subresPath, utils.TagOrBranchRegexp.String())

		case resKind == System && subresKind == Id && !utils.HashRegexp.MatchString(subresPath):
			return nil, fmt.Errorf("%w: invalid system id '%s', should be hex encoded SHA-1 hash", ErrInvalidURI, subresPath)

		case resKind != System && subresKind == Id && !utils.UUIDRegexp.MatchString(subresPath):
			return nil, fmt.Errorf("%w: invalid %s id '%s', should be UUID type", ErrInvalidURI, subresKind, subresPath)

		case subresKind == Deployment && !(numParts == 4 || numParts == 6):
			return nil, fmt.Errorf("%w: invalid number of path parts '%d' for deployment subpath '%s'", ErrInvalidURI, numParts-3, subresPath)

		case subresKind == Release && !(numParts == 4 || numParts == 5):
			return nil, fmt.Errorf("%w: invalid number of path parts '%d' for release subpath '%s'", ErrInvalidURI, numParts-3, subresPath)
		}
	}

	return &URIType{
		authority: u.Host,
		kind:      resKind,
		name:      resName,
		subKind:   subresKind,
		subPath:   subresPath,
	}, nil
}

func (u *URIType) Authority() string {
	return u.authority
}

func (u *URIType) Kind() Kind {
	return u.kind
}

func (u *URIType) Name() string {
	return u.name
}

func (u *URIType) SubKind() SubKind {
	return u.subKind
}

func (u *URIType) SubPath() string {
	return u.subPath
}

func (u *URIType) SubPathWithKind() string {
	return strings.Join([]string{u.subKind.string(), u.subPath}, PathSeparator)
}

func (u *URIType) String() string {
	return u.Path()
}

func (u *URIType) URL() string {
	return fmt.Sprintf("%s://%s/%s", KubeFoxScheme, u.authority, u.Path())
}

func (u *URIType) Path() string {
	switch {
	case u.name == "":
		return strings.Join([]string{u.kind.string()}, PathSeparator)
	case u.subKind == None:
		return strings.Join([]string{u.kind.string(), u.name}, PathSeparator)
	case u.subPath == "":
		return strings.Join([]string{u.kind.string(), u.name, u.subKind.string()}, PathSeparator)
	default:
		return strings.Join([]string{u.kind.string(), u.name, u.subKind.string(), u.subPath}, PathSeparator)
	}
}

func (u *URIType) HeadPath() string {
	if u.name == "" {
		return ""
	}
	return strings.Join([]string{u.kind.string(), u.name, HeadPath}, PathSeparator)
}

func (u *URIType) Key() Key {
	return u.key(PathSeparator)
}

func (u *URIType) HTTPKey() string {
	return string(u.key(HTTPSeparator))
}

func (u *URIType) key(separator string) Key {
	if u.subKind != None {
		return Key(strings.Join([]string{u.name, u.subKind.string(), u.subPath}, separator))
	} else {
		return Key(u.name)
	}
}

func (u *URIType) MetadataPath() string {
	if u.name == "" {
		return ""
	}
	return strings.Join([]string{u.kind.string(), u.name, Metadata.String()}, PathSeparator)
}

func (u *URIType) UnmarshalJSON(data []byte) error {
	uriStr := ""
	if err := json.Unmarshal(data, &uriStr); err != nil {
		return err
	}

	parsed, err := Parse(uriStr)
	if err != nil {
		return err
	}
	*u = *parsed.(*URIType)

	return nil
}

func (u *URIType) MarshalJSON() ([]byte, error) {
	return json.Marshal(u.URL())
}
