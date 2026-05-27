package filterconfig

import (
	"strings"

	"go.emeland.io/modelsrv/pkg/events"
)

// ParseTypeTokens parses resource kind names into a set (case-insensitive; accepts WireKind aliases).
func ParseTypeTokens(tokens []string) (map[events.ResourceType]struct{}, error) {
	out := make(map[events.ResourceType]struct{})
	for _, raw := range tokens {
		rt, err := parseOneResourceType(raw)
		if err != nil {
			return nil, err
		}
		out[rt] = struct{}{}
	}
	if len(out) == 0 {
		return nil, nil
	}
	return out, nil
}

// NormalizeAnnotationKeys returns a lowercased set of annotation names (trims space, skips empties).
func NormalizeAnnotationKeys(tokens []string) map[string]struct{} {
	out := make(map[string]struct{})
	for _, raw := range tokens {
		s := strings.TrimSpace(raw)
		if s == "" {
			continue
		}
		out[strings.ToLower(s)] = struct{}{}
	}
	if len(out) == 0 {
		return nil
	}
	return out
}
