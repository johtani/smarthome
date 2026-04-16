/*
Package llm provides a client for resolving natural language to subcommands using LLMs.
*/
package llm

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"strings"

	"github.com/johtani/smarthome/internal/resolver"
	"go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
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
	if c.Endpoint == "" && c.Model == "" {
		return nil
	}
	if c.Endpoint != "" && c.Model == "" {
		errs = append(errs, "llm.model is required when llm.endpoint is set")
	}
	if c.Endpoint == "" && c.Model != "" {
		errs = append(errs, "llm.endpoint is required when llm.model is set")
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

const maxTraceBodyLength = 4096

func truncateForTrace(v string) string {
	if len(v) <= maxTraceBodyLength {
		return v
	}
	return v[:maxTraceBodyLength] + "...truncated"
}

func hashText(v string) string {
	sum := sha256.Sum256([]byte(strings.TrimSpace(v)))
	return hex.EncodeToString(sum[:12])
}

func traceRequestBody(payload map[string]any) string {
	tracePayload := map[string]any{
		"model": payload["model"],
		"messages": []map[string]string{
			{"role": "system", "content": "<redacted>"},
			{"role": "user", "content": "<redacted>"},
		},
		"response_format": payload["response_format"],
	}
	b, err := json.Marshal(tracePayload)
	if err != nil {
		return `{"error":"failed to marshal trace payload"}`
	}
	return string(b)
}

// Resolve resolves the given text to a command using the LLM.
func (c *Client) Resolve(ctx context.Context, text string, commandList string, promptVersion string) (ResolvedCommand, error) {
	ctx, span := otel.Tracer("llm").Start(ctx, "llm.Resolve", trace.WithAttributes(
		attribute.String("llm.input_text_hash", hashText(text)),
		attribute.String("llm.model", c.config.Model),
		attribute.String("resolver.path", "llm"),
		attribute.String("resolver.prompt_version", promptVersion),
	))
	defer span.End()
	if requestID, ok := resolver.RequestIDFromContext(ctx); ok {
		span.SetAttributes(attribute.String("resolver.request_id", requestID))
	}
	if channel, ok := resolver.ChannelFromContext(ctx); ok {
		span.SetAttributes(attribute.String("resolver.channel", channel))
	}

	// OpenAI Chat Completion API 互換のパラメータを構築
	// Structured Outputs (JSON mode) を使用することを想定

	systemPrompt := fmt.Sprintf(`あなたはスマートホームの操作を補助するアシスタントです。
以下のコマンドリストから、ユーザーの意図に最も合致するものを選び、JSON形式で返してください。
合致するものがない場合は、commandに空文字列を入れてください。

選択ルール:
- コマンドの args 指定に必ず従ってください（required/optional/enum/prefix）。
- ユーザー入力に固有名詞（アーティスト名/曲名/アルバム名など）が含まれる再生要求は search and play を優先してください。
- start music はランダム再生用途としてのみ使ってください。
- 使うコマンドで解釈できない自由文字列を args に入れないでください。

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
	span.SetAttributes(
		attribute.String("llm.endpoint", c.config.Endpoint),
		attribute.String("llm.request_body", truncateForTrace(traceRequestBody(payload))),
	)

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

	responseBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return ResolvedCommand{}, fmt.Errorf("failed to read response body: %w", err)
	}
	span.SetAttributes(
		attribute.Int("llm.response_status_code", resp.StatusCode),
		attribute.String("llm.response_body", truncateForTrace(string(responseBody))),
	)

	if resp.StatusCode != http.StatusOK {
		slog.ErrorContext(
			ctx,
			"LLM API returned error status",
			"status",
			resp.StatusCode,
			"endpoint",
			c.config.Endpoint,
			"response_body",
			truncateForTrace(string(responseBody)),
		)
		return ResolvedCommand{}, fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	var result struct {
		Choices []struct {
			Message struct {
				Content string `json:"content"`
			} `json:"message"`
		} `json:"choices"`
	}

	if err := json.Unmarshal(responseBody, &result); err != nil {
		return ResolvedCommand{}, fmt.Errorf("failed to decode response: %w", err)
	}

	if len(result.Choices) == 0 {
		return ResolvedCommand{}, fmt.Errorf("no choices in response")
	}

	content := result.Choices[0].Message.Content
	span.SetAttributes(attribute.String("llm.response_content", content))

	var resolved ResolvedCommand
	if err := json.Unmarshal([]byte(content), &resolved); err != nil {
		return ResolvedCommand{}, fmt.Errorf("failed to unmarshal resolved command: %w", err)
	}

	span.SetAttributes(
		attribute.String("llm.resolved_command", resolved.Command),
		attribute.String("llm.resolved_args", resolved.Args),
		attribute.String("llm.thought", resolved.Thought),
		attribute.String("resolver.resolved_command", resolved.Command),
		attribute.String("resolver.resolved_args", resolved.Args),
	)

	return resolved, nil
}
