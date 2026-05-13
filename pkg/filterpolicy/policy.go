// Package filterpolicy applies include/exclude predicates (include, then exclude).
package filterpolicy

import (
	"strings"

	"go.emeland.io/modelsrv/pkg/events"
	"go.emeland.io/modelsrv/pkg/model/annotations"

	"github.com/emeland-io/modelsrv-black-white-filter/pkg/filterconfig"
)

// Policy is derived from a validated [filterconfig.Config].
type Policy struct {
	includeTypes map[events.ResourceType]struct{}
	excludeTypes map[events.ResourceType]struct{}
	includeAnn   map[string]struct{}
	excludeAnn   map[string]struct{}
}

// FromConfig builds a [Policy].
func FromConfig(c filterconfig.Config) *Policy {
	return &Policy{
		includeTypes: c.IncludeTypes,
		excludeTypes: c.ExcludeTypes,
		includeAnn:   c.IncludeAnnotationKeys,
		excludeAnn:   c.ExcludeAnnotationKeys,
	}
}

// AllowResourceType reports whether a replicated event for rt may be applied (include, then exclude).
func (p *Policy) AllowResourceType(rt events.ResourceType) bool {
	if rt == events.UnknownResourceType {
		return false
	}
	if len(p.includeTypes) > 0 {
		if _, ok := p.includeTypes[rt]; !ok {
			return false
		}
	}
	if len(p.excludeTypes) > 0 {
		if _, ok := p.excludeTypes[rt]; ok {
			return false
		}
	}
	return true
}

type annotated interface {
	GetAnnotations() annotations.Annotations
}

// FilterAndRetainObject applies annotation filters to obj and reports whether it should remain in a
// batched replication event. It returns false for nil; when include-annotations is configured, it
// returns false for values that are not annotatable, and false when the object had annotation keys but
// none survived the include filter (all keys were disallowed). Exclude-only annotation rules keep any
// object (non-annotated values are unchanged by [FilterAnnotationsOnObject]).
func (p *Policy) FilterAndRetainObject(obj any) bool {
	if obj == nil {
		return false
	}
	if len(p.includeAnn) > 0 {
		a, ok := obj.(annotated)
		if !ok {
			return false
		}
		hadBefore := hasAnyAnnotationKeys(a)
		p.FilterAnnotationsOnObject(obj)
		if !hadBefore {
			return true
		}
		return hasAnyAnnotationKeys(a)
	}
	p.FilterAnnotationsOnObject(obj)
	return true
}

func hasAnyAnnotationKeys(a annotated) bool {
	if a == nil {
		return false
	}
	ann := a.GetAnnotations()
	if ann == nil {
		return false
	}
	for range ann.GetKeys() {
		return true
	}
	return false
}

// FilterAnnotationsOnObject strips annotation keys on a replicated resource object (mutates in place).
// It uses include-then-exclude semantics for annotation names (case insensitive).
func (p *Policy) FilterAnnotationsOnObject(obj any) {
	if obj == nil {
		return
	}
	a, ok := obj.(annotated)
	if !ok {
		return
	}
	ann := a.GetAnnotations()
	if ann == nil {
		return
	}
	var keys []string
	for k := range ann.GetKeys() {
		keys = append(keys, k)
	}
	for _, k := range keys {
		lk := strings.ToLower(k)
		if len(p.includeAnn) > 0 {
			if _, allowed := p.includeAnn[lk]; !allowed {
				ann.Delete(k)
				continue
			}
		}
		if len(p.excludeAnn) > 0 {
			if _, denied := p.excludeAnn[lk]; denied {
				ann.Delete(k)
			}
		}
	}
}
