// Copyright (c) Mainflux
// SPDX-License-Identifier: Apache-2.0

package main

import (
	"context"
	"fmt"
	"log"
	"os"

	"github.com/mainflux/mainflux/coap"
	"github.com/mainflux/mainflux/coap/api"
	"github.com/mainflux/mainflux/coap/tracing"
	"github.com/mainflux/mainflux/internal"
	thingsClient "github.com/mainflux/mainflux/internal/clients/grpc/things"
	jaegerClient "github.com/mainflux/mainflux/internal/clients/jaeger"
	"github.com/mainflux/mainflux/internal/env"
	"github.com/mainflux/mainflux/internal/server"
	coapserver "github.com/mainflux/mainflux/internal/server/coap"
	httpserver "github.com/mainflux/mainflux/internal/server/http"
	mflog "github.com/mainflux/mainflux/logger"
	"github.com/mainflux/mainflux/pkg/messaging/brokers"
	pstracing "github.com/mainflux/mainflux/pkg/messaging/tracing"
	"golang.org/x/sync/errgroup"
)

const (
	svcName        = "coap_adapter"
	envPrefix      = "MF_COAP_ADAPTER_"
	envPrefixHttp  = "MF_COAP_ADAPTER_HTTP_"
	envPrefixCoap  = "MF_COAP_ADAPTER_COAP_"
	defSvcHttpPort = "5683"
	defSvcCoapPort = "5683"
)

type config struct {
	LogLevel  string `env:"MF_INFLUX_READER_LOG_LEVEL"  envDefault:"info"`
	BrokerURL string `env:"MF_BROKER_URL"               envDefault:"nats://localhost:4222"`
	JaegerURL string `env:"MF_JAEGER_URL"               envDefault:"localhost:6831"`
}

func main() {
	ctx, cancel := context.WithCancel(context.Background())
	g, ctx := errgroup.WithContext(ctx)

	cfg := config{}
	if err := env.Parse(&cfg); err != nil {
		log.Fatalf("failed to load %s configuration : %s", svcName, err)
	}

	logger, err := mflog.New(os.Stdout, cfg.LogLevel)
	if err != nil {
		log.Fatalf("failed to init logger: %s", err)
	}

	tc, tcHandler, err := thingsClient.Setup(envPrefix, cfg.JaegerURL)
	if err != nil {
		logger.Fatal(err.Error())
	}
	defer tcHandler.Close()
	logger.Info("Successfully connected to things grpc server " + tcHandler.Secure())

	tracer, traceCloser, err := jaegerClient.NewTracer(svcName, cfg.JaegerURL)
	if err != nil {
		logger.Fatal(fmt.Sprintf("failed to init Jaeger: %s", err))
	}
	defer traceCloser.Close()

	nps, err := brokers.NewPubSub(cfg.BrokerURL, "", logger)
	if err != nil {
		logger.Fatal(fmt.Sprintf("failed to connect to message broker: %s", err))
	}
	nps = pstracing.NewPubSub(tracer, nps)
	defer nps.Close()

	svc := coap.New(tc, nps)

	svc = tracing.New(tracer, svc)

	svc = api.LoggingMiddleware(svc, logger)

	counter, latency := internal.MakeMetrics(svcName, "api")
	svc = api.MetricsMiddleware(svc, counter, latency)

	httpServerConfig := server.Config{Port: defSvcHttpPort}
	if err := env.Parse(&httpServerConfig, env.Options{Prefix: envPrefixHttp, AltPrefix: envPrefix}); err != nil {
		logger.Fatal(fmt.Sprintf("failed to load %s HTTP server configuration : %s", svcName, err))
	}
	hs := httpserver.New(ctx, cancel, svcName, httpServerConfig, api.MakeHTTPHandler(), logger)

	coapServerConfig := server.Config{Port: defSvcCoapPort}
	if err := env.Parse(&httpServerConfig, env.Options{Prefix: envPrefixCoap, AltPrefix: envPrefix}); err != nil {
		logger.Fatal(fmt.Sprintf("failed to load %s CoAP server configuration : %s", svcName, err))
	}
	cs := coapserver.New(ctx, cancel, svcName, coapServerConfig, api.MakeCoAPHandler(svc, logger), logger)

	g.Go(func() error {
		return hs.Start()
	})
	g.Go(func() error {
		return cs.Start()
	})
	g.Go(func() error {
		return server.StopSignalHandler(ctx, cancel, logger, svcName, hs, cs)
	})

	if err := g.Wait(); err != nil {
		logger.Error(fmt.Sprintf("CoAP adapter service terminated: %s", err))
	}
}
