package main

import (
	"flag"
	"fmt"
	"github.com/johtani/smarthome/server/cron"
	"github.com/johtani/smarthome/server/mcp"
	"github.com/johtani/smarthome/server/slack"
	"github.com/johtani/smarthome/subcommand"
	"os"
	"strings"
)

func printHelp(commandsHelp string) string {
	fmt.Println("SlackBot用Serverを起動する場合は-serverオプションをつけてください")
	fmt.Println("MCPServerとして起動する場合は-mcpオプションをつけてください")
	fmt.Println("コマンドモードで利用可能なコマンドは次の通りです。")
	fmt.Printf(commandsHelp)
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
	var configPath string
	flag.BoolVar(&serverFlag, "server", false, "SlackBot用Serverを起動するかどうか")
	flag.BoolVar(&mcpFlag, "mcp", false, "MCPServerとして起動するかどうか")
	flag.StringVar(&configPath, "config", subcommand.ConfigFileName, "MCPServerの時の設定ファイルのパス")
	flag.Parse()

	if serverFlag {
		config := subcommand.LoadConfig()
		_, _ = fmt.Fprintf(os.Stdout, "%v\n", os.Getpid())
		go cron.Run(config)
		return slack.Run(config)
	} else if mcpFlag {
		config := subcommand.LoadConfigWithPath(configPath)
		mcp.Run(config)
		return nil
	} else {
		config := subcommand.LoadConfigWithPath(configPath)
		return runCmd(config, flag.Args())
	}
}

func runCmd(config subcommand.Config, cmdArgs []string) error {
	if len(cmdArgs) < 2 {
		return fmt.Errorf(printHelp(config.Commands.Help()))
	}
	name := strings.Join(cmdArgs, " ")
	d, args, dymMsg, err := config.Commands.Find(name)
	if err != nil {
		_, _ = fmt.Fprintf(os.Stderr, "%v\n", err)
		printHelp(config.Commands.Help())
		return nil
	}
	c := d.Init(config)
	msg, err := c.Exec(args)
	if err != nil {
		return err
	}
	if len(dymMsg) > 0 {
		msg = strings.Join([]string{dymMsg, msg}, "\n")
	}
	_, _ = fmt.Fprintln(os.Stdout, msg)

	return nil
}
