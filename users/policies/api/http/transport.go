// Copyright (c) Mainflux
// SPDX-License-Identifier: Apache-2.0

package http

import (
	"context"
	"encoding/json"
	"net/http"
	"strings"

	kithttp "github.com/go-kit/kit/transport/http"
	"github.com/go-zoo/bone"
	"github.com/mainflux/mainflux/internal/api"
	"github.com/mainflux/mainflux/internal/apiutil"
	"github.com/mainflux/mainflux/logger"
	"github.com/mainflux/mainflux/pkg/errors"
	"github.com/mainflux/mainflux/users/policies"
	"go.opentelemetry.io/contrib/instrumentation/github.com/go-kit/kit/otelkit"
)

// MakeHandler returns a HTTP handler for API endpoints.
func MakeHandler(svc policies.Service, mux *bone.Mux, logger logger.Logger) http.Handler {
	opts := []kithttp.ServerOption{
		kithttp.ServerErrorEncoder(apiutil.LoggingErrorEncoder(logger, api.EncodeError)),
	}
	mux.Post("/authorize", kithttp.NewServer(
		otelkit.EndpointMiddleware(otelkit.WithOperation("authorize"))(authorizeEndpoint(svc)),
		decodeAuthorize,
		api.EncodeResponse,
		opts...,
	))

	mux.Post("/policies", kithttp.NewServer(
		otelkit.EndpointMiddleware(otelkit.WithOperation("add_policy"))(createPolicyEndpoint(svc)),
		decodePolicyCreate,
		api.EncodeResponse,
		opts...,
	))

	mux.Put("/policies", kithttp.NewServer(
		otelkit.EndpointMiddleware(otelkit.WithOperation("update_policy"))(updatePolicyEndpoint(svc)),
		decodePolicyUpdate,
		api.EncodeResponse,
		opts...,
	))

	mux.Get("/policies", kithttp.NewServer(
		otelkit.EndpointMiddleware(otelkit.WithOperation("list_policies"))(listPolicyEndpoint(svc)),
		decodeListPoliciesRequest,
		api.EncodeResponse,
		opts...,
	))

	mux.Delete("/policies/:subject/:object", kithttp.NewServer(
		otelkit.EndpointMiddleware(otelkit.WithOperation("delete_policy"))(deletePolicyEndpoint(svc)),
		deletePolicyRequest,
		api.EncodeResponse,
		opts...,
	))

	return mux
}

func decodeAuthorize(_ context.Context, r *http.Request) (interface{}, error) {
	if !strings.Contains(r.Header.Get("Content-Type"), api.ContentType) {
		return nil, errors.ErrUnsupportedContentType
	}

	var authReq authorizeReq
	if err := json.NewDecoder(r.Body).Decode(&authReq); err != nil {
		return nil, errors.Wrap(errors.ErrMalformedEntity, err)
	}

	return authReq, nil
}

func decodePolicyCreate(_ context.Context, r *http.Request) (interface{}, error) {
	if !strings.Contains(r.Header.Get("Content-Type"), api.ContentType) {
		return nil, errors.ErrUnsupportedContentType
	}

	var m policies.Policy
	if err := json.NewDecoder(r.Body).Decode(&m); err != nil {
		return nil, errors.Wrap(errors.ErrMalformedEntity, err)
	}

	req := createPolicyReq{
		token:   apiutil.ExtractBearerToken(r),
		Subject: m.Subject,
		Object:  m.Object,
		Actions: m.Actions,
	}
	return req, nil
}

func decodePolicyUpdate(_ context.Context, r *http.Request) (interface{}, error) {
	if !strings.Contains(r.Header.Get("Content-Type"), api.ContentType) {
		return nil, errors.ErrUnsupportedContentType
	}
	var m policies.Policy
	if err := json.NewDecoder(r.Body).Decode(&m); err != nil {
		return nil, errors.Wrap(errors.ErrMalformedEntity, err)
	}

	req := updatePolicyReq{
		token:   apiutil.ExtractBearerToken(r),
		Subject: m.Subject,
		Object:  m.Object,
		Actions: m.Actions,
	}

	return req, nil
}

func decodeListPoliciesRequest(_ context.Context, r *http.Request) (interface{}, error) {
	total, err := apiutil.ReadNumQuery[uint64](r, api.OffsetKey, api.DefOffset)
	if err != nil {
		return nil, err
	}
	offset, err := apiutil.ReadNumQuery[uint64](r, api.OffsetKey, api.DefOffset)
	if err != nil {
		return nil, err
	}
	limit, err := apiutil.ReadNumQuery[uint64](r, api.LimitKey, api.DefLimit)
	if err != nil {
		return nil, err
	}
	ownerID, err := apiutil.ReadStringQuery(r, api.OwnerKey, "")
	if err != nil {
		return nil, err
	}
	subject, err := apiutil.ReadStringQuery(r, api.SubjectKey, "")
	if err != nil {
		return nil, err
	}
	object, err := apiutil.ReadStringQuery(r, api.ObjectKey, "")
	if err != nil {
		return nil, err
	}
	action, err := apiutil.ReadStringQuery(r, api.ActionKey, "")
	if err != nil {
		return nil, err
	}

	req := listPolicyReq{
		token:   apiutil.ExtractBearerToken(r),
		Total:   total,
		Offset:  offset,
		Limit:   limit,
		OwnerID: ownerID,
		Subject: subject,
		Object:  object,
		Actions: action,
	}
	return req, nil
}

func deletePolicyRequest(_ context.Context, r *http.Request) (interface{}, error) {
	req := deletePolicyReq{
		token:   apiutil.ExtractBearerToken(r),
		Subject: bone.GetValue(r, "subject"),
		Object:  bone.GetValue(r, "object"),
	}

	return req, nil
}
