package api_test

import (
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/mainflux/mainflux"
	adapter "github.com/mainflux/mainflux/http"
	"github.com/mainflux/mainflux/http/api"
	"github.com/mainflux/mainflux/http/mocks"
	manager "github.com/mainflux/mainflux/manager/client"
	"github.com/stretchr/testify/assert"
)

const (
	id           = "123e4567-e89b-12d3-a456-000000000001"
	token        = "auth_token"
	invalidToken = "invalid_token"
	msg          = `[{"n":"current","t":-1,"v":1.6}]`
)

func newService() mainflux.MessagePublisher {
	pub := mocks.NewPublisher()
	return adapter.New(pub)
}

func newHTTPServer(pub mainflux.MessagePublisher, mc manager.ManagerClient) *httptest.Server {
	mux := api.MakeHandler(pub, mc)
	return httptest.NewServer(mux)
}

func newManagerServer() *httptest.Server {
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("Authorization") == "invalid_token" {
			w.WriteHeader(http.StatusForbidden)
			return
		}
		w.WriteHeader(http.StatusOK)
	}))
}

func newManagerClient(url string) manager.ManagerClient {
	return manager.NewClient(url)
}

type testRequest struct {
	client      *http.Client
	method      string
	url         string
	contentType string
	token       string
	body        io.Reader
}

func (tr testRequest) make() (*http.Response, error) {
	req, err := http.NewRequest(tr.method, tr.url, tr.body)
	if err != nil {
		return nil, err
	}
	if tr.token != "" {
		req.Header.Set("Authorization", tr.token)
	}
	if tr.contentType != "" {
		req.Header.Set("Content-Type", tr.contentType)
	}
	return tr.client.Do(req)
}

func TestPublish(t *testing.T) {
	mcServer := newManagerServer()
	defer mcServer.Close()
	mc := newManagerClient(mcServer.URL)

	pub := newService()
	ts := newHTTPServer(pub, mc)
	defer ts.Close()
	client := ts.Client()

	cases := map[string]struct {
		chanID      string
		msg         string
		contentType string
		auth        string
		status      int
	}{
		"publish message":                                  {id, msg, "application/senml+json", token, http.StatusAccepted},
		"publish message with no authorization token":      {id, msg, "application/senml+json", "", http.StatusForbidden},
		"publish message with invalid authorization token": {id, msg, "application/senml+json", invalidToken, http.StatusForbidden},
		"publish message with no content type":             {id, msg, "", token, http.StatusAccepted},
		"publish message with invalid channel id":          {"1", msg, "application/senml+json", token, http.StatusNotFound},
	}

	for desc, tc := range cases {
		req := testRequest{
			client:      client,
			method:      http.MethodPost,
			url:         fmt.Sprintf("%s/channels/%s/messages", ts.URL, tc.chanID),
			contentType: tc.contentType,
			token:       tc.auth,
			body:        strings.NewReader(tc.msg),
		}
		res, err := req.make()
		assert.Nil(t, err, fmt.Sprintf("%s: unexpected error %s", desc, err))
		assert.Equal(t, tc.status, res.StatusCode, fmt.Sprintf("%s: expected status code %d got %d", desc, tc.status, res.StatusCode))
	}
}
