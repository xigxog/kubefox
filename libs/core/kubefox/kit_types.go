package kubefox

type Entrypoint func(Kit) error

type EntrypointMatcher struct {
	entrypoint Entrypoint
	matcher    *EventMatcher
}

func (m *EntrypointMatcher) Rule() string {
	return m.matcher.Rule()
}

func (m *EntrypointMatcher) Match(evt DataEvent) bool {
	return m.matcher.Match(evt)
}
