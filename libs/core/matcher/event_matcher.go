package matcher

import (
	"fmt"
	"net/textproto"
	"regexp"
	"sort"
	"strings"
	"sync"

	"github.com/vulcand/predicate"
	"github.com/xigxog/kubefox/libs/core/kubefox"
)

type param struct {
	name  string
	regex *regexp.Regexp
}

type EventMatcher struct {
	routes []*kubefox.Route
	parser predicate.Parser
	mutex  sync.RWMutex
}

func New() *EventMatcher {
	m := &EventMatcher{
		routes: make([]*kubefox.Route, 0),
	}

	// Create a new parser and define the supported operators and methods
	m.parser, _ = predicate.NewParser(predicate.Def{
		Functions: map[string]interface{}{
			"All":        m.all,
			"Header":     m.header,
			"Host":       m.host,
			"Method":     m.method,
			"Path":       m.path,
			"PathPrefix": m.pathPrefix,
			"Query":      m.query,
			"Type":       m.eventType,
		},
		Operators: predicate.Operators{
			AND: and,
			OR:  or,
			NOT: not,
		},
	})

	return m
}

func (m *EventMatcher) IsEmpty() bool {
	return len(m.routes) == 0
}

func (m *EventMatcher) AddRoutes(routes []*kubefox.Route) error {
	for _, r := range routes {
		if r.ResolvedRule == "" {
			return fmt.Errorf("route %d has unresolved rule", r.Id)
		}
		if r.Component == nil || r.EventContext == nil {
			return fmt.Errorf("route %d is missing context", r.Id)
		}

		parsed, err := m.parser.Parse(r.ResolvedRule)
		if err != nil {
			r.ParseErr = fmt.Errorf("invalid route '%s': parsing '%s' failed; %w", r.Rule, r.ResolvedRule, err)
			continue
		}
		r.Predicate = parsed.(kubefox.EventPredicate)

		m.mutex.Lock()
		m.routes = append(m.routes, r)
		m.mutex.Unlock()
	}

	// Sort rules, longest (most specific) rule should be tested first.
	sort.SliceStable(m.routes, func(i, j int) bool {
		return m.routes[i].Priority > m.routes[j].Priority
	})

	return nil
}

func (m *EventMatcher) Match(evt *kubefox.Event) (*kubefox.Route, bool) {
	m.mutex.RLock()
	defer m.mutex.RUnlock()

	for _, r := range m.routes {
		if matched, _ := r.Match(evt); matched {
			if evt.Target != nil {
				// Ensure matched route belongs to event target.
				if evt.Target.Name != "" && evt.Target.Name != r.Component.Name {
					continue
				}
				if evt.Target.Commit != "" && evt.Target.Commit != r.Component.Commit {
					continue
				}
			}

			return r, true
		}
	}

	return nil, false
}

func (m *EventMatcher) all() kubefox.EventPredicate {
	return func(e *kubefox.Event) bool {
		return true
	}
}

func (m *EventMatcher) header(key, val string) (kubefox.EventPredicate, error) {
	if key == "" {
		return nil, fmt.Errorf("header key must be provided")
	}
	key = textproto.CanonicalMIMEHeaderKey(key)

	regex, err := extractRegex(val)
	if err != nil {
		return nil, fmt.Errorf("invalid regex of header predicate %s: %w", val, err)
	}

	return func(e *kubefox.Event) bool {
		return matchMap(key, val, regex, e.ValueMap(kubefox.ValKeyHeader))
	}, nil
}

func (m *EventMatcher) host(s string) (kubefox.EventPredicate, error) {
	parts, params, err := split(s, '.')
	if err != nil {
		return nil, err
	}

	return func(e *kubefox.Event) bool {
		return matchParts(kubefox.ValKeyHost, ".", parts, params, e, false)
	}, nil
}

func (m *EventMatcher) method(s ...string) kubefox.EventPredicate {
	return func(e *kubefox.Event) bool {
		m := e.Value(kubefox.ValKeyMethod)
		for _, v := range s {
			if strings.EqualFold(m, v) {
				return true
			}
		}
		return false
	}
}

func (m *EventMatcher) path(s string) (kubefox.EventPredicate, error) {
	parts, params, err := split(s, '/')
	if err != nil {
		return nil, err
	}

	return func(e *kubefox.Event) bool {
		return matchParts(kubefox.ValKeyPath, "/", parts, params, e, false)
	}, nil
}

func (m *EventMatcher) pathPrefix(s string) (kubefox.EventPredicate, error) {
	parts, params, err := split(s, '/')
	if err != nil {
		return nil, err
	}

	return func(e *kubefox.Event) bool {
		return matchParts(kubefox.ValKeyPath, "/", parts, params, e, true)
	}, nil
}

