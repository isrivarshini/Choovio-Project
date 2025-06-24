// Copyright (c) Mainflux
// SPDX-License-Identifier: Apache-2.0

// Package main contains lora main function to start the lora service.
package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"time"

	mqttPaho "github.com/eclipse/paho.mqtt.golang"
	r "github.com/go-redis/redis/v8"
	chclient "github.com/mainflux/callhome/pkg/client"
	"github.com/mainflux/mainflux"
	"github.com/mainflux/mainflux/internal"
	"github.com/mainflux/mainflux/internal/env"
	"github.com/mainflux/mainflux/internal/server"
	httpserver "github.com/mainflux/mainflux/internal/server/http"
	mflog "github.com/mainflux/mainflux/logger"
	"github.com/mainflux/mainflux/lora"
	"github.com/mainflux/mainflux/lora/api"
	"github.com/mainflux/mainflux/lora/mqtt"
	"github.com/mainflux/mainflux/pkg/messaging"
	"github.com/mainflux/mainflux/pkg/messaging/brokers"
	"github.com/mainflux/mainflux/pkg/messaging/tracing"
	"github.com/mainflux/mainflux/pkg/uuid"
	"golang.org/x/sync/errgroup"

	jaegerClient "github.com/mainflux/mainflux/internal/clients/jaeger"
	redisClient "github.com/mainflux/mainflux/internal/clients/redis"
	"github.com/mainflux/mainflux/lora/redis"
)

const (
	svcName           = "lora-adapter"
	envPrefix         = "MF_LORA_ADAPTER_"
	envPrefixHttp     = "MF_LORA_ADAPTER_HTTP_"
	envPrefixRouteMap = "MF_LORA_ADAPTER_ROUTE_MAP_"
	envPrefixThingsES = "MF_THINGS_ES_"
	defSvcHttpPort    = "9017"

	thingsRMPrefix   = "thing"
	channelsRMPrefix = "channel"
	connsRMPrefix    = "connection"
)

type config struct {
	LogLevel       string        `env:"MF_BOOTSTRAP_LOG_LEVEL"              envDefault:"info"`
	LoraMsgURL     string        `env:"MF_LORA_ADAPTER_MESSAGES_URL"        envDefault:"tcp://localhost:1883"`
	LoraMsgUser    string        `env:"MF_LORA_ADAPTER_MESSAGES_USER"       envDefault:""`
	LoraMsgPass    string        `env:"MF_LORA_ADAPTER_MESSAGES_PASS"       envDefault:""`
	LoraMsgTopic   string        `env:"MF_LORA_ADAPTER_MESSAGES_TOPIC"      envDefault:"application/+/device/+/event/up"`
	LoraMsgTimeout time.Duration `env:"MF_LORA_ADAPTER_MESSAGES_TIMEOUT"    envDefault:"30s"`
	ESConsumerName string        `env:"MF_LORA_ADAPTER_EVENT_CONSUMER"      envDefault:"lora"`
	BrokerURL      string        `env:"MF_BROKER_URL"                       envDefault:"nats://localhost:4222"`
	JaegerURL      string        `env:"MF_JAEGER_URL"                       envDefault:"localhost:6831"`
	SendTelemetry  bool          `env:"MF_SEND_TELEMETRY"                   envDefault:"true"`
	InstanceID     string        `env:"MF_LORA_ADAPTER_INSTANCE_ID"         envDefault:""`
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

	instanceID := cfg.InstanceID
	if instanceID == "" {
		instanceID, err = uuid.New().ID()
		if err != nil {
			log.Fatalf("Failed to generate instanceID: %s", err)
		}
	}

	rmConn, err := redisClient.Setup(envPrefixRouteMap)
	if err != nil {
		logger.Fatal(fmt.Sprintf("failed to setup route map redis client : %s", err))
	}
	var exitCode int
	defer mflog.ExitWithError(&exitCode)
	defer rmConn.Close()

	tp, err := jaegerClient.NewProvider(svcName, cfg.JaegerURL, instanceID)
	if err != nil {
		logger.Error(fmt.Sprintf("Failed to init Jaeger: %s", err))
	}
	defer func() {
		if err := tp.Shutdown(ctx); err != nil {
			logger.Error(fmt.Sprintf("Error shutting down tracer provider: %v", err))
		}
	}()
	tracer := tp.Tracer(svcName)

	pub, err := brokers.NewPublisher(cfg.BrokerURL)
	if err != nil {
		logger.Error(fmt.Sprintf("failed to connect to message broker: %s", err))
		exitCode = 1
		return
	}
	pub = tracing.New(tracer, pub)
	defer pub.Close()

	svc := newService(pub, rmConn, thingsRMPrefix, channelsRMPrefix, connsRMPrefix, logger)

	esConn, err := redisClient.Setup(envPrefixThingsES)
	if err != nil {
		logger.Error(fmt.Sprintf("failed to setup things event store redis client : %s", err))
		exitCode = 1
		return
	}
	defer esConn.Close()

	mqttConn, err := connectToMQTTBroker(cfg.LoraMsgURL, cfg.LoraMsgUser, cfg.LoraMsgPass, cfg.LoraMsgTimeout, logger)
	if err != nil {
		logger.Error(err.Error())
		exitCode = 1
		return
	}

	g.Go(func() error {
		return subscribeToLoRaBroker(svc, mqttConn, cfg.LoraMsgTimeout, cfg.LoraMsgTopic, logger)
	})
	go subscribeToThingsES(svc, esConn, cfg.ESConsumerName, logger)

	httpServerConfig := server.Config{Port: defSvcHttpPort}
	if err := env.Parse(&httpServerConfig, env.Options{Prefix: envPrefixHttp, AltPrefix: envPrefix}); err != nil {
		logger.Error(fmt.Sprintf("failed to load %s HTTP server configuration : %s", svcName, err))
		exitCode = 1
		return
	}
	hs := httpserver.New(ctx, cancel, svcName, httpServerConfig, api.MakeHandler(instanceID), logger)

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
		logger.Error(fmt.Sprintf("LoRa adapter terminated: %s", err))
	}
}

