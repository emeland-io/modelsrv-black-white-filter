package filteredmodel_test

import (
	"context"

	"github.com/google/uuid"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"go.uber.org/mock/gomock"

	"go.emeland.io/modelsrv/pkg/backend"
	"go.emeland.io/modelsrv/pkg/events"
	"go.emeland.io/modelsrv/pkg/mocks"
	"go.emeland.io/modelsrv/pkg/model/system"

	"github.com/emeland-io/modelsrv-black-white-filter/pkg/filterconfig"
	"github.com/emeland-io/modelsrv-black-white-filter/pkg/filteredmodel"
	"github.com/emeland-io/modelsrv-black-white-filter/pkg/filterpolicy"
)

var _ = Describe("filteredmodel.Wrap", func() {
	It("drops Apply for disallowed resource types (mock)", func() {
		cfg, err := filterconfig.BuildConfig([]string{"System"}, nil, nil, nil)
		Expect(err).NotTo(HaveOccurred())
		pol := filterpolicy.FromConfig(cfg)

		ctrl := gomock.NewController(GinkgoT())
		defer ctrl.Finish()
		inner := mocks.NewMockModel(ctrl)
		inner.EXPECT().Apply(gomock.Any()).Times(0)

		w := filteredmodel.Wrap(inner, pol)
		ev := events.Event{
			ResourceType: events.NodeResource,
			Operation:    events.CreateOperation,
			ResourceId:   uuid.New(),
			Objects:      []any{struct{}{}},
		}
		Expect(w.Apply(ev)).To(Succeed())
	})

	It("delegates Apply for allowed types (mock)", func() {
		cfg, err := filterconfig.BuildConfig([]string{"System"}, nil, nil, nil)
		Expect(err).NotTo(HaveOccurred())
		pol := filterpolicy.FromConfig(cfg)

		ctrl := gomock.NewController(GinkgoT())
		defer ctrl.Finish()
		inner := mocks.NewMockModel(ctrl)
		inner.EXPECT().Apply(gomock.Any()).Return(nil).Times(1)

		w := filteredmodel.Wrap(inner, pol)
		ev := events.Event{
			ResourceType: events.SystemResource,
			Operation:    events.CreateOperation,
			ResourceId:   uuid.New(),
			Objects:      []any{struct{}{}},
		}
		Expect(w.Apply(ev)).To(Succeed())
	})

	It("does not store rejected resources on a real backend model", func() {
		cfg, err := filterconfig.BuildConfig([]string{"System"}, nil, nil, nil)
		Expect(err).NotTo(HaveOccurred())
		pol := filterpolicy.FromConfig(cfg)

		b, err := backend.New()
		Expect(err).NotTo(HaveOccurred())
		w := filteredmodel.Wrap(b.GetModel(), pol)

		ctx := context.Background()
		seq0, err := b.GetEventManager().GetCurrentSequenceId(ctx)
		Expect(err).NotTo(HaveOccurred())

		nid := uuid.New()
		nodeObj := events.Event{
			ResourceType: events.NodeResource,
			Operation:    events.CreateOperation,
			ResourceId:   nid,
			Objects:      []any{nodeStub{id: nid}},
		}
		Expect(w.Apply(nodeObj)).To(Succeed())
		Expect(b.GetModel().GetNodeById(nid)).To(BeNil())

		seq1, err := b.GetEventManager().GetCurrentSequenceId(ctx)
		Expect(err).NotTo(HaveOccurred())
		Expect(seq1).To(Equal(seq0))
	})

	It("stores allowed resources and strips annotations before apply", func() {
		cfg, err := filterconfig.BuildConfig([]string{"System"}, nil, []string{"keep"}, []string{"drop"})
		Expect(err).NotTo(HaveOccurred())
		pol := filterpolicy.FromConfig(cfg)

		b, err := backend.New()
		Expect(err).NotTo(HaveOccurred())
		w := filteredmodel.Wrap(b.GetModel(), pol)

		sid := uuid.New()
		sys := system.NewSystem(b.GetModel().GetSink(), sid)
		sys.SetDisplayName("svc")
		sys.GetAnnotations().Add("Keep", "1")
		sys.GetAnnotations().Add("DROP", "2")
		sys.GetAnnotations().Add("Other", "3")

		ev := events.Event{
			ResourceType: events.SystemResource,
			Operation:    events.CreateOperation,
			ResourceId:   sid,
			Objects:      []any{sys},
		}
		Expect(w.Apply(ev)).To(Succeed())

		got := b.GetModel().GetSystemById(sid)
		Expect(got).NotTo(BeNil())
		Expect(got.GetAnnotations().GetValue("Keep")).To(Equal("1"))
		Expect(got.GetAnnotations().GetValue("DROP")).To(BeEmpty())
		Expect(got.GetAnnotations().GetValue("Other")).To(BeEmpty())
	})

	It("applies only objects that remain after per-object annotation filtering", func() {
		cfg, err := filterconfig.BuildConfig([]string{"System"}, nil, []string{"keep"}, nil)
		Expect(err).NotTo(HaveOccurred())
		pol := filterpolicy.FromConfig(cfg)

		b, err := backend.New()
		Expect(err).NotTo(HaveOccurred())
		w := filteredmodel.Wrap(b.GetModel(), pol)

		goodID := uuid.New()
		good := system.NewSystem(b.GetModel().GetSink(), goodID)
		good.SetDisplayName("keep-me")
		good.GetAnnotations().Add("keep", "1")

		badID := uuid.New()
		bad := system.NewSystem(b.GetModel().GetSink(), badID)
		bad.SetDisplayName("drop-me")
		bad.GetAnnotations().Add("other", "2")

		ev := events.Event{
			ResourceType: events.SystemResource,
			Operation:    events.CreateOperation,
			ResourceId:   goodID,
			Objects:      []any{bad, good},
		}
		Expect(w.Apply(ev)).To(Succeed())

		Expect(b.GetModel().GetSystemById(goodID)).NotTo(BeNil())
		Expect(b.GetModel().GetSystemById(badID)).To(BeNil())
	})
})

type nodeStub struct {
	id uuid.UUID
}
