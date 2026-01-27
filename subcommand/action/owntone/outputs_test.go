package owntone

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestDisplayOutputsAction_Run(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("/api/outputs", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		// Return outputs with some diversity to test sorting and filtering
		_, _ = w.Write([]byte(`{
			"outputs": [
				{"id": "2", "name": "B Output", "selected": true},
				{"id": "1", "name": "A Output", "selected": true},
				{"id": "3", "name": "C Output", "selected": false}
			]
		}`))
	})

	server := httptest.NewServer(mux)
	defer server.Close()
	client := NewClient(Config{Url: server.URL})
	action := NewDisplayOutputsAction(client)

	got, err := action.Run(context.Background(), "")
	if err != nil {
		t.Fatalf("Run() error = %v", err)
	}

	// Should be sorted by name: A, B. C is excluded because it's not selected.
	if !strings.Contains(got, "A Output") || !strings.Contains(got, "B Output") {
		t.Errorf("Run() result missing expected outputs: %s", got)
	}

	// Check order (A before B)
	aIdx := strings.Index(got, "A Output")
	bIdx := strings.Index(got, "B Output")
	if aIdx > bIdx {
		t.Errorf("Outputs are not sorted by name: A should be before B")
	}
}

func TestDisplayOutputsAction_Run_OnlySelected(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("/api/outputs", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{
			"outputs": [
				{"id": "1", "name": "Selected", "selected": true},
				{"id": "2", "name": "Not Selected", "selected": false}
			]
		}`))
	})

	server := httptest.NewServer(mux)
	defer server.Close()
	client := NewClient(Config{Url: server.URL})
	action := NewDisplayOutputsAction(client, true)

	got, err := action.Run(context.Background(), "")
	if err != nil {
		t.Fatalf("Run() error = %v", err)
	}

	if !strings.Contains(got, "Selected") {
		t.Errorf("Run() result missing selected output: %s", got)
	}
	if strings.Contains(got, "Not Selected") {
		t.Errorf("Run() result should not contain non-selected output: %s", got)
	}
}
