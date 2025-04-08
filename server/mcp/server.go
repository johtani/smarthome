package mcp

import (
	"context"
	"errors"
	"github.com/johtani/smarthome/subcommand"
	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
	"strings"
)

func Run(config subcommand.Config) {
	s := NewMCPServer()

	// 登録してあるコマンドをMCPのツールとして登録していく
	for _, definition := range config.Commands.Definitions {
		s.AddTool(NewMCPTool(definition, config))
	}

	if err := server.ServeStdio(s); err != nil {
		panic(err)
	}
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
		return mcp.NewTool(strings.ReplaceAll(definition.Name, " ", "_"), mcp.WithDescription(definition.Description)),
			func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
				msg, err := definition.Init(config).Exec("")
				if err != nil {
					return nil, errors.New(definition.Name + ": " + err.Error())
				}
				return mcp.NewToolResultText(msg), nil
			}
	} else {
		// TODO Definition.ArgsをmcpのArgのリストに変換する
		return mcp.NewTool(strings.ReplaceAll(definition.Name, " ", "_"), mcp.WithDescription(definition.Description)),
			func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
				// TODO Definition.Argsをもとにrequest.paramsから入力を取り出して、文字列連結する
				msg, err := definition.Init(config).Exec("")
				if err != nil {
					return nil, errors.New(definition.Name + ": " + err.Error())
				}
				return mcp.NewToolResultText(msg), nil
			}
	}
}
