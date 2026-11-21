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
	"os/signal"
	"strings"
	"syscall"

	"github.com/go-co-op/gocron/v2"
	"github.com/johtani/smarthome/internal/configstore"
	"github.com/johtani/smarthome/internal/otel"
	"github.com/johtani/smarthome/internal/resolver"
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
	if err := run(); err != nil {
		_, _ = fmt.Fprintf(os.Stderr, "%v", err)
		os.Exit(1)
	}
}

func run() error {
	var serverFlag bool
	var mcpFlag bool
	var configDir string
	var configFile string
	flag.BoolVar(&serverFlag, "server", false, "SlackBot用Serverを起動するかどうか")
	flag.BoolVar(&mcpFlag, "mcp", false, "MCPServerとして起動するかどうか")
	flag.StringVar(&configDir, "config-dir", subcommand.ConfigDirName, "設定ファイルのディレクトリパス")
	flag.StringVar(&configFile, "config", "", "[deprecated] -config-dirを使用してください")
	flag.Parse()

	ctx := context.Background()
	logLevel, err := otel.ParseLogLevel(os.Getenv("SMARTHOME_LOG_LEVEL"))
	if err != nil {
		return err
	}
	shutdown, err := otel.SetupOTEL(ctx, "smarthome")
	if err != nil {
		return err
	}
	stderrHandler := otel.NewTracingHandler(slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{Level: logLevel}))
	slog.SetDefault(slog.New(otel.NewLoggerHandler(stderrHandler, logLevel)))
	defer func() {
		_ = shutdown(ctx)
	}()

	loadConfig := func() (subcommand.Config, error) {
		if configFile != "" {
			slog.Warn("-config is deprecated, use -config-dir instead")
			return subcommand.LoadConfigWithPath(configFile)
		}
		return subcommand.LoadConfigFromDir(configDir)
	}

	config, err := loadConfig()
	if err != nil {
		return fmt.Errorf("設定の読み込みに失敗: %w", err)
	}
	logLLMDisabled(config)
	configStore := configstore.New(config)

	// Hot Reload のためのシグナル監視
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGHUP)
	go func() {
		for range sigChan {
			slog.InfoContext(ctx, "SIGHUP受信、設定を再読み込みします")
			if err := configStore.Reload(loadConfig); err != nil {
				slog.ErrorContext(ctx, "設定の再読み込みに失敗（古い設定を維持します）", "error", err)
				continue
			}
			logLLMDisabled(configStore.Get())
			slog.InfoContext(ctx, "設定を更新しました")
		}
	}()

	switch {
	case serverFlag:
		_, _ = fmt.Fprintf(os.Stdout, "%v\n", os.Getpid())

		return runServer(configStore, cron.Start, slack.Run)
	case mcpFlag:
		return mcp.Run(configStore)
	default:
		return runCmd(ctx, configStore, flag.Args())
	}
}

func runServer(store *configstore.Store, startCron func(*configstore.Store) (gocron.Scheduler, error), runSlack func(*configstore.Store) error) error {
	scheduler, err := startCron(store)
	if err != nil {
		return fmt.Errorf("cron起動失敗: %w", err)
	}
	defer func() {
		if err := scheduler.Shutdown(); err != nil {
			slog.Error("cron停止失敗", "error", err)
		}
	}()
	return runSlack(store)
}

func logLLMDisabled(config subcommand.Config) {
	if strings.TrimSpace(config.LLM.Endpoint) == "" {
		slog.Info("LLM is disabled; natural language resolution is unavailable (command matching only)")
	}
}

func runCmd(ctx context.Context, configStore *configstore.Store, cmdArgs []string) error {
	ctx, _ = resolver.EnsureRequestID(ctx)
	ctx = resolver.WithChannel(ctx, "cli")

	config := configStore.Get()
	if len(cmdArgs) < 1 {
		return fmt.Errorf("%s", printHelp(config.Commands.Help()))
	}
	name := strings.Join(cmdArgs, " ")
	d, args, dymMsg, err := config.Commands.Find(ctx, config, name)
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
