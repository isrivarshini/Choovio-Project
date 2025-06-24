// Copyright (c) Mainflux
// SPDX-License-Identifier: Apache-2.0

package sdk_test

// import (
// 	"fmt"
// 	"net/http"
// 	"net/http/httptest"
// 	"testing"

// 	adapter "github.com/mainflux/mainflux/http"
// 	"github.com/mainflux/mainflux/http/api"
// 	"github.com/mainflux/mainflux/http/mocks"
// 	"github.com/mainflux/mainflux/internal/apiutil"
// 	"github.com/mainflux/mainflux/pkg/errors"
// 	sdk "github.com/mainflux/mainflux/pkg/sdk/go"
// 	"github.com/mainflux/mainflux/things/policies"
// 	"github.com/stretchr/testify/assert"
// )

// var unexpectedJSONEnd = errors.New("unexpected end of JSON input")

// func newMessageService(cc policies.AuthServiceClient) adapter.Service {
// 	pub := mocks.NewPublisher()

// 	return adapter.New(pub, cc)
// }

// func newMessageServer(svc adapter.Service) *httptest.Server {
// 	mux := api.MakeHandler(svc, instanceID)

// 	return httptest.NewServer(mux)
// }

// func TestSendMessage(t *testing.T) {
// 	chanID := "1"
// 	atoken := "auth_token"
// 	invalidToken := "invalid_token"
// 	msg := `[{"n":"current","t":-1,"v":1.6}]`
// 	thingsClient := mocks.NewThingsClient(map[string]string{atoken: chanID})
// 	pub := newMessageService(thingsClient)
// 	ts := newMessageServer(pub)
// 	defer ts.Close()
// 	sdkConf := sdk.Config{
// 		HTTPAdapterURL:  ts.URL,
// 		MsgContentType:  contentType,
// 		TLSVerification: false,
// 	}

// 	mfsdk := sdk.NewSDK(sdkConf)

// 	cases := map[string]struct {
// 		chanID string
// 		msg    string
// 		auth   string
// 		err    errors.SDKError
// 	}{
// 		"publish message": {
// 			chanID: chanID,
// 			msg:    msg,
// 			auth:   atoken,
// 			err:    nil,
// 		},
// 		"publish message without authorization token": {
// 			chanID: chanID,
// 			msg:    msg,
// 			auth:   "",
// 			err:    errors.NewSDKErrorWithStatus(errors.Wrap(apiutil.ErrValidation, apiutil.ErrBearerKey), http.StatusUnauthorized),
// 		},
// 		"publish message with invalid authorization token": {
// 			chanID: chanID,
// 			msg:    msg,
// 			auth:   invalidToken,
// 			err:    errors.NewSDKErrorWithStatus(unexpectedJSONEnd, http.StatusUnauthorized),
// 		},
// 		"publish message with wrong content type": {
// 			chanID: chanID,
// 			msg:    "text",
// 			auth:   atoken,
// 			err:    nil,
// 		},
// 		"publish message to wrong channel": {
// 			chanID: "",
// 			msg:    msg,
// 			auth:   atoken,
// 			err:    errors.NewSDKErrorWithStatus(errors.Wrap(apiutil.ErrValidation, errors.ErrMalformedEntity), http.StatusBadRequest),
// 		},
// 		"publish message unable to authorize": {
// 			chanID: chanID,
// 			msg:    msg,
// 			auth:   "invalid-token",
// 			err:    errors.NewSDKErrorWithStatus(unexpectedJSONEnd, http.StatusUnauthorized),
// 		},
// 	}
// 	for desc, tc := range cases {
// 		err := mfsdk.SendMessage(tc.chanID, tc.msg, tc.auth)
// 		switch tc.err {
// 		case nil:
// 			assert.Nil(t, err, fmt.Sprintf("%s: got unexpected error: %s", desc, err))
// 		default:
// 			assert.Equal(t, tc.err.Error(), err.Error(), fmt.Sprintf("%s: expected error %s, got %s", desc, tc.err, err))
// 		}
// 	}

// }

// func TestSetContentType(t *testing.T) {
// 	chanID := "1"
// 	atoken := "auth_token"
// 	thingsClient := mocks.NewThingsClient(map[string]string{atoken: chanID})

// 	pub := newMessageService(thingsClient)
// 	ts := newMessageServer(pub)
// 	defer ts.Close()

// 	sdkConf := sdk.Config{
// 		HTTPAdapterURL:  ts.URL,
// 		MsgContentType:  contentType,
// 		TLSVerification: false,
// 	}
// 	mfsdk := sdk.NewSDK(sdkConf)

// 	cases := []struct {
// 		desc  string
// 		cType sdk.ContentType
// 		err   errors.SDKError
// 	}{
// 		{
// 			desc:  "set senml+json content type",
// 			cType: "application/senml+json",
// 			err:   nil,
// 		},
// 		{
// 			desc:  "set invalid content type",
// 			cType: "invalid",
// 			err:   errors.NewSDKError(apiutil.ErrUnsupportedContentType),
// 		},
// 	}
// 	for _, tc := range cases {
// 		err := mfsdk.SetContentType(tc.cType)
// 		assert.Equal(t, tc.err, err, fmt.Sprintf("%s: expected error %s, got %s", tc.desc, tc.err, err))
// 	}
// }
