// Command main is the standalone filtered modelserver binary.
package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/emeland-io/modelsrv-black-white-filter/pkg/filterconfig"
	"github.com/emeland-io/modelsrv-black-white-filter/pkg/server"
)

func main() {
	listen := flag.String("listen", "localhost:8080", "TCP address for this server (host:port)")
	upstream := flag.String("upstream", "", "Upstream modelsrv API base URL including /api (required)")
	configPath := flag.String("config", "config/config.yaml", "Path to YAML file with filter whitelist/blacklist")
	flag.Parse()

	if *upstream == "" {
		fmt.Fprintln(os.Stderr, "missing required -upstream")
		flag.Usage()
		os.Exit(2)
	}

	fc, err := filterconfig.LoadFile(*configPath)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	if err := server.Run(ctx, server.Config{
		ListenAddr:      *listen,
		UpstreamAPIBase: *upstream,
		Filter:          fc,
	}); err != nil && ctx.Err() == nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
