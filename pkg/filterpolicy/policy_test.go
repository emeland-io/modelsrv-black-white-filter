package filterpolicy_test

import (
	"github.com/google/uuid"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"go.emeland.io/modelsrv/pkg/events"
	"go.emeland.io/modelsrv/pkg/model/system"

	"github.com/emeland-io/modelsrv-black-white-filter/pkg/filterconfig"
	"github.com/emeland-io/modelsrv-black-white-filter/pkg/filterpolicy"
)

var _ = Describe("filterpolicy.Policy", func() {
	Describe("AllowResourceType", func() {
		It("allows everything when no include/exclude lists", func() {
			p := filterpolicy.FromConfig(filterconfig.Config{})
			Expect(p.AllowResourceType(events.SystemResource)).To(BeTrue())
		})

		It("allows only included types", func() {
			cfg, err := filterconfig.BuildConfig([]string{"System"}, nil, nil, nil)
			Expect(err).NotTo(HaveOccurred())
			p := filterpolicy.FromConfig(cfg)
			Expect(p.AllowResourceType(events.SystemResource)).To(BeTrue())
			Expect(p.AllowResourceType(events.NodeResource)).To(BeFalse())
		})

		It("applies exclude after include", func() {
			cfg, err := filterconfig.BuildConfig([]string{"System", "Node"}, []string{"Node"}, nil, nil)
			Expect(err).NotTo(HaveOccurred())
			p := filterpolicy.FromConfig(cfg)
			Expect(p.AllowResourceType(events.SystemResource)).To(BeTrue())
			Expect(p.AllowResourceType(events.NodeResource)).To(BeFalse())
		})
	})

	Describe("FilterAnnotationsOnObject", func() {
		It("keeps only included keys then removes excluded (case insensitive)", func() {
			cfg, err := filterconfig.BuildConfig(nil, nil, []string{"KeepA", "KeepB"}, []string{"keepb"})
			Expect(err).NotTo(HaveOccurred())
			p := filterpolicy.FromConfig(cfg)
			s := system.NewSystem(events.NewDummySink(), uuid.New())
			s.GetAnnotations().Add("keepA", "1")
			s.GetAnnotations().Add("KeepB", "2")
			s.GetAnnotations().Add("Drop", "3")
			p.FilterAnnotationsOnObject(s)
			Expect(s.GetAnnotations().GetValue("keepA")).To(Equal("1"))
			Expect(s.GetAnnotations().GetValue("KeepB")).To(BeEmpty())
			Expect(s.GetAnnotations().GetValue("Drop")).To(BeEmpty())
		})
	})

	Describe("FilterAndRetainObject", func() {
		It("drops non-annotated values when include-annotations is set", func() {
			cfg, err := filterconfig.BuildConfig(nil, nil, []string{"x"}, nil)
			Expect(err).NotTo(HaveOccurred())
			p := filterpolicy.FromConfig(cfg)
			Expect(p.FilterAndRetainObject(struct{}{})).To(BeFalse())
		})

		It("drops annotatable objects that had keys but none survive the include list", func() {
			cfg, err := filterconfig.BuildConfig(nil, nil, []string{"allowed"}, nil)
			Expect(err).NotTo(HaveOccurred())
			p := filterpolicy.FromConfig(cfg)
			s := system.NewSystem(events.NewDummySink(), uuid.New())
			s.GetAnnotations().Add("other", "1")
			Expect(p.FilterAndRetainObject(s)).To(BeFalse())
			n := 0
			for range s.GetAnnotations().GetKeys() {
				n++
			}
			Expect(n).To(Equal(0))
		})

		It("retains annotatable objects with no keys when include-annotations is set", func() {
			cfg, err := filterconfig.BuildConfig(nil, nil, []string{"allowed"}, nil)
			Expect(err).NotTo(HaveOccurred())
			p := filterpolicy.FromConfig(cfg)
			s := system.NewSystem(events.NewDummySink(), uuid.New())
			Expect(p.FilterAndRetainObject(s)).To(BeTrue())
		})
	})
})
