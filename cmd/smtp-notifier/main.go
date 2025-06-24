// Copyright (c) Mainflux
// SPDX-License-Identifier: Apache-2.0

// Package main contains smtp-notifier main function to start the smtp-notifier service.
package main

import (
	"context"
	"fmt"
	"log"
	"os"

	"github.com/jmoiron/sqlx"
	chclient "github.com/mainflux/callhome/pkg/client"
	"github.com/mainflux/mainflux"
	"github.com/mainflux/mainflux/consumers"
	"github.com/mainflux/mainflux/consumers/notifiers"
	"github.com/mainflux/mainflux/consumers/notifiers/api"
	notifierpg "github.com/mainflux/mainflux/consumers/notifiers/postgres"
	"github.com/mainflux/mainflux/consumers/notifiers/smtp"
	"github.com/mainflux/mainflux/consumers/notifiers/tracing"
	"github.com/mainflux/mainflux/internal"
	authclient "github.com/mainflux/mainflux/internal/clients/grpc/auth"
	jaegerclient "github.com/mainflux/mainflux/internal/clients/jaeger"
	pgclient "github.com/mainflux/mainflux/internal/clients/postgres"
	"github.com/mainflux/mainflux/internal/email"
	"github.com/mainflux/mainflux/internal/env"
	"github.com/mainflux/mainflux/internal/server"
	httpserver "github.com/mainflux/mainflux/internal/server/http"
	mflog "github.com/mainflux/mainflux/logger"
	"github.com/mainflux/mainflux/pkg/messaging/brokers"
	brokerstracing "github.com/mainflux/mainflux/pkg/messaging/brokers/tracing"
	"github.com/mainflux/mainflux/pkg/ulid"
	"github.com/mainflux/mainflux/pkg/uuid"
	"go.opentelemetry.io/otel/trace"
	"golang.org/x/sync/errgroup"
)

const (
	svcName        = "smtp-notifier"
	envPrefixDB    = "MF_SMTP_NOTIFIER_DB_"
	envPrefixHTTP  = "MF_SMTP_NOTIFIER_HTTP_"
	defDB          = "subscriptions"
	defSvcHTTPPort = "9015"
)

type config struct {
	LogLevel      string  `env:"MF_SMTP_NOTIFIER_LOG_LEVEL"    envDefault:"info"`
	ConfigPath    string  `env:"MF_SMTP_NOTIFIER_CONFIG_PATH"  envDefault:"/config.toml"`
	From          string  `env:"MF_SMTP_NOTIFIER_FROM_ADDR"    envDefault:""`
	BrokerURL     string  `env:"MF_MESSAGE_BROKER_URL"         envDefault:"nats://localhost:4222"`
	JaegerURL     string  `env:"MF_JAEGER_URL"                 envDefault:"http://jaeger:14268/api/traces"`
	SendTelemetry bool    `env:"MF_SEND_TELEMETRY"             envDefault:"true"`
	InstanceID    string  `env:"MF_SMTP_NOTIFIER_INSTANCE_ID"  envDefault:""`
	TraceRatio    float64 `env:"MF_JAEGER_TRACE_RATIO"         envDefault:"1.0"`
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

	var exitCode int
	defer mflog.ExitWithError(&exitCode)

	if cfg.InstanceID == "" {
		if cfg.InstanceID, err = uuid.New().ID(); err != nil {
			logger.Error(fmt.Sprintf("failed to generate instanceID: %s", err))
			exitCode = 1
			return
		}
	}

	dbConfig := pgclient.Config{Name: defDB}
	db, err := pgclient.SetupWithConfig(envPrefixDB, *notifierpg.Migration(), dbConfig)
	if err != nil {
		logger.Fatal(err.Error())
		exitCode = 1
		return
	}
	defer db.Close()

	ec := email.Config{}
	if err := env.Parse(&ec); err != nil {
		logger.Error(fmt.Sprintf("failed to load email configuration : %s", err))
		exitCode = 1
		return
	}

	httpServerConfig := server.Config{Port: defSvcHTTPPort}
	if err := env.Parse(&httpServerConfig, env.Options{Prefix: envPrefixHTTP}); err != nil {
		logger.Error(fmt.Sprintf("failed to load %s HTTP server configuration : %s", svcName, err))
		exitCode = 1
		return
	}

	tp, err := jaegerclient.NewProvider(svcName, cfg.JaegerURL, cfg.InstanceID, cfg.TraceRatio)
	if err != nil {
		logger.Error(fmt.Sprintf("failed to init Jaeger: %s", err))
		exitCode = 1
		return
	}
	defer func() {
		if err := tp.Shutdown(ctx); err != nil {
			logger.Error(fmt.Sprintf("Error shutting down tracer provider: %v", err))
		}
	}()
	tracer := tp.Tracer(svcName)

	pubSub, err := brokers.NewPubSub(ctx, cfg.BrokerURL, logger)
	if err != nil {
		logger.Error(fmt.Sprintf("failed to connect to message broker: %s", err))
		exitCode = 1
		return
	}
	defer pubSub.Close()
	pubSub = brokerstracing.NewPubSub(httpServerConfig, tracer, pubSub)

	auth, authHandler, err := authclient.Setup(svcName)
	if err != nil {
		logger.Error(err.Error())
		exitCode = 1
		return
	}
	defer authHandler.Close()

	logger.Info("Successfully connected to auth grpc server " + authHandler.Secure())

	svc, err := newService(db, tracer, auth, cfg, ec, logger)
	if err != nil {
		logger.Error(err.Error())
		exitCode = 1
		return
	}

	if err = consumers.Start(ctx, svcName, pubSub, svc, cfg.ConfigPath, logger); err != nil {
		logger.Error(fmt.Sprintf("failed to create Postgres writer: %s", err))
		exitCode = 1
		return
	}

	hs := httpserver.New(ctx, cancel, svcName, httpServerConfig, api.MakeHandler(svc, logger, cfg.InstanceID), logger)

	if cfg.SendTelemetry {
		chc := chclient.New(svcName, mainflux.Version, logger, cancel)
		go chc.CallHome(ctx)
	}

	g.Go(func() error {
		return hs.Start()
	})

	g.Go(func() error {
		return server.StopSignalHandler(ctx, cancel, logger, svcName, hs)
	})

	if err := g.Wait(); err != nil {
		logger.Error(fmt.Sprintf("SMTP notifier service terminated: %s", err))
	}
}

func newService(db *sqlx.DB, tracer trace.Tracer, auth mainflux.AuthServiceClient, c config, ec email.Config, logger mflog.Logger) (notifiers.Service, error) {
	database := notifierpg.NewDatabase(db, tracer)
	repo := tracing.New(tracer, notifierpg.New(database))
	idp := ulid.New()

	agent, err := email.New(&ec)
	if err != nil {
		return nil, fmt.Errorf("failed to create email agent: %s", err)
	}

	notifier := smtp.New(agent)
	svc := notifiers.New(auth, repo, idp, notifier, c.From)
	svc = api.LoggingMiddleware(svc, logger)
	counter, latency := internal.MakeMetrics("notifier", "smtp")
	svc = api.MetricsMiddleware(svc, counter, latency)

	return svc, nil
}