func (m *EventMatcher) query(key, val string) (kubefox.EventPredicate, error) {
	if key == "" {
		return nil, fmt.Errorf("query param key must be provided")
	}

	regex, err := extractRegex(val)
	if err != nil {
		return nil, fmt.Errorf("invalid regex of query predicate %s: %w", val, err)
	}

	return func(e *kubefox.Event) bool {
		return matchMap(key, val, regex, e.ValueMap(kubefox.ValKeyQuery))
	}, nil
}

func (m *EventMatcher) eventType(s string) kubefox.EventPredicate {
	return func(e *kubefox.Event) bool {
		return e.GetType() == s
	}
}

// Logical operator AND that combines predicates
func and(a, b kubefox.EventPredicate) kubefox.EventPredicate {
	return func(e *kubefox.Event) bool {
		return a(e) && b(e)
	}
}

// Logical operator OR that combines predicates
func or(a, b kubefox.EventPredicate) kubefox.EventPredicate {
	return func(e *kubefox.Event) bool {
		return a(e) || b(e)
	}
}

// Logical operator NOT that negates predicates
func not(a kubefox.EventPredicate) kubefox.EventPredicate {
	return func(e *kubefox.Event) bool {
		return !a(e)
	}
}

func extractRegex(val string) (regex *regexp.Regexp, err error) {
	if strings.HasPrefix(val, "{") && strings.HasSuffix(val, "}") {
		r := val[1 : len(val)-1]
		if r == "" {
			r = ".*"
		} else {
			r = strings.TrimSuffix(strings.TrimPrefix(r, "^"), "$")
		}
		regex, err = regexp.Compile("^" + r + "$")
		if err != nil {
			return
		}
	}

	return
}

func matchMap(key, val string, regex *regexp.Regexp, m map[string][]string) bool {
	if valArr, found := m[key]; found {
		for _, v := range valArr {
			if regex != nil && regex.MatchString(v) {
				return true
			} else if v == val {
				return true
			}
		}
	}
	return false
}

func matchParts(val string, sep string, parts []string, params map[int]*param, e *kubefox.Event, prefix bool) bool {
	evtParts := strings.Split(strings.Trim(e.Value(val), sep), sep)

	if len(parts) > len(evtParts) {
		return false
	}
	if !prefix && len(parts) != len(evtParts) {
		return false
	}

	tmpParams := make(map[string]string)
	for i, v := range parts {
		evtV := evtParts[i]

		if p, found := params[i]; found {
			if !p.regex.MatchString(evtV) {
				return false
			}

			if p.name != "" {
				tmpParams[p.name] = evtV
			}

		} else if v != evtV {
			return false
		}
	}

	for k, v := range tmpParams {
		e.SetParam(k, v)
	}

	return true
}

// If this is ever placed in a hot path it should be optimized to use slices
// instead of copying strings as it currently does for clarity.
func split(s string, sep byte) ([]string, map[int]*param, error) {
	parts := make([]string, 0)
	params := make(map[int]*param)

	for i := 0; i < len(s); {
		switch s[i] {
		case sep:
			i++

		case '{':
			start := i
			b := strings.Builder{}
			var c byte
			for i < len(s) {
				c = s[i]
				var l byte
				if i > 0 {
					l = s[i-1]
				}
				b.WriteByte(c)
				i++
				if c == '}' && l != '\\' {
					break
				}
			}
			if c != '}' {
				return nil, nil, fmt.Errorf("unclosed bracket started at index %d of predicate %s", start, s)
			}
			regexParts := strings.Split(b.String()[1:len(b.String())-1], ":")

			var n, r string
			if len(regexParts) == 1 {
				n = regexParts[0]
			} else {
				n = regexParts[0]
				r = strings.Join(regexParts[1:], "")
			}
			if r == "" {
				r = fmt.Sprintf("[^%s]+", string(sep))
			}
			rt := strings.TrimSuffix(strings.TrimPrefix(r, "^"), "$")
			re, err := regexp.Compile("^" + rt + "$")
			if err != nil {
				return nil, nil, fmt.Errorf("invalid regex at index %d of predicate %s: %w", start+len(n)+1, s, err)
			}

			parts = append(parts, b.String())
			params[len(parts)-1] = &param{name: n, regex: re}

		default:
			b := strings.Builder{}
			for i < len(s) {
				c := s[i]
				if c == sep {
					break
				}
				if c == '{' && s[i-1] != '\\' {
					return nil, nil, fmt.Errorf("found mix of literal and regex in same part at index %d of predicate %s", i, s)
				}
				if !(c == '\\' && i < len(s)-1 && s[i+1] == '{') {
					b.WriteByte(c)
				}
				i++
			}
			parts = append(parts, b.String())
		}
	}

	return parts, params, nil
}
