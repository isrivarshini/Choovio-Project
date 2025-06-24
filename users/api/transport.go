// Copyright (c) Abstract Machines
// SPDX-License-Identifier: Apache-2.0

package api

import (
	"log/slog"
	"net/http"

	"github.com/absmach/magistrala"
	"github.com/absmach/magistrala/pkg/groups"
	"github.com/absmach/magistrala/pkg/oauth2"
	"github.com/absmach/magistrala/users"
	"github.com/go-chi/chi/v5"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

// MakeHandler returns a HTTP handler for Users and Groups API endpoints.
func MakeHandler(cls users.Service, grps groups.Service, mux *chi.Mux, logger *slog.Logger, instanceID string, providers ...oauth2.Provider) http.Handler {
	clientsHandler(cls, mux, logger, providers...)
	groupsHandler(grps, mux, logger)

	mux.Get("/health", magistrala.Health("users", instanceID))
	mux.Handle("/metrics", promhttp.Handler())

	return mux
}
