/*
Package llm provides a client for resolving natural language to subcommands using LLMs.
*/
package llm

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	"go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp"
)

// Config is the configuration for the LLM client.
type Config struct {
	APIKey   string `json:"api_key"`
	Endpoint string `json:"endpoint"`
	Model    string `json:"model"`
}

// Validate validates the LLM configuration.
func (c Config) Validate() error {
	var errs []string
	if len(c.Endpoint) == 0 {
		errs = append(errs, "llm.endpoint is required")
	}
	if len(c.Model) == 0 {
		errs = append(errs, "llm.model is required")
	}
	if len(errs) > 0 {
		return fmt.Errorf("llm config validation failed: %s", strings.Join(errs, ", "))
	}
	return nil
}

// Client is a client for the LLM API.
type Client struct {
	config     Config
	httpClient *http.Client
}

// NewClient creates a new LLM client.
func NewClient(config Config) *Client {
	return &Client{
		config: config,
		httpClient: &http.Client{
			Transport: otelhttp.NewTransport(http.DefaultTransport),
		},
	}
}

// ResolvedCommand represents the result of LLM resolution.
type ResolvedCommand struct {
	Command string `json:"command"`
	Args    string `json:"args"`
	Thought string `json:"thought"`
}

// Resolve resolves the given text to a command using the LLM.
func (c *Client) Resolve(ctx context.Context, text string, commandList string) (ResolvedCommand, error) {
	// OpenAI Chat Completion API 互換のパラメータを構築
	// Structured Outputs (JSON mode) を使用することを想定

	systemPrompt := fmt.Sprintf(`あなたはスマートホームの操作を補助するアシスタントです。
以下のコマンドリストから、ユーザーの意図に最も合致するものを選び、JSON形式で返してください。
合致するものがない場合は、commandに空文字列を入れてください。

コマンドリスト:
%s

返却形式:
{
  "command": "コマンド名",
  "args": "引数",
  "thought": "なぜそのコマンドを選んだかの理由"
}`, commandList)

	payload := map[string]any{
		"model": c.config.Model,
		"messages": []map[string]string{
			{"role": "system", "content": systemPrompt},
			{"role": "user", "content": text},
		},
		"response_format": map[string]string{"type": "json_object"},
	}

	jsonData, err := json.Marshal(payload)
	if err != nil {
		return ResolvedCommand{}, fmt.Errorf("failed to marshal payload: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.config.Endpoint, strings.NewReader(string(jsonData)))
	if err != nil {
		return ResolvedCommand{}, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	if c.config.APIKey != "" {
		req.Header.Set("Authorization", "Bearer "+c.config.APIKey)
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return ResolvedCommand{}, fmt.Errorf("failed to send request: %w", err)
	}
	defer func() {
		_ = resp.Body.Close()
	}()

	if resp.StatusCode != http.StatusOK {
		return ResolvedCommand{}, fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	var result struct {
		Choices []struct {
			Message struct {
				Content string `json:"content"`
			} `json:"message"`
		} `json:"choices"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return ResolvedCommand{}, fmt.Errorf("failed to decode response: %w", err)
	}

	if len(result.Choices) == 0 {
		return ResolvedCommand{}, fmt.Errorf("no choices in response")
	}

	var resolved ResolvedCommand
	if err := json.Unmarshal([]byte(result.Choices[0].Message.Content), &resolved); err != nil {
		return ResolvedCommand{}, fmt.Errorf("failed to unmarshal resolved command: %w", err)
	}

	return resolved, nil
}
