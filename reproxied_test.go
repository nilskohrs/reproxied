package reproxied_test

import (
	"context"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/nilskohrs/reproxied"
)

type ClientMock struct {
	executedRequest []*http.Request
}

func (mock *ClientMock) Do(req *http.Request) (*http.Response, error) {
	mock.executedRequest = append(mock.executedRequest, req)
	return &http.Response{Body: ioutil.NopCloser(strings.NewReader("")), StatusCode: 200}, nil
}

func TestShouldChangeHost(t *testing.T) {
	clientMock := &ClientMock{}
	cfg := reproxied.CreateConfig()
	cfg.Proxy = "http://proxy:3128"
	cfg.TargetHost = "https://target.com"
	ctx := context.Background()
	next := http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {})

	handler, err := reproxied.NewWithClient(ctx, next, cfg, "reProxied", clientMock)
	if err != nil {
		t.Fatal(err)
	}

	recorder := httptest.NewRecorder()

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, "https://internal.url/", nil)
	if err != nil {
		t.Fatal(err)
	}

	handler.ServeHTTP(recorder, req)

	if clientMock.executedRequest[0].Host != "target.com" {
		t.Errorf("expected request host to be updated to \"target.com\" but was actually: %v", clientMock.executedRequest[0].Host)
	}
}

func TestShouldKeepHostHeader(t *testing.T) {
	clientMock := &ClientMock{}

	cfg := reproxied.CreateConfig()
	cfg.Proxy = "http://proxy:3128"
	cfg.TargetHost = "https://target.com"
	cfg.KeepHostHeader = true

	ctx := context.Background()
	next := http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {})

	handler, err := reproxied.NewWithClient(ctx, next, cfg, "reProxied", clientMock)
	if err != nil {
		t.Fatal(err)
	}

	recorder := httptest.NewRecorder()

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, "https://internal.url/", nil)
	if err != nil {
		t.Fatal(err)
	}

	handler.ServeHTTP(recorder, req)

	if clientMock.executedRequest[0].URL.Host != "target.com" {
		t.Errorf("expected request host to be updated to \"target.com\" but was actually: %v", clientMock.executedRequest[0].URL.Host)
	}
	if clientMock.executedRequest[0].Host != "internal.url" {
		t.Errorf("expected request header host to be kept to \"internal.url\" but was actually: %v", clientMock.executedRequest[0].Host)
	}
}
