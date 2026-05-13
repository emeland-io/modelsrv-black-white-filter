package server_test

import (
	"net/http"
	"net/http/httptest"
	"sync/atomic"
	"testing"

	"github.com/google/uuid"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"go.emeland.io/modelsrv/pkg/events"
	"go.emeland.io/modelsrv/pkg/model/system"

	"github.com/emeland-io/modelsrv-black-white-filter/pkg/filterconfig"
	"github.com/emeland-io/modelsrv-black-white-filter/pkg/server"
)

func TestServer(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "server Suite")
}

var _ = Describe("server.Bundle recording and subscriber path", func() {
	It("notifies HTTP subscribers only when filtered Apply records an event", func() {
		var pushes atomic.Int32
		mux := http.NewServeMux()
		mux.HandleFunc("/api/events/push", func(w http.ResponseWriter, r *http.Request) {
			pushes.Add(1)
			w.WriteHeader(http.StatusOK)
		})
		srv := httptest.NewServer(mux)
		defer srv.Close()

		cfg, err := filterconfig.BuildConfig([]string{"System"}, nil, nil, nil)
		Expect(err).NotTo(HaveOccurred())
		bundle, err := server.NewBundle(cfg)
		Expect(err).NotTo(HaveOccurred())
		Expect(bundle.Backend().GetEventManager().AddSubscriber(srv.URL + "/api")).To(Succeed())

		sid := uuid.New()
		sys := system.NewSystem(bundle.Backend().GetModel().GetSink(), sid)
		sys.SetDisplayName("x")
		ev := events.Event{
			ResourceType: events.SystemResource,
			Operation:    events.CreateOperation,
			ResourceId:   sid,
			Objects:      []any{sys},
		}
		Expect(bundle.Model.Apply(ev)).To(Succeed())

		Eventually(pushes.Load, "2s", "10ms").Should(BeNumerically(">=", 1))

		pushes.Store(0)
		nodeID := uuid.New()
		nodeEv := events.Event{
			ResourceType: events.NodeResource,
			Operation:    events.CreateOperation,
			ResourceId:   nodeID,
			Objects:      []any{struct{}{}},
		}
		Expect(bundle.Model.Apply(nodeEv)).To(Succeed())

		Consistently(pushes.Load, "300ms", "30ms").Should(Equal(int32(0)))
	})
})
