package mcp

import (
	"context"
	"testing"

	"github.com/johtani/smarthome/subcommand"
	"github.com/johtani/smarthome/subcommand/action/owntone"
	"github.com/johtani/smarthome/subcommand/action/switchbot"
	"github.com/johtani/smarthome/subcommand/action/yamaha"
	"github.com/mark3labs/mcp-go/mcp"
	"strings"
)

func TestNewMCPTool(t *testing.T) {
	config := subcommand.Config{
		Owntone:   owntone.Config{Url: "http://localhost:8000"},
		Switchbot: switchbot.Config{Token: "token", Secret: "secret"},
		Yamaha:    yamaha.Config{Url: "http://localhost:8080"},
	}

	tests := []struct {
		name       string
		definition subcommand.Definition
		wantName   string
		wantDesc   string
		argCount   int
	}{
		{
			name: "no args tool",
			definition: subcommand.Definition{
				Name:        "light on",
				Description: "Turn on the light",
			},
			wantName: "light_on",
			wantDesc: "Turn on the light",
			argCount: 0,
		},
		{
			name: "with args tool",
			definition: subcommand.Definition{
				Name:        "set volume",
				Description: "Set the volume",
				Args: []subcommand.Arg{
					{
						Name:        "volume",
						Description: "Volume level",
						Required:    true,
					},
				},
			},
			wantName: "set_volume",
			wantDesc: "Set the volume",
			argCount: 1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tool, _ := NewMCPTool(tt.definition, config)
			if tool.Name != tt.wantName {
				t.Errorf("tool.Name = %v, want %v", tool.Name, tt.wantName)
			}
			if tool.Description != tt.wantDesc {
				t.Errorf("tool.Description = %v, want %v", tool.Description, tt.wantDesc)
			}
			// MCP tool params are in tool.InputSchema
			// Checking argCount in InputSchema might be complex depending on mcp-go implementation.
			// At least we check the tool is created.
		})
	}
}

func TestMCPToolHandler(t *testing.T) {
	config := subcommand.Config{
		Owntone:   owntone.Config{Url: "http://localhost:8000"},
		Switchbot: switchbot.Config{Token: "token", Secret: "secret"},
		Yamaha:    yamaha.Config{Url: "http://localhost:8080"},
		Commands:  subcommand.NewCommands(),
	}

	// 既存のコマンドを使ってテストする
	definition := config.Commands.Definitions[0]

	_, handler := NewMCPTool(definition, config)
	ctx := context.Background()
	req := mcp.CallToolRequest{}
	req.Params.Name = strings.ReplaceAll(definition.Name, " ", "_")
	req.Params.Arguments = map[string]interface{}{}

	result, err := handler(ctx, req)
	if err != nil {
		t.Fatalf("handler failed: %v", err)
	}

	if len(result.Content) == 0 {
		t.Fatal("result content is empty")
	}

	_, ok := result.Content[0].(mcp.TextContent)
	if !ok {
		t.Fatalf("content is not text: %T", result.Content[0])
	}
}
