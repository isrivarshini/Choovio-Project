// Copyright (c) Mainflux
// SPDX-License-Identifier: Apache-2.0

//go:build !test

package api

import (
	"context"
	"fmt"
	"time"

	"github.com/mainflux/mainflux/consumers"
	mflog "github.com/mainflux/mainflux/logger"
)

var _ consumers.BlockingConsumer = (*loggingMiddleware)(nil)

type loggingMiddleware struct {
	logger   mflog.Logger
	consumer consumers.BlockingConsumer
}

// LoggingMiddleware adds logging facilities to the adapter.
func LoggingMiddleware(consumer consumers.BlockingConsumer, logger mflog.Logger) consumers.BlockingConsumer {
	return &loggingMiddleware{
		logger:   logger,
		consumer: consumer,
	}
}

// ConsumeBlocking logs the consume request. It logs the time it took to complete the request.
// If the request fails, it logs the error.
func (lm *loggingMiddleware) ConsumeBlocking(ctx context.Context, msgs interface{}) (err error) {
	defer func(begin time.Time) {
		message := fmt.Sprintf("Method consume took %s to complete", time.Since(begin))
		if err != nil {
			lm.logger.Warn(fmt.Sprintf("%s with error: %s.", message, err))
			return
		}
		lm.logger.Info(fmt.Sprintf("%s without errors.", message))
	}(time.Now())

	return lm.consumer.ConsumeBlocking(ctx, msgs)
}
