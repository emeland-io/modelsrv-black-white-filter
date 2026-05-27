// Package filteredmodel wraps [model.Model] and filters [events.Event] before Apply.
package filteredmodel

import (
	"go.emeland.io/modelsrv/pkg/events"
	"go.emeland.io/modelsrv/pkg/model"

	"github.com/emeland-io/modelsrv-black-white-filter/pkg/filterpolicy"
)

// Wrap returns a [model.Model] that delegates to inner except for [model.Model.Apply],
// which enforces resource-type and annotation predicates before storage.
func Wrap(inner model.Model, policy *filterpolicy.Policy) model.Model {
	if policy == nil {
		return inner
	}
	return &wrapped{
		Model:  inner,
		policy: policy,
	}
}

type wrapped struct {
	model.Model
	policy *filterpolicy.Policy
}

// Apply implements [model.Model] (via embedding); filtered events are dropped and never stored.
func (w *wrapped) Apply(ev events.Event) error {
	if !w.policy.AllowResourceType(ev.ResourceType) {
		return nil
	}
	if ev.Operation == events.DeleteOperation || len(ev.Objects) == 0 {
		return w.Model.Apply(ev)
	}
	var kept []any
	for _, o := range ev.Objects {
		if w.policy.FilterAndRetainObject(o) {
			kept = append(kept, o)
		}
	}
	if len(kept) == 0 {
		return nil
	}
	ev.Objects = kept
	return w.Model.Apply(ev)
}
