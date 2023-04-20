package kubefox

import (
	"strings"

	"github.com/vulcand/predicate"
)

type EventMatcher struct {
	rule string
	eval eventPredicate
}

func NewEventMatcher(rule string) (*EventMatcher, error) {
	// Create a new parser and define the supported operators and methods
	p, err := predicate.NewParser(predicate.Def{
		Functions: map[string]interface{}{
			"Hook":       hook,
			"Method":     method,
			"Path":       path,
			"PathPrefix": pathPrefix,
			"TargetKind": targetKind,
			"Type":       eventType,
		},
		Operators: predicate.Operators{
			AND: and,
			OR:  or,
			NOT: not,
		},
	})
	if err != nil {
		return nil, err
	}

	// removes any extra whitespace
	rule = strings.Join(strings.Fields(rule), " ")
	eval, err := p.Parse(rule)
	if err != nil {
		return nil, err
	}

	return &EventMatcher{
		rule: rule,
		eval: eval.(eventPredicate),
	}, nil
}

func (m *EventMatcher) Rule() string {
	return m.rule
}

func (m *EventMatcher) Match(e Event) bool {
	return m.eval(e)
}

// takes an Event and returns true or false if rule matches
type eventPredicate func(e Event) bool

func hook(s string) eventPredicate {
	return func(e Event) bool {
		return strings.EqualFold(e.Kube().GetHook().String(), s)
	}
}

func method(s string) eventPredicate {
	return func(e Event) bool {
		return strings.EqualFold(e.HTTP().GetMethod(), s)
	}
}

func path(s string) eventPredicate {
	return func(e Event) bool {
		return matchPath(s, e, false)
	}
}

func pathPrefix(s string) eventPredicate {
	return func(e Event) bool {
		return matchPath(s, e, true)
	}
}

func matchPath(path string, e Event, prefix bool) bool {
	h := e.HTTP()

	tPath := strings.Trim(path, "/")
	tEvtPath := strings.Trim(h.GetPath(), "/")
	parts := strings.Split(tPath, "/")
	eventParts := strings.Split(tEvtPath, "/")

	if len(parts) > len(eventParts) {
		return false
	}
	if !prefix && len(parts) != len(eventParts) {
		return false
	}

	tmpArgs := make(map[string]string)
	for i, v := range parts {
		eventV := eventParts[i]
		if strings.HasPrefix(v, "{") && strings.HasSuffix(v, "}") {
			k := v[1 : len(v)-1]
			if k != "" {
				tmpArgs[k] = eventV
			}
			continue
		}

		if v != eventV {
			return false
		}
	}

	for k, v := range tmpArgs {
		h.SetArg(k, v)
	}

	return true
}

func eventType(s string) eventPredicate {
	return func(e Event) bool {
		return e.GetType() == s
	}
}

func targetKind(s string) eventPredicate {
	return func(e Event) bool {
		return e.Kube().GetTargetKind() == s
	}
}

// Logical operator AND that combines predicates
func and(a, b eventPredicate) eventPredicate {
	return func(e Event) bool {
		return a(e) && b(e)
	}
}

// Logical operator OR that combines predicates
func or(a, b eventPredicate) eventPredicate {
	return func(e Event) bool {
		return a(e) || b(e)
	}
}

// Logical operator NOT that negates predicates
func not(a eventPredicate) eventPredicate {
	return func(e Event) bool {
		return !a(e)
	}
}
