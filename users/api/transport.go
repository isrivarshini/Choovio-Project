// Copyright (c) Mainflux
// SPDX-License-Identifier: Apache-2.0

package api

import (
	"context"
	"encoding/json"
	"net/http"
	"strings"

	kitot "github.com/go-kit/kit/tracing/opentracing"
	kithttp "github.com/go-kit/kit/transport/http"
	"github.com/go-zoo/bone"
	"github.com/mainflux/mainflux"
	"github.com/mainflux/mainflux/internal/httputil"
	"github.com/mainflux/mainflux/pkg/errors"
	"github.com/mainflux/mainflux/pkg/uuid"
	"github.com/mainflux/mainflux/users"
	opentracing "github.com/opentracing/opentracing-go"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

const (
	contentType = "application/json"
	offsetKey   = "offset"
	limitKey    = "limit"
	emailKey    = "email"
	metadataKey = "metadata"
	defOffset   = 0
	defLimit    = 10
)

// MakeHandler returns a HTTP handler for API endpoints.
func MakeHandler(svc users.Service, tracer opentracing.Tracer) http.Handler {
	opts := []kithttp.ServerOption{
		kithttp.ServerErrorEncoder(encodeError),
	}

	mux := bone.New()

	mux.Post("/users", kithttp.NewServer(
		kitot.TraceServer(tracer, "register")(registrationEndpoint(svc)),
		decodeCreateUserReq,
		encodeResponse,
		opts...,
	))

	mux.Get("/users/profile", kithttp.NewServer(
		kitot.TraceServer(tracer, "view_profile")(viewProfileEndpoint(svc)),
		decodeViewProfile,
		encodeResponse,
		opts...,
	))

	mux.Get("/users/:userID", kithttp.NewServer(
		kitot.TraceServer(tracer, "view_user")(viewUserEndpoint(svc)),
		decodeViewUser,
		encodeResponse,
		opts...,
	))

	mux.Get("/users", kithttp.NewServer(
		kitot.TraceServer(tracer, "list_users")(listUsersEndpoint(svc)),
		decodeListUsers,
		encodeResponse,
		opts...,
	))

	mux.Put("/users", kithttp.NewServer(
		kitot.TraceServer(tracer, "update_user")(updateUserEndpoint(svc)),
		decodeUpdateUser,
		encodeResponse,
		opts...,
	))

	mux.Post("/password/reset-request", kithttp.NewServer(
		kitot.TraceServer(tracer, "res-req")(passwordResetRequestEndpoint(svc)),
		decodePasswordResetRequest,
		encodeResponse,
		opts...,
	))

	mux.Put("/password/reset", kithttp.NewServer(
		kitot.TraceServer(tracer, "reset")(passwordResetEndpoint(svc)),
		decodePasswordReset,
		encodeResponse,
		opts...,
	))

	mux.Patch("/password", kithttp.NewServer(
		kitot.TraceServer(tracer, "reset")(passwordChangeEndpoint(svc)),
		decodePasswordChange,
		encodeResponse,
		opts...,
	))

	mux.Get("/groups/:groupId", kithttp.NewServer(
		kitot.TraceServer(tracer, "list_members")(listMembersEndpoint(svc)),
		decodeListMembersRequest,
		encodeResponse,
		opts...,
	))

	mux.Post("/tokens", kithttp.NewServer(
		kitot.TraceServer(tracer, "login")(loginEndpoint(svc)),
		decodeCredentials,
		encodeResponse,
		opts...,
	))

	mux.GetFunc("/health", mainflux.Health("users"))
	mux.Handle("/metrics", promhttp.Handler())

	return mux
}

func decodeViewUser(_ context.Context, r *http.Request) (interface{}, error) {
	t, err := httputil.ExtractAuthToken(r)
	if err != nil {
		return nil, err
	}
	req := viewUserReq{
		token:  t,
		userID: bone.GetValue(r, "userID"),
	}
	return req, nil
}

func decodeViewProfile(_ context.Context, r *http.Request) (interface{}, error) {
	t, err := httputil.ExtractAuthToken(r)
	if err != nil {
		return nil, err
	}
	req := viewUserReq{
		token: t,
	}
	return req, nil
}

func decodeListUsers(_ context.Context, r *http.Request) (interface{}, error) {
	o, err := httputil.ReadUintQuery(r, offsetKey, defOffset)
	if err != nil {
		return nil, err
	}

	l, err := httputil.ReadUintQuery(r, limitKey, defLimit)
	if err != nil {
		return nil, err
	}

	e, err := httputil.ReadStringQuery(r, emailKey, "")
	if err != nil {
		return nil, err
	}

	m, err := httputil.ReadMetadataQuery(r, metadataKey, nil)
	if err != nil {
		return nil, err
	}

	t, err := httputil.ExtractAuthToken(r)
	if err != nil {
		return nil, err
	}
	req := listUsersReq{
		token:    t,
		offset:   o,
		limit:    l,
		email:    e,
		metadata: m,
	}
	return req, nil
}

func decodeUpdateUser(_ context.Context, r *http.Request) (interface{}, error) {
	var req updateUserReq
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		return nil, errors.Wrap(errors.ErrMalformedEntity, err)
	}

	t, err := httputil.ExtractAuthToken(r)
	if err != nil {
		return nil, err
	}
	req.token = t
	return req, nil
}

