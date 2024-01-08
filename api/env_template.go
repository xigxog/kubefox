// Copyright 2023 XigXog
//
// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.
//
// SPDX-License-Identifier: MPL-2.0

package api

import (
	"fmt"
	"regexp"
	"strings"
	"text/template"
	"text/template/parse"
)

// +kubebuilder:object:generate=false
type EnvTemplate struct {
	template  string
	envSchema *EnvSchema
	tree      *parse.Tree
}

type templateData struct {
	Vars    map[string]*envVar
	Env     map[string]*envVar
	Secrets map[string]*envVar
}

// envVar is used to override String() method of Val so that it returns a regexp
// for arrays.
type envVar struct {
	Val *Val
}

func NewEnvTemplate(template string) (*EnvTemplate, error) {
	r := &EnvTemplate{
		template: template,
		envSchema: &EnvSchema{
			Vars:    make(EnvVarSchema),
			Secrets: make(EnvVarSchema),
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
					r.envSchema.Vars[name] = &EnvVarDefinition{
						Required: true,
					}
				case "Secrets":
					r.envSchema.Secrets[name] = &EnvVarDefinition{
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

func (r *EnvTemplate) Resolve(data *Data) (string, error) {
	if data == nil {
		data = &Data{}
	}

	envVarData := &templateData{
		Vars:    make(map[string]*envVar),
		Secrets: make(map[string]*envVar),
	}
	for k, v := range data.Vars {
		envVarData.Vars[k] = &envVar{Val: v}
	}
	for k, v := range data.ResolvedSecrets {
		envVarData.Secrets[k] = &envVar{Val: v}
	}
	// Supports use of {{.Env.<NAME>}} or {{.Vars.<NAME>}}.
	envVarData.Env = envVarData.Vars

	tpl := template.New("route").Option("missingkey=zero")
	tpl.Tree = r.tree

	var buf strings.Builder
	if err := tpl.Execute(&buf, envVarData); err != nil {
		return "", err
	}

	// Despite setting `missingkey=zero` above "<no value>" is still injected
	// for missing keys.
	return strings.ReplaceAll(buf.String(), "<no value>", ""), nil
}

func (e EnvVarSchema) Validate(vars map[string]*Val, src *ProblemSource) []Problem {
	var problems []Problem
	for varName, varDef := range e {
		val, found := vars[varName]

		if !found && varDef.Required {
			src := *src
			src.Path = fmt.Sprintf("%s.%s", src.Path, varName)
			problems = append(problems, Problem{
				Type:    ProblemTypeVarNotFound,
				Message: fmt.Sprintf(`Variable "%s" not found but is required.`, varName),
				Causes:  []ProblemSource{src},
			})
		}
		if found && varDef.Type != "" && val.EnvVarType() != varDef.Type {
			src := *src
			src.Path = fmt.Sprintf("%s.%s.type", src.Path, varName)
			src.Value = (*string)(&varDef.Type)
			problems = append(problems, Problem{
				Type: ProblemTypeVarWrongType,
				Message: fmt.Sprintf(`Variable "%s" has wrong type; wanted "%s" got "%s".`,
					varName, varDef.Type, val.EnvVarType()),
				Causes: []ProblemSource{src},
			})
		}
	}

	return problems
}

func (e *envVar) String() string {
	if e.Val.Type == ArrayNumber || e.Val.Type == ArrayString {
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
