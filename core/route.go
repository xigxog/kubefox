package core

import (
	"regexp"
	"strings"
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

	tpl *template.Template
}

type tplData struct {
	Env map[string]any
}

var (
	paramRegexp = regexp.MustCompile(`([^\\])({[^}]+})`)
)

func (r *Route) Resolve(envVars map[string]*api.Val, funcMap template.FuncMap) error {
	r.ResolvedRule = ""
	// removes any extra whitespace
	resolved := strings.Join(strings.Fields(r.Rule), " ")

	env := make(map[string]any, len(envVars))
	for k, v := range envVars {
		var a any
		if v.Type == api.ArrayNumber || v.Type == api.ArrayString {
			// Convert array to regex that matches any of the values.
			b := strings.Builder{}
			b.WriteString("{:")
			for _, s := range v.ArrayString() {
				b.WriteString("^")
				b.WriteString(regexp.QuoteMeta(s))
				b.WriteString("$|")
			}
			a = strings.TrimSuffix(b.String(), "|") + "}"

		} else {
			a = v.Any()
		}

		env[k] = a
	}

	// Resolve template fields from vars.
	if r.tpl == nil {
		tpl := template.New("route").Option("missingkey=zero")
		if funcMap != nil {
			tpl.Funcs(funcMap)
		}
		if _, err := tpl.Parse(resolved); err != nil {
			return err
		}
		r.tpl = tpl
	}
	var buf strings.Builder
	if err := r.tpl.Execute(&buf, &tplData{Env: env}); err != nil {
		return err
	}
	resolved = strings.ReplaceAll(buf.String(), "<no value>", "")

	// Normalize path args so they don't affect length.
	normalized := paramRegexp.ReplaceAllString(resolved, "$1{}")
	r.Priority = len(normalized)
	r.ResolvedRule = resolved

	return nil
}

func (r *Route) Match(evt *Event) (bool, error) {
	if r.ParseErr != nil {
		return false, r.ParseErr
	}

	return r.Predicate != nil && r.Predicate(evt), nil
}
