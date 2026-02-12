package mcp

import (
	"context"
	"errors"
	"fmt"
	"github.com/johtani/smarthome/subcommand"
	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
	"strings"
)

func Run(config subcommand.Config) error {
	s := NewMCPServer()

	// 登録してあるコマンドをMCPのツールとして登録していく
	for _, definition := range config.Commands.Definitions {
		s.AddTool(NewMCPTool(definition, config))
	}

	if err := server.ServeStdio(s); err != nil {
		return fmt.Errorf("MCPサーバーの起動に失敗しました: %w", err)
	}
	return nil
}

func NewMCPServer() *server.MCPServer {
	return server.NewMCPServer(
		"Smrat Home MCP",
		"0.1.0",
		server.WithLogging(),
	)
}

func NewMCPTool(definition subcommand.Definition, config subcommand.Config) (mcp.Tool, server.ToolHandlerFunc) {
	if definition.Args == nil {
		tmp := mcp.NewTool(strings.ReplaceAll(definition.Name, " ", "_"), mcp.WithDescription(definition.Description))
		return tmp,
			func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
				msg, err := definition.Init(config).Exec(ctx, "")
				if err != nil {
					return nil, errors.New(definition.Name + ": " + err.Error())
				}
				return mcp.NewToolResultText(msg), nil
			}
	} else {
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
				params := []string{}
				for _, arg := range definition.Args {
					if p, ok := request.Params.Arguments[arg.Name]; ok {
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
}
