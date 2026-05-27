package filterconfig

import (
	"fmt"
	"os"

	"gopkg.in/yaml.v3"
)

// configFile is the top-level YAML document (e.g. config/config.yaml).
type configFile struct {
	Filter filterSpec `yaml:"filter"`
}

// filterSpec is the filter block: a whitelist and a blacklist, each listing resources and annotations.
// Whitelist is applied first (only those types/keys if non-empty), then blacklist removes matches.
type filterSpec struct {
	Whitelist sideLists `yaml:"whitelist"`
	Blacklist sideLists `yaml:"blacklist"`
}

// sideLists holds resource-kind names and annotation key names for one side of the filter.
type sideLists struct {
	Resources   []string `yaml:"resources"`
	Annotations []string `yaml:"annotations"`
}

// LoadFile reads and parses a YAML file into [Config].
func LoadFile(path string) (Config, error) {
	b, err := os.ReadFile(path)
	if err != nil {
		return Config{}, fmt.Errorf("read config: %w", err)
	}
	return ParseYAML(b)
}

// ParseYAML decodes YAML bytes into [Config].
func ParseYAML(data []byte) (Config, error) {
	var doc configFile
	if err := yaml.Unmarshal(data, &doc); err != nil {
		return Config{}, fmt.Errorf("yaml: %w", err)
	}
	cfg, err := BuildConfig(
		doc.Filter.Whitelist.Resources,
		doc.Filter.Blacklist.Resources,
		doc.Filter.Whitelist.Annotations,
		doc.Filter.Blacklist.Annotations,
	)
	if err != nil {
		return Config{}, fmt.Errorf("filter: %w", err)
	}
	return cfg, nil
}
