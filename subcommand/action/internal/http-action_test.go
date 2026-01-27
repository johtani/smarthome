package internal

import (
	"context"
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
