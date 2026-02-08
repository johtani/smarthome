package internal

import (
	"context"
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
