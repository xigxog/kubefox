package core

import (
	"fmt"
	"regexp"
	"strings"
	"sync"
	"text/template"
	"text/template/parse"

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

	tree *parse.Tree

	mutex sync.Mutex
}

type envVar struct {
	Val *api.Val
}

var (
	paramRegexp = regexp.MustCompile(`([^\\])({[^}]+})`)
)

func (r *Route) Match(evt *Event) (bool, error) {
	if r.ParseErr != nil {
		return false, r.ParseErr
	}

	return r.Predicate != nil && r.Predicate(evt), nil
}

func (r *Route) Resolve(envVars map[string]*api.Val) error {
	r.ResolvedRule = ""

	if err := r.buildTree(); err != nil {
		return err
	}

	envFuncs := make(template.FuncMap, len(envVars))
	for k, v := range envVars {
		// Create copies for use in func().
		name := k
		val := *v
		envFuncs[name] = func() *envVar {
			return &envVar{Val: &val}
		}
	}

	// Check if there is an env var named unique.
	curUnique, curUniqueFound := envFuncs["unique"].(func() *envVar)
	envFuncs["unique"] = func(v ...any) (*envVar, error) {
		// If unique is called with no args the env var will be returned, if it
		// exists, otherwise the unique function will be used.
		if len(v) == 0 && curUniqueFound {
			return curUnique(), nil
		}
		if len(v) != 1 {
			return nil, fmt.Errorf("wrong number of args for 'unique': want 1, got %d", len(v))
		}
		envVar, ok := v[0].(*envVar)
		if !ok {
			return nil, fmt.Errorf("wrong type of arg for 'unique': want 'EnvVar', got '%T'", v[0])
		}

		return envVar, nil
	}

	tpl := template.New("route").Option("missingkey=zero")
	tpl.Tree = r.tree
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

func (r *Route) BuildEnvSchema() error {
	if err := r.buildTree(); err != nil {
		return err
	}

	r.mutex.Lock()
	defer r.mutex.Unlock()

	if r.RouteSpec.EnvSchema != nil {
		return nil
	}

	envSchema := map[string]*api.EnvVarSchema{}
	for _, n := range r.tree.Root.Nodes {
		action, ok := n.(*parse.ActionNode)
		if !ok {
			continue
		}

		var (
			name   string
			unique bool
		)
		switch cmds := action.Pipe.Cmds; len(cmds) {
		case 1:
			name = cmds[0].String()
		case 2:
			if cmds[1].String() != "unique" {
				return fmt.Errorf("wrong modifier for EnvVar: want 'unique', got '%s'", cmds[1].String())
			}
			name = cmds[0].String()
			unique = true
		default:
			return fmt.Errorf("wrong number of commands: want 1 or 2, got %d", len(cmds))
		}

		envSchema[name] = &api.EnvVarSchema{
			Required: true,
			Unique:   unique,
		}
	}

	r.RouteSpec.EnvSchema = envSchema
	return nil
}

func (r *Route) buildTree() error {
	r.mutex.Lock()
	defer r.mutex.Unlock()

	if r.tree != nil {
		return nil
	}

	// removes any extra whitespace
	resolved := strings.Join(strings.Fields(r.Rule), " ")

	t := parse.New("route")
	t.Mode = t.Mode | parse.SkipFuncCheck
	if _, err := t.Parse(resolved, "{{", "}}", map[string]*parse.Tree{}); err != nil {
		return err
	}

	r.tree = t
	return nil
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
