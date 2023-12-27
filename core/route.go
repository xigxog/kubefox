package core

import (
	"regexp"

	"github.com/xigxog/kubefox/api"
)

var (
	RuleParamRegexp = regexp.MustCompile(`([^\\])({[^}]+})`)
)

type Route struct {
	*EnvTemplate

	Id           int
	ResolvedRule string
	Priority     int

	Component    *Component
	EventContext *EventContext
}

func NewRoute(id int, rule string) (*Route, error) {
	tpl, err := NewEnvTemplate(rule)
	if err != nil {
		return nil, err
	}

	return &Route{
		EnvTemplate: tpl,
		Id:          id,
	}, nil
}

func (r *Route) Resolve(data *api.VirtualEnvData) (err error) {
	if r.ResolvedRule, err = r.EnvTemplate.Resolve(data); err != nil {
		return
	}
	// Normalize path args so they don't affect length.
	r.Priority = len(RuleParamRegexp.ReplaceAllString(r.ResolvedRule, "$1{}"))

	return
}
