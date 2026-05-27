// Package server wires a filtered follower modelserver on top of go.emeland.io/modelsrv.
package server

import (
	"context"
	"fmt"
	"log/slog"
	"net"
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
	// ListenAddr is passed to [endpoint.StartWebListener] (e.g. "localhost:8080").
	ListenAddr string
	// UpstreamAPIBase is the upstream API root including /api (e.g. "http://host:9090/api").
	UpstreamAPIBase string
	// CallbackURL is the URL the upstream will use to reach this server's /api endpoint.
	// When empty it is derived from ListenAddr, substituting 0.0.0.0 or a bare port
	// with localhost so the result is always routable.
	CallbackURL string
	Filter      filterconfig.Config
	// RegisterRetry is how long to sleep between upstream registration attempts.
	RegisterRetry time.Duration
	// RegisterTimeout caps the total time spent retrying upstream registration.
	// Defaults to 30 s when zero.
	RegisterTimeout time.Duration
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
	timeout := cfg.RegisterTimeout
	if timeout <= 0 {
		timeout = 30 * time.Second
	}

	callback := cfg.CallbackURL
	if callback == "" {
		var err error
		callback, err = routableCallback(cfg.ListenAddr)
		if err != nil {
			return err
		}
	}

	bundle, err := NewBundle(cfg.Filter)
	if err != nil {
		return fmt.Errorf("backend: %w", err)
	}
	b := bundle.Backend()

	if err := endpoint.StartWebListener(bundle.Model, b.GetEventManager(), cfg.ListenAddr); err != nil {
		return fmt.Errorf("web listener: %w", err)
	}

	upstream, err := client.NewModelSrvClient(cfg.UpstreamAPIBase)
	if err != nil {
		return fmt.Errorf("upstream client: %w", err)
	}

	slog.Info("registering with upstream", "callback", callback, "upstream", cfg.UpstreamAPIBase, "timeout", timeout)
	var lastErr error
	for attempt, start := 1, time.Now(); time.Since(start) < timeout; time.Sleep(retry) {
		if err := upstream.Register(callback); err != nil {
			slog.Warn("upstream registration attempt failed", "attempt", attempt, "elapsed", time.Since(start).Round(time.Millisecond), "err", err)
			lastErr = err
			attempt++
			continue
		}
		slog.Info("registered with upstream", "callback", callback)
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

// routableCallback derives the upstream callback URL from listenAddr, replacing
// a wildcard host (empty, 0.0.0.0, ::) with localhost so the URL is always routable.
func routableCallback(listenAddr string) (string, error) {
	host, port, err := net.SplitHostPort(listenAddr)
	if err != nil {
		return "", fmt.Errorf("invalid listen address %q: %w", listenAddr, err)
	}
	if host == "" || host == "0.0.0.0" || host == "::" {
		host = "localhost"
	}
	return fmt.Sprintf("http://%s:%s/api", host, port), nil
}
