package core

import (
	"regexp"
	"strings"
	"text/template"
	"text/template/parse"

	"github.com/xigxog/kubefox/api"
)

type EnvTemplate struct {
	template  string
	envSchema *EnvSchema
	tree      *parse.Tree
}

type EnvSchema struct {
	Vars    map[string]*api.EnvVarDefinition
	Secrets map[string]*api.EnvVarDefinition
}

type envVarData struct {
	Vars    map[string]*envVar
	Env     map[string]*envVar
	Secrets map[string]*envVar
}

type envVar struct {
	Val *api.Val
}

func NewEnvTemplate(template string) (*EnvTemplate, error) {
	r := &EnvTemplate{
		template: template,
		envSchema: &EnvSchema{
			Vars:    make(map[string]*api.EnvVarDefinition),
			Secrets: make(map[string]*api.EnvVarDefinition),
		},
	}

	// removes any extra whitespace
	resolved := strings.Join(strings.Fields(r.template), " ")

	r.tree = parse.New("route")
	if _, err := r.tree.Parse(resolved, "{{", "}}", map[string]*parse.Tree{}); err != nil {
		return nil, err
	}

	for _, n := range r.tree.Root.Nodes {
		action, ok := n.(*parse.ActionNode)
		if !ok {
			continue
		}
		if action.Pipe == nil {
			continue
		}

		for _, cmd := range action.Pipe.Cmds {
			for _, arg := range cmd.Args {
				field, ok := arg.(*parse.FieldNode)
				if !ok {
					continue
				}
				if len(field.Ident) != 2 {
					continue
				}

				section, name := field.Ident[0], field.Ident[1]
				switch section {
				case "Vars", "Env":
					r.envSchema.Vars[name] = &api.EnvVarDefinition{
						Required: true,
					}
				case "Secrets":
					r.envSchema.Secrets[name] = &api.EnvVarDefinition{
						Required: true,
					}
				}
			}
		}
	}

	return r, nil
}

func (r *EnvTemplate) Template() string {
	return r.template
}

func (r *EnvTemplate) EnvSchema() *EnvSchema {
	return r.envSchema
}

func (r *EnvTemplate) Resolve(data *api.VirtualEnvData) (string, error) {
	envVarData := &envVarData{
		Vars:    make(map[string]*envVar),
		Secrets: make(map[string]*envVar),
	}
	for k, v := range data.Vars {
		envVarData.Vars[k] = &envVar{Val: v}
	}
	for k, v := range data.ResolvedSecrets {
		envVarData.Secrets[k] = &envVar{Val: v}
	}
	envVarData.Env = envVarData.Vars

	tpl := template.New("route").Option("missingkey=zero")
	tpl.Tree = r.tree

	var buf strings.Builder
	if err := tpl.Execute(&buf, envVarData); err != nil {
		return "", err
	}

	return strings.ReplaceAll(buf.String(), "<no value>", ""), nil
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
