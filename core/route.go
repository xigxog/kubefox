package core

import "github.com/xigxog/kubefox/api"

type Route struct {
	*Rule

	ResolvedRule string
	Priority     int

	Component    *Component
	EventContext *EventContext
}

func (r *Route) Resolve(data *api.VirtualEnvData) error {
	resolved, priority, err := r.Rule.Resolve(data)
	if err != nil {
		return err
	}

	r.ResolvedRule = resolved
	r.Priority = priority

	return nil
}
