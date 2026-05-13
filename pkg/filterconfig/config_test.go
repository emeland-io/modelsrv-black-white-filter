package filterconfig_test

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"go.emeland.io/modelsrv/pkg/events"

	"github.com/emeland-io/modelsrv-black-white-filter/pkg/filterconfig"
)

var _ = Describe("filterconfig.BuildConfig", func() {
	It("parses types case-insensitively and accepts ApiInstance alias", func() {
		cfg, err := filterconfig.BuildConfig([]string{"system", "apiinstance", "NODE"}, nil, nil, nil)
		Expect(err).NotTo(HaveOccurred())
		Expect(cfg.IncludeTypes).To(HaveKey(events.SystemResource))
		Expect(cfg.IncludeTypes).To(HaveKey(events.APIInstanceResource))
		Expect(cfg.IncludeTypes).To(HaveKey(events.NodeResource))
	})

	It("returns an error for unknown resource types", func() {
		_, err := filterconfig.BuildConfig([]string{"System", "NotReal"}, nil, nil, nil)
		Expect(err).To(HaveOccurred())
	})

	It("normalizes annotation keys to a lowercased set", func() {
		cfg, err := filterconfig.BuildConfig(nil, nil, []string{"Foo", "bar", "BAZ"}, nil)
		Expect(err).NotTo(HaveOccurred())
		Expect(cfg.IncludeAnnotationKeys).To(HaveKey("foo"))
		Expect(cfg.IncludeAnnotationKeys).To(HaveKey("bar"))
		Expect(cfg.IncludeAnnotationKeys).To(HaveKey("baz"))
	})
})

var _ = Describe("filterconfig.ParseYAML", func() {
	It("loads whitelist and blacklist from YAML", func() {
		y := []byte(`
filter:
  whitelist:
    resources: [System, Node]
    annotations: [Keep]
  blacklist:
    resources: [Node]
    annotations: [drop]
`)
		cfg, err := filterconfig.ParseYAML(y)
		Expect(err).NotTo(HaveOccurred())
		Expect(cfg.IncludeTypes).To(HaveKey(events.SystemResource))
		Expect(cfg.IncludeTypes).To(HaveKey(events.NodeResource))
		Expect(cfg.ExcludeTypes).To(HaveKey(events.NodeResource))
		Expect(cfg.IncludeAnnotationKeys).To(HaveKey("keep"))
		Expect(cfg.ExcludeAnnotationKeys).To(HaveKey("drop"))
	})

	It("loads resources-only lists", func() {
		y := []byte(`
filter:
  whitelist:
    resources: [API]
  blacklist:
    resources: []
    annotations: []
`)
		cfg, err := filterconfig.ParseYAML(y)
		Expect(err).NotTo(HaveOccurred())
		Expect(cfg.IncludeTypes).To(HaveKey(events.APIResource))
	})
})
