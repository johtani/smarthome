// Command resolver-compare runs the legacy resolver and guards offline predictions.
package main

import (
	"context"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"os"
	"time"

	"github.com/johtani/smarthome/subcommand"
	"github.com/johtani/smarthome/subcommand/action/llm"
)

type sample struct {
	Text        string
	Predictions map[string]llm.ResolvedCommand
}

type request struct {
	Help, Model, Endpoint, PromptVersion string
	Samples                              []sample
}

type decision struct {
	Raw   llm.ResolvedCommand
	Final llm.ResolvedCommand
}

type result struct {
	Decisions   map[string]decision
	LegacyMS    float64
	LegacyError string `json:",omitempty"`
}

func run() error {
	configDir := flag.String("config-dir", "config", "configuration directory with matching command definitions")
	catalog := flag.Bool("catalog", false, "print production command help")
	flag.Parse()
	commands, err := subcommand.LoadCommandDefinitions(*configDir)
	if err != nil {
		return err
	}
	help := commands.Help()
	if *catalog {
		_, err = fmt.Print(help)
		return err
	}
	var req request
	if err := json.NewDecoder(os.Stdin).Decode(&req); err != nil {
		return err
	}
	if req.Help != help {
		return fmt.Errorf("command catalog changed during comparison")
	}
	client := llm.NewClient(llm.Config{Model: req.Model, Endpoint: req.Endpoint, APIKey: os.Getenv("LM_API_KEY")})
	results := make([]result, 0, len(req.Samples))
	for i, sample := range req.Samples {
		ctx, cancel := context.WithTimeout(context.Background(), 2*time.Minute)
		start := time.Now()
		raw, err := client.Resolve(ctx, sample.Text, help, req.PromptVersion)
		elapsed := time.Since(start)
		cancel()
		if err != nil && !errors.Is(err, llm.ErrInvalidModelOutput) {
			return fmt.Errorf("legacy sample %d: %w", i+1, err)
		}
		out := result{Decisions: make(map[string]decision), LegacyMS: elapsed.Seconds() * 1000}
		if err != nil {
			out.LegacyError = err.Error()
		}
		out.Decisions["legacy"] = decision{raw, commands.GuardResolverOutput(sample.Text, raw)}
		for name, pred := range sample.Predictions {
			out.Decisions[name] = decision{pred, commands.GuardResolverOutput(sample.Text, pred)}
		}
		results = append(results, out)
	}
	return json.NewEncoder(os.Stdout).Encode(results)
}

func main() {
	if err := run(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
