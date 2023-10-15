package engine

import (
	"github.com/xigxog/kubefox/libs/core/kubefox"
	"github.com/xigxog/kubefox/libs/core/matcher"
)

type Matcher interface {
	Match(evt *kubefox.Event) *kubefox.MatchedEvent
}

type DeploymentMatcher struct {
	Deployment  string
	Environment string
	Matchers    []*matcher.EventMatcher
	Error       error
}

type ReleaseMatchers map[string]*ReleaseMatcher

type ReleaseMatcher struct {
	Release  string
	Matchers []*matcher.EventMatcher
	Error    error
}

func (rms ReleaseMatchers) Match(evt *kubefox.Event) *kubefox.MatchedEvent {
	for _, rm := range rms {
		if rm.Error != nil {
			continue
		}
		for _, m := range rm.Matchers {
			if comp, rt, match := m.Match(evt); match {
				if !updateTarget(comp, evt) {
					continue
				}
				evt.Release = rm.Release
				return &kubefox.MatchedEvent{
					Event:   evt,
					RouteId: int64(rt.Id),
				}
			}
		}
	}

	return nil
}

func (dm *DeploymentMatcher) Match(evt *kubefox.Event) *kubefox.MatchedEvent {
	if dm.Error != nil {
		return nil
	}

	for _, m := range dm.Matchers {
		if comp, rt, match := m.Match(evt); match {
			if !updateTarget(comp, evt) {
				continue
			}
			return &kubefox.MatchedEvent{
				Event:   evt,
				RouteId: int64(rt.Id),
			}
		}
	}

	return nil
}

func updateTarget(tgt *kubefox.Component, evt *kubefox.Event) bool {
	if evt.Target == nil {
		evt.Target = tgt
		return true
	}

	if evt.Target.Name != tgt.Name {
		return false
	}
	if evt.Target.Commit == "" {
		evt.Target.Commit = tgt.Commit
	}
	if evt.Target.Commit != tgt.Commit {
		return false
	}

	return true
}
