/*
Package main is the entry point for the smarthome application.
It switches between server mode (Slack, MCP) and command line mode.
*/
package main

import (
	"context"
	"flag"
	"fmt"
	"log/slog"
	"os"
	"strings"

	"github.com/johtani/smarthome/internal/otel"
	"github.com/johtani/smarthome/server/cron"
	"github.com/johtani/smarthome/server/mcp"
	"github.com/johtani/smarthome/server/slack"
	"github.com/johtani/smarthome/subcommand"
)

func printHelp(commandsHelp string) string {
	fmt.Println("SlackBot用Serverを起動する場合は-serverオプションをつけてください")
	fmt.Println("MCPServerとして起動する場合は-mcpオプションをつけてください")
	fmt.Println("コマンドモードで利用可能なコマンドは次の通りです。")
	fmt.Print(commandsHelp)
	return `コマンドを指定してください。
smarthome <コマンド名>`
}

func main() {
	opts := &slog.HandlerOptions{
		Level: slog.LevelInfo,
	}
	var handler slog.Handler = slog.NewTextHandler(os.Stderr, opts)
	slog.SetDefault(slog.New(handler))

	if err := run(); err != nil {
		_, _ = fmt.Fprintf(os.Stderr, "%v", err)
		os.Exit(1)
	}
}

func run() error {
	var serverFlag bool
	var mcpFlag bool
	var configPath string
	flag.BoolVar(&serverFlag, "server", false, "SlackBot用Serverを起動するかどうか")
	flag.BoolVar(&mcpFlag, "mcp", false, "MCPServerとして起動するかどうか")
	flag.StringVar(&configPath, "config", subcommand.ConfigFileName, "MCPServerの時の設定ファイルのパス")
	flag.Parse()

	ctx := context.Background()
	shutdown, err := otel.SetupOTEL(ctx, "smarthome")
	if err != nil {
		return err
	}
	defer func() {
		_ = shutdown(ctx)
	}()

	switch {
	case serverFlag:
		config, err := subcommand.LoadConfig()
		if err != nil {
			return fmt.Errorf("設定の読み込みに失敗: %w", err)
		}
		_, _ = fmt.Fprintf(os.Stdout, "%v\n", os.Getpid())

		// cronはgoroutineで実行するため、エラーハンドリングが必要
		errChan := make(chan error, 1)
		go func() {
			if err := cron.Run(config); err != nil {
				errChan <- err
			}
		}()

		// cron起動時のエラーチェック（非ブロッキング）
		select {
		case err := <-errChan:
			return fmt.Errorf("cron起動失敗: %w", err)
		default:
		}

		return slack.Run(config)
	case mcpFlag:
		config, err := subcommand.LoadConfigWithPath(configPath)
		if err != nil {
			return fmt.Errorf("設定の読み込みに失敗: %w", err)
		}
		return mcp.Run(config)
	default:
		config, err := subcommand.LoadConfigWithPath(configPath)
		if err != nil {
			return fmt.Errorf("設定の読み込みに失敗: %w", err)
		}
		return runCmd(ctx, config, flag.Args())
	}
}

func runCmd(ctx context.Context, config subcommand.Config, cmdArgs []string) error {
	if len(cmdArgs) < 2 {
		return fmt.Errorf("%s", printHelp(config.Commands.Help()))
	}
	name := strings.Join(cmdArgs, " ")
	d, args, dymMsg, err := config.Commands.Find(name)
	if err != nil {
		_, _ = fmt.Fprintf(os.Stderr, "%v\n", err)
		printHelp(config.Commands.Help())
		return nil
	}
	c := d.Init(config)
	msg, err := c.Exec(ctx, args)
	if err != nil {
		return err
	}
	if len(dymMsg) > 0 {
		msg = strings.Join([]string{dymMsg, msg}, "\n")
	}
	_, _ = fmt.Fprintln(os.Stdout, msg)

	return nil
}
