// Copyright (c) Mainflux
// SPDX-License-Identifier: Apache-2.0

package certs_test

import (
	"context"
	"fmt"
	"net/http/httptest"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/mainflux/mainflux"
	bsmocks "github.com/mainflux/mainflux/bootstrap/mocks"
	"github.com/mainflux/mainflux/certs"
	"github.com/mainflux/mainflux/certs/mocks"
	"github.com/mainflux/mainflux/logger"
	"github.com/mainflux/mainflux/pkg/errors"
	mfsdk "github.com/mainflux/mainflux/pkg/sdk/go"
	"github.com/mainflux/mainflux/things"
	httpapi "github.com/mainflux/mainflux/things/api/things/http"
	thmocks "github.com/mainflux/mainflux/things/mocks"
	"github.com/opentracing/opentracing-go/mocktracer"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const (
	wrongValue = "wrong-value"
	email      = "user@example.com"
	token      = "token"
	thingsNum  = 1
	thingKey   = "thingKey"
	thingID    = "1"
	ttl        = "1h"
	certNum    = 10

	cfgAuthTimeout = "1s"

	caPath            = "../docker/ssl/certs/ca.crt"
	caKeyPath         = "../docker/ssl/certs/ca.key"
	cfgSignHoursValid = "24h"
)

func newService(tokens map[string]string) (certs.Service, error) {
	ac := bsmocks.NewAuthClient(map[string]string{token: email})
	server := newThingsServer(newThingsService(ac))

	policies := []thmocks.MockSubjectSet{{Object: "users", Relation: "member"}}
	auth := thmocks.NewAuthService(tokens, map[string][]thmocks.MockSubjectSet{email: policies})
	config := mfsdk.Config{
		ThingsURL: server.URL,
	}

	sdk := mfsdk.NewSDK(config)
	repo := mocks.NewCertsRepository()

	tlsCert, caCert, err := certs.LoadCertificates(caPath, caKeyPath)
	if err != nil {
		return nil, err
	}

	authTimeout, err := time.ParseDuration(cfgAuthTimeout)
	if err != nil {
		return nil, err
	}

	pki := mocks.NewPkiAgent(tlsCert, caCert, cfgSignHoursValid, authTimeout)

	return certs.New(auth, repo, sdk, pki), nil
}

func newThingsService(auth mainflux.AuthServiceClient) things.Service {
	ths := make(map[string]things.Thing, thingsNum)
	for i := 0; i < thingsNum; i++ {
		id := strconv.Itoa(i + 1)
		ths[id] = things.Thing{
			ID:    id,
			Key:   thingKey,
			Owner: email,
		}
	}

	return bsmocks.NewThingsService(ths, map[string]things.Channel{}, auth)
}

func TestIssueCert(t *testing.T) {
	svc, err := newService(map[string]string{token: email})
	require.Nil(t, err, fmt.Sprintf("unexpected service creation error: %s\n", err))

	cases := []struct {
		token   string
		desc    string
		thingID string
		ttl     string
		key     string
		err     error
	}{
		{
			desc:    "issue new cert",
			token:   token,
			thingID: thingID,
			ttl:     ttl,
			err:     nil,
		},
		{
			desc:    "issue new cert for non existing thing id",
			token:   token,
			thingID: "2",
			ttl:     ttl,
			err:     certs.ErrFailedCertCreation,
		},
		{
			desc:    "issue new cert for non existing thing id",
			token:   wrongValue,
			thingID: thingID,
			ttl:     ttl,
			err:     errors.ErrAuthentication,
		},
	}

	for _, tc := range cases {
		c, err := svc.IssueCert(context.Background(), tc.token, tc.thingID, tc.ttl)
		assert.True(t, errors.Contains(err, tc.err), fmt.Sprintf("%s: expected %s got %s\n", tc.desc, tc.err, err))
		cert, _ := certs.ReadCert([]byte(c.ClientCert))
		if cert != nil {
			assert.True(t, strings.Contains(cert.Subject.CommonName, thingKey), fmt.Sprintf("%s: expected %s got %s\n", tc.desc, tc.err, err))
		}
	}

}

func TestRevokeCert(t *testing.T) {
	svc, err := newService(map[string]string{token: email})
	require.Nil(t, err, fmt.Sprintf("unexpected service creation error: %s\n", err))

	_, err = svc.IssueCert(context.Background(), token, thingID, ttl)
	require.Nil(t, err, fmt.Sprintf("unexpected service creation error: %s\n", err))

	cases := []struct {
		token   string
		desc    string
		thingID string
		err     error
	}{
		{
			desc:    "revoke cert",
			token:   token,
			thingID: thingID,
			err:     nil,
		},
		{
			desc:    "revoke cert for invalid token",
			token:   wrongValue,
			thingID: thingID,
			err:     errors.ErrAuthentication,
		},
		{
			desc:    "revoke cert for invalid thing id",
			token:   token,
			thingID: "2",
			err:     certs.ErrFailedCertRevocation,
		},
	}

	for _, tc := range cases {
		_, err := svc.RevokeCert(context.Background(), tc.token, tc.thingID)
		assert.True(t, errors.Contains(err, tc.err), fmt.Sprintf("%s: expected %s got %s\n", tc.desc, tc.err, err))
	}

}

func TestListCerts(t *testing.T) {
	svc, err := newService(map[string]string{token: email})
	require.Nil(t, err, fmt.Sprintf("unexpected service creation error: %s\n", err))

	for i := 0; i < certNum; i++ {
		_, err = svc.IssueCert(context.Background(), token, thingID, ttl)
		require.Nil(t, err, fmt.Sprintf("unexpected cert creation error: %s\n", err))
	}

	cases := []struct {
		token   string
		desc    string
		thingID string
		offset  uint64
		limit   uint64
		size    uint64
		err     error
	}{
		{
			desc:    "list all certs with valid token",
			token:   token,
			thingID: thingID,
			offset:  0,
			limit:   certNum,
			size:    certNum,
			err:     nil,
		},
		{
			desc:    "list all certs with invalid token",
			token:   wrongValue,
			thingID: thingID,
			offset:  0,
			limit:   certNum,
			size:    0,
			err:     errors.ErrAuthentication,
		},
		{
			desc:    "list half certs with valid token",
			token:   token,
			thingID: thingID,
			offset:  certNum / 2,
			limit:   certNum,
			size:    certNum / 2,
			err:     nil,
		},
		{
			desc:    "list last cert with valid token",
			token:   token,
			thingID: thingID,
			offset:  certNum - 1,
			limit:   certNum,
			size:    1,
			err:     nil,
		},
	}

	for _, tc := range cases {
		page, err := svc.ListCerts(context.Background(), tc.token, tc.thingID, tc.offset, tc.limit)
		size := uint64(len(page.Certs))
		assert.Equal(t, tc.size, size, fmt.Sprintf("%s: expected %d got %d\n", tc.desc, tc.size, size))
		assert.True(t, errors.Contains(err, tc.err), fmt.Sprintf("%s: expected %s got %s\n", tc.desc, tc.err, err))
	}
}

func TestListSerials(t *testing.T) {
	svc, err := newService(map[string]string{token: email})
	require.Nil(t, err, fmt.Sprintf("unexpected service creation error: %s\n", err))

	var issuedCerts []certs.Cert
	for i := 0; i < certNum; i++ {
		cert, err := svc.IssueCert(context.Background(), token, thingID, ttl)
		assert.Nil(t, err, fmt.Sprintf("unexpected cert creation error: %s\n", err))

		crt := certs.Cert{
			OwnerID: cert.OwnerID,
			ThingID: cert.ThingID,
			Serial:  cert.Serial,
			Expire:  cert.Expire,
		}
		issuedCerts = append(issuedCerts, crt)
	}

	cases := []struct {
		token   string
		desc    string
		thingID string
		offset  uint64
		limit   uint64
		certs   []certs.Cert
		err     error
	}{
		{
			desc:    "list all certs with valid token",
			token:   token,
			thingID: thingID,
			offset:  0,
			limit:   certNum,
			certs:   issuedCerts,
			err:     nil,
		},
		{
			desc:    "list all certs with invalid token",
			token:   wrongValue,
			thingID: thingID,
			offset:  0,
			limit:   certNum,
			certs:   nil,
			err:     errors.ErrAuthentication,
		},
		{
			desc:    "list half certs with valid token",
			token:   token,
			thingID: thingID,
			offset:  certNum / 2,
			limit:   certNum,
			certs:   issuedCerts[certNum/2:],
			err:     nil,
		},
		{
			desc:    "list last cert with valid token",
			token:   token,
			thingID: thingID,
			offset:  certNum - 1,
			limit:   certNum,
			certs:   []certs.Cert{issuedCerts[certNum-1]},
			err:     nil,
		},
	}

	for _, tc := range cases {
		page, err := svc.ListSerials(context.Background(), tc.token, tc.thingID, tc.offset, tc.limit)
		assert.Equal(t, tc.certs, page.Certs, fmt.Sprintf("%s: expected %v got %v\n", tc.desc, tc.certs, page.Certs))
		assert.True(t, errors.Contains(err, tc.err), fmt.Sprintf("%s: expected %s got %s\n", tc.desc, tc.err, err))
	}
}

func TestViewCert(t *testing.T) {
	svc, err := newService(map[string]string{token: email})
	require.Nil(t, err, fmt.Sprintf("unexpected service creation error: %s\n", err))

	ic, err := svc.IssueCert(context.Background(), token, thingID, ttl)
	require.Nil(t, err, fmt.Sprintf("unexpected cert creation error: %s\n", err))

	cert := certs.Cert{
		ThingID:    thingID,
		ClientCert: ic.ClientCert,
		Serial:     ic.Serial,
		Expire:     ic.Expire,
	}

	cases := []struct {
		token    string
		desc     string
		serialID string
		cert     certs.Cert
		err      error
	}{
		{
			desc:     "list cert with valid token and serial",
			token:    token,
			serialID: cert.Serial,
			cert:     cert,
			err:      nil,
		},
		{
			desc:     "list cert with invalid token",
			token:    wrongValue,
			serialID: cert.Serial,
			cert:     certs.Cert{},
			err:      errors.ErrAuthentication,
		},
		{
			desc:     "list cert with invalid serial",
			token:    token,
			serialID: wrongValue,
			cert:     certs.Cert{},
			err:      errors.ErrNotFound,
		},
	}

	for _, tc := range cases {
		cert, err := svc.ViewCert(context.Background(), tc.token, tc.serialID)
		assert.Equal(t, tc.cert, cert, fmt.Sprintf("%s: expected %v got %v\n", tc.desc, tc.cert, cert))
		assert.True(t, errors.Contains(err, tc.err), fmt.Sprintf("%s: expected %s got %s\n", tc.desc, tc.err, err))
	}
}

func newThingsServer(svc things.Service) *httptest.Server {
	logger := logger.NewMock()
	mux := httpapi.MakeHandler(mocktracer.New(), svc, logger)
	return httptest.NewServer(mux)
}
