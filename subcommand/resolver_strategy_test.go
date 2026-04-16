package subcommand

import (
	"testing"

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