func decodeCredentials(_ context.Context, r *http.Request) (interface{}, error) {
	if !strings.Contains(r.Header.Get("Content-Type"), contentType) {
		return nil, errors.ErrUnsupportedContentType
	}

	var user users.User
	if err := json.NewDecoder(r.Body).Decode(&user); err != nil {
		return nil, errors.Wrap(errors.ErrMalformedEntity, err)
	}

	return userReq{user}, nil
}

func decodeCreateUserReq(_ context.Context, r *http.Request) (interface{}, error) {
	if !strings.Contains(r.Header.Get("Content-Type"), contentType) {
		return nil, errors.ErrUnsupportedContentType
	}

	t, err := httputil.ExtractAuthToken(r)
	if err != nil {
		return nil, err
	}

	var user users.User
	if err := json.NewDecoder(r.Body).Decode(&user); err != nil {
		return nil, errors.Wrap(errors.ErrMalformedEntity, err)
	}

	req := createUserReq{
		user:  user,
		token: t,
	}

	return req, nil
}

func decodePasswordResetRequest(_ context.Context, r *http.Request) (interface{}, error) {
	if !strings.Contains(r.Header.Get("Content-Type"), contentType) {
		return nil, errors.ErrUnsupportedContentType
	}

	var req passwResetReq

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		return nil, errors.Wrap(errors.ErrMalformedEntity, err)
	}

	req.Host = r.Header.Get("Referer")
	return req, nil
}

func decodePasswordReset(_ context.Context, r *http.Request) (interface{}, error) {
	if !strings.Contains(r.Header.Get("Content-Type"), contentType) {
		return nil, errors.ErrUnsupportedContentType
	}

	var req resetTokenReq
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		return nil, errors.Wrap(errors.ErrMalformedEntity, err)
	}

	return req, nil
}

func decodePasswordChange(_ context.Context, r *http.Request) (interface{}, error) {
	if !strings.Contains(r.Header.Get("Content-Type"), contentType) {
		return nil, errors.ErrUnsupportedContentType
	}

	var req passwChangeReq
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		return nil, errors.Wrap(errors.ErrMalformedEntity, err)
	}

	t, err := httputil.ExtractAuthToken(r)
	if err != nil {
		return nil, err
	}
	req.Token = t

	return req, nil
}

func decodeListMembersRequest(_ context.Context, r *http.Request) (interface{}, error) {
	o, err := httputil.ReadUintQuery(r, offsetKey, defOffset)
	if err != nil {
		return nil, err
	}

	l, err := httputil.ReadUintQuery(r, limitKey, defLimit)
	if err != nil {
		return nil, err
	}

	m, err := httputil.ReadMetadataQuery(r, metadataKey, nil)
	if err != nil {
		return nil, err
	}

	t, err := httputil.ExtractAuthToken(r)
	if err != nil {
		return nil, err
	}
	req := listMemberGroupReq{
		token:    t,
		groupID:  bone.GetValue(r, "groupId"),
		offset:   o,
		limit:    l,
		metadata: m,
	}
	return req, nil
}

func encodeResponse(_ context.Context, w http.ResponseWriter, response interface{}) error {
	if ar, ok := response.(mainflux.Response); ok {
		for k, v := range ar.Headers() {
			w.Header().Set(k, v)
		}
		w.Header().Set("Content-Type", contentType)
		w.WriteHeader(ar.Code())

		if ar.Empty() {
			return nil
		}
	}

	return json.NewEncoder(w).Encode(response)
}

func encodeError(_ context.Context, err error, w http.ResponseWriter) {
	switch {
	case errors.Contains(err, errors.ErrInvalidQueryParams),
		errors.Contains(err, errors.ErrMalformedEntity),
		errors.Contains(err, users.ErrPasswordFormat):
		w.WriteHeader(http.StatusBadRequest)
	case errors.Contains(err, errors.ErrAuthentication),
		errors.Contains(err, users.ErrRecoveryToken):
		w.WriteHeader(http.StatusUnauthorized)
	case errors.Contains(err, errors.ErrAuthorization):
		w.WriteHeader(http.StatusForbidden)
	case errors.Contains(err, errors.ErrConflict),
		errors.Contains(err, errors.ErrConflict):
		w.WriteHeader(http.StatusConflict)
	case errors.Contains(err, errors.ErrUnsupportedContentType):
		w.WriteHeader(http.StatusUnsupportedMediaType)
	case errors.Contains(err, errors.ErrNotFound):
		w.WriteHeader(http.StatusNotFound)

	case errors.Contains(err, uuid.ErrGeneratingID):
		w.WriteHeader(http.StatusInternalServerError)

	case errors.Contains(err, errors.ErrCreateEntity),
		errors.Contains(err, errors.ErrUpdateEntity),
		errors.Contains(err, errors.ErrViewEntity),
		errors.Contains(err, errors.ErrRemoveEntity):
		w.WriteHeader(http.StatusInternalServerError)

	default:
		w.WriteHeader(http.StatusInternalServerError)
	}

	if errorVal, ok := err.(errors.Error); ok {
		w.Header().Set("Content-Type", contentType)
		if err := json.NewEncoder(w).Encode(httputil.ErrorRes{Err: errorVal.Msg()}); err != nil {
			w.WriteHeader(http.StatusInternalServerError)
		}
	}
}