func connectToMQTTBroker(url, user, password string, timeout time.Duration, logger mflog.Logger) (mqttPaho.Client, error) {
	opts := mqttPaho.NewClientOptions()
	opts.AddBroker(url)
	opts.SetUsername(user)
	opts.SetPassword(password)
	opts.SetOnConnectHandler(func(c mqttPaho.Client) {
		logger.Info("Connected to Lora MQTT broker")
	})
	opts.SetConnectionLostHandler(func(c mqttPaho.Client, err error) {
		logger.Error(fmt.Sprintf("MQTT connection lost: %s", err))
	})

	client := mqttPaho.NewClient(opts)

	if token := client.Connect(); token.WaitTimeout(timeout) && token.Error() != nil {
		return nil, fmt.Errorf("failed to connect to Lora MQTT broker: %s", token.Error())
	}

	return client, nil
}

func subscribeToLoRaBroker(svc lora.Service, mc mqttPaho.Client, timeout time.Duration, topic string, logger mflog.Logger) error {
	mqtt := mqtt.NewBroker(svc, mc, timeout, logger)
	logger.Info("Subscribed to Lora MQTT broker")
	if err := mqtt.Subscribe(topic); err != nil {
		return fmt.Errorf("failed to subscribe to Lora MQTT broker: %s", err)
	}
	return nil
}

func subscribeToThingsES(svc lora.Service, client *r.Client, consumer string, logger mflog.Logger) {
	eventStore := redis.NewEventStore(svc, client, consumer, logger)
	logger.Info("Subscribed to Redis Event Store")
	if err := eventStore.Subscribe(context.Background(), "mainflux.things"); err != nil {
		logger.Warn(fmt.Sprintf("Lora-adapter service failed to subscribe to Redis event source: %s", err))
	}
}

func newRouteMapRepository(client *r.Client, prefix string, logger mflog.Logger) lora.RouteMapRepository {
	logger.Info(fmt.Sprintf("Connected to %s Redis Route-map", prefix))
	return redis.NewRouteMapRepository(client, prefix)
}

func newService(pub messaging.Publisher, rmConn *r.Client, thingsRMPrefix, channelsRMPrefix, connsRMPrefix string, logger mflog.Logger) lora.Service {
	thingsRM := newRouteMapRepository(rmConn, thingsRMPrefix, logger)
	chansRM := newRouteMapRepository(rmConn, channelsRMPrefix, logger)
	connsRM := newRouteMapRepository(rmConn, connsRMPrefix, logger)

	svc := lora.New(pub, thingsRM, chansRM, connsRM)
	svc = api.LoggingMiddleware(svc, logger)
	counter, latency := internal.MakeMetrics("lora_adapter", "api")
	svc = api.MetricsMiddleware(svc, counter, latency)

	return svc
}
