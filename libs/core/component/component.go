package component

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/xigxog/kubefox/libs/core/api/uri"
	"github.com/xigxog/kubefox/libs/core/grpc"
)

var (
	IdRegExp      = regexp.MustCompile(`[a-f0-9]{5}`)
	GitHashRegExp = regexp.MustCompile(`[a-f0-9]{7}`)
)

type InvalidURI struct {
	Err string
}

func (e *InvalidURI) Error() string {
	return e.Err
}

type Component interface {
	GetApp() string
	SetApp(string)
	GetName() string
	SetName(string)
	GetGitHash() string
	SetGitHash(string)
	GetId() string
	SetId(string)

	GetURI() string
	GetKey() string
	GetHTTPKey() string

	GetRequestSubject() string
	GetResponseSubject() string
	GetAsyncSubject() string
	GetRequestConsumer() string
	GetResponseConsumer() string
	GetAsyncConsumer() string
	GetStream() string
	GetSubjectWildcard() string
}

type Fields struct {
	App     string
	Name    string
	GitHash string
	Id      string
}

func New(fields Fields) Component {
	return &grpc.Component{
		App:     fields.App,
		Name:    fields.Name,
		GitHash: fields.GitHash,
		Id:      fields.Id,
	}
}

func Copy(src Component) Component {
	return &grpc.Component{
		App:     src.GetApp(),
		Name:    src.GetName(),
		GitHash: src.GetGitHash(),
		Id:      src.GetId(),
	}
}

func Equal(lhs, rhs Component) bool {
	eq := lhs.GetName() == rhs.GetName() &&
		lhs.GetGitHash() == rhs.GetGitHash()

	if lhs.GetId() != "" && rhs.GetId() != "" {
		eq = eq && lhs.GetId() == rhs.GetId()
	}

	if lhs.GetApp() != "" && rhs.GetApp() != "" {
		eq = eq && lhs.GetApp() == rhs.GetApp()
	}

	return eq
}

func ParseURI(u string) (Component, error) {
	if u == "" {
		return nil, &InvalidURI{Err: "empty uri"}
	}
	uriParts := strings.Split(u, ":")

	if uriParts[0] != uri.KubeFoxScheme {
		return nil, &InvalidURI{Err: fmt.Sprintf("invalid scheme %s", uriParts[0])}
	}
	if uriParts[1] != "component" {
		return nil, &InvalidURI{Err: fmt.Sprintf("invalid path prefix %s", uriParts[1])}
	}

	var app, name, hash, id string
	switch len(uriParts) {
	case 4:
		app = uriParts[2]
		name = uriParts[3]
	case 5:
		app = uriParts[2]
		name = uriParts[3]
		hash = uriParts[4]
	case 6:
		app = uriParts[2]
		name = uriParts[3]
		hash = uriParts[4]
		id = uriParts[5]
	default:
		return nil, &InvalidURI{Err: "invalid number of path parts"}
	}

	if app == "" {
		return nil, &InvalidURI{Err: "invalid path, app missing"}
	}
	if name == "" {
		return nil, &InvalidURI{Err: "invalid path, name missing"}
	}

	// if (id != "" && hash == "") || (id == "" && hash != "") {
	// 	return nil, &InvalidURI{Err: "id and git hash must be provided together"}
	// }

	// if valid := IdRegExp.MatchString(id); id != "" && !valid {
	// 	return nil, &InvalidURI{Err: "invalid id"}
	// }
	// if valid := HashRegExp.MatchString(hash); hash != "" && !valid {
	// 	return nil, &InvalidURI{Err: "invalid hash"}
	// }

	return &grpc.Component{
		App:     app,
		Name:    name,
		GitHash: hash,
		Id:      id,
	}, nil
}
