package core

import (
	"regexp"
	"strings"
	"text/template"

	sprig "github.com/go-task/slim-sprig"
)

// Takes an Event and returns true or false if rule matches.
type EventPredicate func(e *Event) bool

type Route struct {
	Id       int    `json:"id"`
	Rule     string `json:"rule"`
	Priority int    `json:"priority"`

	ResolvedRule string         `json:"-"`
	Predicate    EventPredicate `json:"-"`
	ParseErr     error          `json:"-"`

	Component    *Component    `json:"-"`
	EventContext *EventContext `json:"-"`

	tpl *template.Template
}

type tplData struct {
	Env map[string]any
}

var (
	paramRegexp = regexp.MustCompile(`([^\\])({[^}]+})`)
)

func (r *Route) Resolve(envVars map[string]*Val) error {
	r.ResolvedRule = ""
	// removes any extra whitespace
	resolved := strings.Join(strings.Fields(r.Rule), " ")

	env := make(map[string]any, len(envVars))
	for k, v := range envVars {
		var a any
		if v.Type == ArrayNumber || v.Type == ArrayString {
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
		tpl.Funcs(sprig.FuncMap())
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
