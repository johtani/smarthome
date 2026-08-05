package internal

import (
	"context"
	"errors"
	"io"
	"net/http"
	"strings"
	"testing"
)

func TestBuildHttpRequestWithParams(t *testing.T) {
	ctx := context.Background()
	method := "GET"
	url := "http://example.com"
	params := map[string]string{
		"foo": "bar",
		"baz": "qux",
	}

	req, err := BuildHttpRequestWithParams(ctx, method, url, params)
	if err != nil {
		t.Fatalf("BuildHttpRequestWithParams() error = %v", err)
	}

	if req.Method != method {
		t.Errorf("Method got = %v, want %v", req.Method, method)
	}

	if req.URL.Host != "example.com" {
		t.Errorf("Host got = %v, want %v", req.URL.Host, "example.com")
	}

	q := req.URL.Query()
	for k, v := range params {
		if q.Get(k) != v {
			t.Errorf("Query param %s got = %v, want %v", k, q.Get(k), v)
		}
	}
}

func TestHandleResponse(t *testing.T) {
	tests := []struct {
		name             string
		statusCode       int
		expectedStatuses []int
		wantErr          bool
	}{
		{"default success (200)", http.StatusOK, nil, false},
		{"default failure (400)", http.StatusBadRequest, nil, true},
		{"custom success (204)", http.StatusNoContent, []int{http.StatusNoContent}, false},
		{"multiple success (200 or 204)", http.StatusNoContent, []int{http.StatusOK, http.StatusNoContent}, false},
		{"custom failure (500)", http.StatusInternalServerError, []int{http.StatusOK}, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			res := &http.Response{
				StatusCode: tt.statusCode,
				Body:       io.NopCloser(strings.NewReader("")),
				Header:     make(http.Header),
			}
			err := HandleResponse(res, tt.expectedStatuses...)
			if (err != nil) != tt.wantErr {
				t.Errorf("HandleResponse() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestHandleResponseCapturesErrorBody(t *testing.T) {
	body := strings.Repeat("x", MaxErrorResponseBodySize+100)
	res := &http.Response{
		StatusCode: http.StatusInternalServerError,
		Body:       io.NopCloser(strings.NewReader(body)),
		Header:     http.Header{"Content-Type": []string{"text/plain"}},
	}

	err := HandleResponse(res, http.StatusOK)
	var statusErr *HTTPStatusError
	if !errors.As(err, &statusErr) {
		t.Fatalf("HandleResponse() error type = %T, want *HTTPStatusError", err)
	}
	if statusErr.StatusCode != http.StatusInternalServerError {
		t.Errorf("StatusCode = %d, want %d", statusErr.StatusCode, http.StatusInternalServerError)
	}
	if statusErr.ContentType != "text/plain" {
		t.Errorf("ContentType = %q, want text/plain", statusErr.ContentType)
	}
	if len(statusErr.ResponseBody) != MaxErrorResponseBodySize {
		t.Errorf("ResponseBody length = %d, want %d", len(statusErr.ResponseBody), MaxErrorResponseBodySize)
	}
	if !statusErr.BodyTruncated {
		t.Error("BodyTruncated = false, want true")
	}
	if strings.Contains(err.Error(), body[:100]) {
		t.Error("error message must not expose the response body")
	}
}

func TestDecodeJSONResponse(t *testing.T) {
	type Dummy struct {
		Foo string `json:"foo"`
	}

	tests := []struct {
		name             string
		statusCode       int
		body             string
		expectedStatuses []int
		wantErr          bool
		wantFoo          string
	}{
		{"success", http.StatusOK, `{"foo":"bar"}`, nil, false, "bar"},
		{"unexpected status", http.StatusBadRequest, `{"foo":"bar"}`, []int{http.StatusOK}, true, ""},
		{"invalid json", http.StatusOK, `{"foo":`, nil, true, ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			res := &http.Response{
				StatusCode: tt.statusCode,
				Body:       io.NopCloser(strings.NewReader(tt.body)),
				Header:     make(http.Header),
			}
			var target Dummy
			err := DecodeJSONResponse(res, &target, tt.expectedStatuses...)
			if (err != nil) != tt.wantErr {
				t.Errorf("DecodeJSONResponse() error = %v, wantErr %v", err, tt.wantErr)
			}
			if !tt.wantErr && target.Foo != tt.wantFoo {
				t.Errorf("DecodeJSONResponse() target.Foo = %v, want %v", target.Foo, tt.wantFoo)
			}
		})
	}
}
