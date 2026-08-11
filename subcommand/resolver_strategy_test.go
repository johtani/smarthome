package subcommand

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/johtani/smarthome/subcommand/action/llm"
)

func TestNaturalLanguageResolvers(t *testing.T) {
	t.Run("legacy mode without llm endpoint returns none", func(t *testing.T) {
		config := Config{
			Resolver: ResolverConfig{Mode: ResolverModeLegacy},
			LLM:      llm.Config{},
		}
		resolvers := naturalLanguageResolvers(config)
		if len(resolvers) != 0 {
			t.Fatalf("expected 0 resolvers, got %d", len(resolvers))
		}
	})

	t.Run("legacy mode with llm endpoint returns llm resolver", func(t *testing.T) {
		config := Config{
			Resolver: ResolverConfig{Mode: ResolverModeLegacy},
			LLM:      llm.Config{Endpoint: "http://llm.local"},
		}
		resolvers := naturalLanguageResolvers(config)
		if len(resolvers) != 1 {
			t.Fatalf("expected 1 resolver, got %d", len(resolvers))
		}
		if got := resolvers[0].Path(); got != "llm" {
			t.Fatalf("expected llm resolver, got %q", got)
		}
	})

	t.Run("dspy mode without llm endpoint returns dspy only", func(t *testing.T) {
		config := Config{
			Resolver: ResolverConfig{
				Mode:               ResolverModeDSPy,
				DSPyEndpoint:       "http://dspy.local/resolve",
				DSPyTimeoutSeconds: 5,
			},
			LLM: llm.Config{},
		}
		resolvers := naturalLanguageResolvers(config)
		if len(resolvers) != 1 {
			t.Fatalf("expected 1 resolver, got %d", len(resolvers))
		}
		if got := resolvers[0].Path(); got != "dspy" {
			t.Fatalf("expected dspy resolver, got %q", got)
		}
	})

	t.Run("dspy mode with llm endpoint returns dspy then llm", func(t *testing.T) {
		config := Config{
			Resolver: ResolverConfig{
				Mode:               ResolverModeDSPy,
				DSPyEndpoint:       "http://dspy.local/resolve",
				DSPyTimeoutSeconds: 5,
			},
			LLM: llm.Config{Endpoint: "http://llm.local"},
		}
		resolvers := naturalLanguageResolvers(config)
		if len(resolvers) != 2 {
			t.Fatalf("expected 2 resolvers, got %d", len(resolvers))
		}
		if got := resolvers[0].Path(); got != "dspy" {
			t.Fatalf("expected first resolver dspy, got %q", got)
		}
		if got := resolvers[1].Path(); got != "llm" {
			t.Fatalf("expected second resolver llm, got %q", got)
		}
	})
}

func TestDSPyResolverMetadata(t *testing.T) {
	t.Run("uses structured metadata from direct resolver response", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "application/json")
			_, _ = fmt.Fprint(w, `{"command":"start ps5","args":"","thought":"explicit","model":"lfm","prompt_version":"prompt-v2","artifact_version":"artifact-v3","dataset_version":"dataset-v4"}`)
		}))
		defer server.Close()

		resolved, err := newDSPyResolver(server.URL, time.Second).Resolve(
			t.Context(),
			"PS5やるぞ",
			"  start ps5 : start PS5",
			"request-prompt",
		)
		if err != nil {
			t.Fatalf("Resolve failed: %v", err)
		}
		if resolved.Model != "lfm" || resolved.PromptVersion != "prompt-v2" {
			t.Fatalf("unexpected model metadata: %+v", resolved)
		}
		if resolved.ArtifactVersion != "artifact-v3" || resolved.DatasetVersion != "dataset-v4" {
			t.Fatalf("unexpected version metadata: %+v", resolved)
		}
	})

	t.Run("does not trust metadata generated inside compatibility response", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "application/json")
			_, _ = fmt.Fprint(w, `{"choices":[{"message":{"content":"{\"command\":\"start ps5\",\"args\":\"\",\"thought\":\"explicit\",\"model\":\"untrusted\",\"artifact_version\":\"untrusted\"}"}}]}`)
		}))
		defer server.Close()

		resolved, err := newDSPyResolver(server.URL, time.Second).Resolve(
			t.Context(),
			"PS5やるぞ",
			"  start ps5 : start PS5",
			"request-prompt",
		)
		if err != nil {
			t.Fatalf("Resolve failed: %v", err)
		}
		if resolved.Model != "" || resolved.ArtifactVersion != "" || resolved.DatasetVersion != "" {
			t.Fatalf("untrusted metadata was retained: %+v", resolved)
		}
		if resolved.PromptVersion != "request-prompt" {
			t.Fatalf("prompt version = %q, want request-prompt", resolved.PromptVersion)
		}
	})
}
