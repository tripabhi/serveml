package main

import (
	"context"
	"fmt"
	"log"
	"net"
	"net/http"
	"net/http/httputil"
	"net/url"
	"os"
	"strconv"
	"time"

	"github.com/pkg/errors"

	pkgnet "knative.dev/pkg/network"
	"knative.dev/pkg/signals"

	"knative.dev/networking/pkg/http/proxy"

	"github.com/tripabhi/serveml/pkg/safebatcher"

	flag "github.com/spf13/pflag"
)

var (
	port = flag.String("port", "9081", "Worker Agent's port")
	svcPort = flag.String("service-port", "8000", "Port of Service you want to forward requests to")
	enableBatching = flag.Bool("enable-batching", true, "Enable request batcher")
	maxBatchSize  = flag.String("max-batchsize", "32", "Max Batch Size")
	maxLatency    = flag.String("max-latency", "2000", "Max Latency in milliseconds")
	idleTimeout = flag.String("idle-timeout", "30000", "Idle timeout for each request")
)

func main() {
	flag.Parse()

	var batcherConfig *safebatcher.Config
	if *enableBatching {
		log.Println("Enabled Batching")
		batcherConfig = getBatcherConfig()
	}

	var ctx context.Context = signals.NewContext()

	mainServer := buildServer(ctx, *port , *svcPort, batcherConfig)

	servers := map[string]*http.Server{
		"main" : mainServer,
	}

	errCh := make(chan error)

	for name, server := range servers {
		go func(name string, s *http.Server) {
			l, err := net.Listen("tcp", s.Addr)
			if err != nil {
				errCh <- fmt.Errorf("%s server failed to listen : %w", name, err)
				return
			}

			log.Printf("Serving Server at %s\n", s.Addr)
			if err := s.Serve(l); err != nil && !errors.Is(err, http.ErrServerClosed) {
				errCh <- fmt.Errorf("%s server failed to serve : %w", name, err)
			}
		}(name, server)
	}

	select {
	case err := <-errCh:
		log.Println("Failed to bring up agent, shutting down: "+err.Error())
		os.Exit(1)
	case <-ctx.Done():
		log.Println("Received TERM, attempting to gracefully shutdown servers")
		for serverName, srv := range servers {
			if err := srv.Shutdown(context.Background()); err != nil {
				log.Printf("Failed to shutdown Server %s due to error : %s", serverName, err.Error())
			}
		}
		log.Println("Shutdown complete, exiting...")
	}

}

func getBatcherConfig() *safebatcher.Config {
	maxBatchSizeInt, err := strconv.Atoi(*maxBatchSize)
	if err != nil || maxBatchSizeInt <= 0 {
		log.Printf("Non-Integer max-batchsize provided : %s\n", *maxBatchSize)
		os.Exit(1)
	}

	maxLatencyInt, err := strconv.Atoi(*maxLatency)
	if err != nil || maxLatencyInt <= 0 {
		log.Printf("Non-Integer max-latency provided : %s\n", *maxLatency)
		os.Exit(1)
	}

	idleTimeoutInt, err := strconv.Atoi(*idleTimeout)
	if err != nil || maxLatencyInt <= 0 {
		log.Printf("Non-Integer max-latency provided : %s\n", *idleTimeout)
		os.Exit(1)
	}

	return &safebatcher.Config{
		MaxLatency: time.Duration(maxLatencyInt) * time.Millisecond,
		MaxBatchSize: maxBatchSizeInt,
		IdleTimeout: time.Duration(idleTimeoutInt) * time.Millisecond,
	}
}

func buildServer(ctx context.Context, port string, svcPort string, batcherConfig *safebatcher.Config) *http.Server { 
	target := &url.URL{
		Scheme: "http",
		Host: net.JoinHostPort("127.0.0.1", svcPort),
	}

	maxIdleConns := 1000

	httpProxy := httputil.NewSingleHostReverseProxy(target)
	httpProxy.Transport = pkgnet.NewAutoTransport(maxIdleConns /* max-idle conns */, 
		maxIdleConns /* max-idle conns per host*/ )
	httpProxy.BufferPool = proxy.NewBufferPool()
	httpProxy.FlushInterval = proxy.FlushInterval

	var composedHandler http.Handler = httpProxy

	if batcherConfig != nil {
		composedHandler = safebatcher.NewPredictionRequestBatcher(ctx, batcherConfig, composedHandler)
	}

	return pkgnet.NewServer(":"+port, composedHandler)
}