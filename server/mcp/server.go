/*
Package mcp provides an MCP (Model Context Protocol) server implementation.
It exposes subcommands as tools that can be called by LLMs.
*/
package mcp

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/johtani/smarthome/internal/configstore"
	"github.com/johtani/smarthome/subcommand"
	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
)

// Run starts the MCP server.
func Run(configStore *configstore.Store) error {
	s := NewMCPServer()
	unsubscribe := observeTools(s, configStore)
	defer unsubscribe()

	if err := server.ServeStdio(s); err != nil {
		return fmt.Errorf("MCPサーバーの起動に失敗しました: %w", err)
	}
	return nil
}

func observeTools(s *server.MCPServer, store *configstore.Store) func() {
	return store.Observe(func(config subcommand.Config) {
		// Keep definitions and connection settings from the same snapshot, even
		// when a reload overlaps a call that is already running.
		snapshot := func() subcommand.Config { return config }
		tools := make([]server.ServerTool, 0, len(config.Commands.Definitions))
		for _, definition := range config.Commands.Definitions {
			tool, handler := newMCPTool(definition, snapshot)
			tools = append(tools, server.ServerTool{Tool: tool, Handler: handler})
		}
		s.SetTools(tools...)
	})
}

// NewMCPServer creates and returns a new MCPServer instance.
func NewMCPServer() *server.MCPServer {
	return server.NewMCPServer(
		"Smart Home MCP",
		"0.1.0",
		server.WithLogging(),
		server.WithToolCapabilities(true),
	)
}

// NewMCPTool creates an MCP tool and its handler from a subcommand definition.
func NewMCPTool(definition subcommand.Definition, configStore *configstore.Store) (mcp.Tool, server.ToolHandlerFunc) {
	return newMCPTool(definition, configStore.Get)
}

func newMCPTool(definition subcommand.Definition, getConfig func() subcommand.Config) (mcp.Tool, server.ToolHandlerFunc) {
	if definition.Args == nil {
		tmp := mcp.NewTool(strings.ReplaceAll(definition.Name, " ", "_"), mcp.WithDescription(definition.Description))
		return tmp,
			func(ctx context.Context, _ mcp.CallToolRequest) (*mcp.CallToolResult, error) {
				config := getConfig()
				msg, err := definition.Init(config).Exec(ctx, "")
				if err != nil {
					return nil, errors.New(definition.Name + ": " + err.Error())
				}
				return mcp.NewToolResultText(msg), nil
			}
	}
	args := []mcp.ToolOption{mcp.WithDescription(definition.Description)}
	for _, arg := range definition.Args {

		opts := []mcp.PropertyOption{mcp.Description(arg.Description)}

		if arg.Required {
			opts = append(opts, mcp.Required())
		}
		if len(arg.Enum) > 0 {
			opts = append(opts, mcp.Enum(arg.Enum...))
		}
		args = append(args,
			mcp.WithString(
				arg.Name,
				opts...,
			))

	}
	tmp := mcp.NewTool(strings.ReplaceAll(definition.Name, " ", "_"), args...)
	return tmp,
		func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
			config := getConfig()
			params := []string{}
			for _, arg := range definition.Args {
				if p, ok := request.GetArguments()[arg.Name]; ok {
					if arg.Prefix != "" {
						params = append(params, fmt.Sprint(arg.Prefix, p))
					} else {
						params = append(params, fmt.Sprint(p))
					}
				}
			}
			msg, err := definition.Init(config).Exec(ctx, strings.Join(params, " "))
			if err != nil {
				return nil, errors.New(definition.Name + ": " + err.Error())
			}
			return mcp.NewToolResultText(msg), nil
		}
}
