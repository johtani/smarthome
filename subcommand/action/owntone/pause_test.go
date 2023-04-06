package owntone

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func createMockServer(code int) *httptest.Server {
	return httptest.NewServer(
		http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
			if req.Method != http.MethodPut {
				rw.WriteHeader(http.StatusBadRequest)
				return
			}
			if req.URL.Path != "/api/player/pause" {
				rw.WriteHeader(http.StatusBadRequest)
				return
			}
			rw.WriteHeader(code)
			return
		}))
}

func TestPauseAction_Run(t *testing.T) {
	tests := []struct {
		name    string
		code    int
		wantErr bool
	}{
		{"OK", http.StatusNoContent, false},
		{"NG", http.StatusInternalServerError, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := createMockServer(tt.code)
			defer server.Close()
			t.Setenv(EnvUrl, server.URL)
			a := NewPauseAction()
			if err := a.Run(); (err != nil) != tt.wantErr {
				t.Errorf("Run() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
