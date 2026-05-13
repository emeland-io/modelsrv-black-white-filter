// Package filterconfig parses blacklist/whitelist options for resource types and annotations.
package filterconfig

import (
	"fmt"
	"strings"

	"go.emeland.io/modelsrv/pkg/events"
)

// Config is validated filter configuration (whitelist = include, blacklist = exclude; whitelist first).
type Config struct {
	IncludeTypes          map[events.ResourceType]struct{}
	ExcludeTypes          map[events.ResourceType]struct{}
	IncludeAnnotationKeys map[string]struct{}
	ExcludeAnnotationKeys map[string]struct{}
}

var filterableTypes = []events.ResourceType{
	events.NodeResource,
	events.NodeTypeResource,
	events.ContextResource,
	events.ContextTypeResource,
	events.SystemResource,
	events.SystemInstanceResource,
	events.APIResource,
	events.APIInstanceResource,
	events.ComponentResource,
	events.ComponentInstanceResource,
	events.OrgUnitResource,
	events.GroupResource,
	events.IdentityResource,
	events.FindingResource,
	events.FindingTypeResource,
}

func parseOneResourceType(token string) (events.ResourceType, error) {
	t := strings.TrimSpace(token)
	if t == "" {
		return events.UnknownResourceType, fmt.Errorf("empty type token")
	}
	for _, rt := range filterableTypes {
		if strings.EqualFold(t, rt.String()) {
			return rt, nil
		}
		if strings.EqualFold(t, rt.WireKind()) {
			return rt, nil
		}
	}
	return events.UnknownResourceType, fmt.Errorf("unknown resource type %q", token)
}

// BuildConfig builds [Config] from whitelist/blacklist token lists (same semantics as YAML).
// Pass nil or empty slices for a dimension that is unconstrained.
func BuildConfig(whitelistResources, blacklistResources, whitelistAnnotations, blacklistAnnotations []string) (Config, error) {
	incTypes, err := ParseTypeTokens(whitelistResources)
	if err != nil {
		return Config{}, fmt.Errorf("whitelist resources: %w", err)
	}
	excTypes, err := ParseTypeTokens(blacklistResources)
	if err != nil {
		return Config{}, fmt.Errorf("blacklist resources: %w", err)
	}
	incAnn := NormalizeAnnotationKeys(whitelistAnnotations)
	excAnn := NormalizeAnnotationKeys(blacklistAnnotations)
	return Config{
		IncludeTypes:          incTypes,
		ExcludeTypes:          excTypes,
		IncludeAnnotationKeys: incAnn,
		ExcludeAnnotationKeys: excAnn,
	}, nil
}
