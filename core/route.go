package core

import (
	"fmt"
	"regexp"
	"strings"
	"sync"
	"text/template"

	"github.com/xigxog/kubefox/api"
)

// Takes an Event and returns true or false if rule matches.
type EventPredicate func(e *Event) bool

type Route struct {
	api.RouteSpec

	ResolvedRule string
	Predicate    EventPredicate
	ParseErr     error

	Component    *Component
	EventContext *EventContext

	EnvSchema map[string]*api.EnvVarSchema

	tpl   *template.Template
	mutex sync.Mutex
}

type envVar struct {
	Name   string
	Val    *api.Val
	Schema *api.EnvVarSchema
}

var (
	paramRegexp = regexp.MustCompile(`([^\\])({[^}]+})`)
)

func (r *Route) Resolve(envVars map[string]*api.Val) error {
	r.mutex.Lock()
	defer r.mutex.Unlock()

	r.ResolvedRule = ""

	tpl, err := r.getTpl(envVars)
	if err != nil {
		return err
	}

	r.EnvSchema = make(map[string]*api.EnvVarSchema)
	envFuncs := make(template.FuncMap, len(envVars))
	for k, v := range envVars {
		name := k
		val := *v
		envFuncs[name] = func() *envVar {
			if _, found := r.EnvSchema[name]; !found {
				r.EnvSchema[name] = &api.EnvVarSchema{
					Type:     val.EnvVarType(),
					Required: true,
				}
			}
			return &envVar{
				Name: name,
				Val:  &val,
			}
		}
	}

	curUnique, curUniqueFound := envFuncs["unique"].(func() *envVar)
	envFuncs["unique"] = func(v ...any) (any, error) {
		if len(v) == 0 && curUniqueFound {
			return curUnique(), nil
		}
		if len(v) != 1 {
			return nil, fmt.Errorf("wrong number of args for unique: want 1 got %d", len(v))
		}
		envVar, ok := v[0].(*envVar)
		if !ok {
			return nil, fmt.Errorf("wrong type of arg for unique: want EnvVar got %T", v[0])
		}

		if s, found := r.EnvSchema[envVar.Name]; !found {
			r.EnvSchema[envVar.Name] = &api.EnvVarSchema{
				Type:     envVar.Val.EnvVarType(),
				Required: true,
				Unique:   true,
			}
		} else {
			s.Unique = true
		}

		return envVar, nil
	}

	tpl.Funcs(envFuncs)

	var buf strings.Builder
	if err := tpl.Execute(&buf, map[string]any{}); err != nil {
		return err
	}

	r.ResolvedRule = strings.ReplaceAll(buf.String(), "<no value>", "")
	// Normalize path args so they don't affect length.
	r.Priority = len(paramRegexp.ReplaceAllString(r.ResolvedRule, "$1{}"))

	return nil
}

func (r *Route) getTpl(envVars map[string]*api.Val) (*template.Template, error) {
	if r.tpl != nil {
		// We've already parsed the template.
		return r.tpl, nil
	}

	// removes any extra whitespace
	resolved := strings.Join(strings.Fields(r.Rule), " ")

	tpl := template.New("route").Option("missingkey=zero")
	if _, err := tpl.Parse(resolved); err != nil {
		return nil, err
	}
	r.tpl = tpl

	return r.tpl, nil
}

func (r *Route) Match(evt *Event) (bool, error) {
	if r.ParseErr != nil {
		return false, r.ParseErr
	}

	return r.Predicate != nil && r.Predicate(evt), nil
}

func (e *envVar) String() string {
	if e.Val.Type == api.ArrayNumber || e.Val.Type == api.ArrayString {
		// Convert array to regex that matches any of the values.
		b := strings.Builder{}
		b.WriteString("{")
		for _, s := range e.Val.ArrayString() {
			b.WriteString("^")
			b.WriteString(regexp.QuoteMeta(s))
			b.WriteString("$|")
		}
		return strings.TrimSuffix(b.String(), "|") + "}"
	}

	return e.Val.String()
}
