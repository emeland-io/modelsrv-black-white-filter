// Package server wires a filtered follower modelserver on top of go.emeland.io/modelsrv.
package server

import (
	"context"
	"fmt"
	"time"

	"go.emeland.io/modelsrv/pkg/backend"
	"go.emeland.io/modelsrv/pkg/client"
	"go.emeland.io/modelsrv/pkg/endpoint"
	"go.emeland.io/modelsrv/pkg/model"

	"github.com/emeland-io/modelsrv-black-white-filter/pkg/filterconfig"
	"github.com/emeland-io/modelsrv-black-white-filter/pkg/filteredmodel"
	"github.com/emeland-io/modelsrv-black-white-filter/pkg/filterpolicy"
)

// Bundle is a wired backend with the filtered API model and shared event manager.
type Bundle struct {
	backend backend.Backend
	// Model is the filtered view used for HTTP Apply and reads.
	Model model.Model
}

// Backend returns the underlying [backend.Backend] (recording sink, event manager).
func (b *Bundle) Backend() backend.Backend { return b.backend }

// NewBundle builds a [backend.Backend] without changing upstream libraries and wraps [model.Model.Apply].
func NewBundle(fc filterconfig.Config) (*Bundle, error) {
	b, err := backend.New()
	if err != nil {
		return nil, err
	}
	pol := filterpolicy.FromConfig(fc)
	return &Bundle{
		backend: b,
		Model:   filteredmodel.Wrap(b.GetModel(), pol),
	}, nil
}

// Config drives listener and upstream registration.
type Config struct {
	// ListenAddr is passed to [endpoint.StarWebListener] (e.g. "localhost:8080").
	ListenAddr string
	// UpstreamAPIBase is the upstream API root including /api (e.g. "http://host:9090/api").
	UpstreamAPIBase string
	Filter          filterconfig.Config
	// RegisterRetry is how long to wait for the HTTP listener before POST /events/register.
	RegisterRetry time.Duration
}

// Run starts the web listener, wraps the model with filters, and registers with the upstream.
func Run(ctx context.Context, cfg Config) error {
	if cfg.ListenAddr == "" {
		return fmt.Errorf("listen address required")
	}
	if cfg.UpstreamAPIBase == "" {
		return fmt.Errorf("upstream API base URL required")
	}
	retry := cfg.RegisterRetry
	if retry <= 0 {
		retry = 500 * time.Millisecond
	}

	bundle, err := NewBundle(cfg.Filter)
	if err != nil {
		return fmt.Errorf("backend: %w", err)
	}
	b := bundle.Backend()

	if err := endpoint.StartWebListener(bundle.Model, b.GetEventManager(), cfg.ListenAddr); err != nil {
		return fmt.Errorf("web listener: %w", err)
	}

	callback := fmt.Sprintf("http://%s/api", cfg.ListenAddr)
	upstream, err := client.NewModelSrvClient(cfg.UpstreamAPIBase)
	if err != nil {
		return fmt.Errorf("upstream client: %w", err)
	}

	var lastErr error
	for start := time.Now(); time.Since(start) < 30*time.Second; time.Sleep(retry) {
		if err := upstream.Register(callback); err != nil {
			lastErr = err
			continue
		}
		lastErr = nil
		break
	}
	if lastErr != nil {
		return fmt.Errorf("register with upstream: %w", lastErr)
	}

	<-ctx.Done()
	endpoint.StopWebListener()
	return ctx.Err()
}
