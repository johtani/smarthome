package main

import (
	"flag"
	"fmt"
	"github.com/johtani/smarthome/server/slack"
	"github.com/johtani/smarthome/subcommand"
	"os"
)

func printHelp(commandsHelp string) string {
	fmt.Println("SlackBot用Serverを起動する場合は-serverオプションをつけてください")
	fmt.Println("コマンドモードで利用可能なコマンドは次の通りです。")
	fmt.Printf(commandsHelp)
	return `コマンドを指定してください。
smarthome <コマンド名>`
}

func main() {
	if err := run(); err != nil {
		fmt.Fprintf(os.Stderr, "%v", err)
		os.Exit(1)
	}
}

func run() error {
	config := subcommand.LoadConfig()
	var serverFlag bool
	flag.BoolVar(&serverFlag, "server", false, "SlackBot用Serverを起動するかどうか")
	flag.Parse()
	if serverFlag {
		fmt.Fprintf(os.Stdout, "%v\n", os.Getpid())
		return slack.Run(config)
	} else {
		return runCmd(config)
	}
}

func runCmd(config subcommand.Config) error {
	if len(os.Args) < 2 {
		return fmt.Errorf(printHelp(config.Commands.Help()))
	}
	name := os.Args[1]
	d, err := config.Commands.Find(name, false)
	if err != nil {
		fmt.Fprintf(os.Stderr, "command[%v] is not found.\n", name)
		printHelp(config.Commands.Help())
	} else {
		c := d.Init(config)
		msg, err := c.Exec()
		if err != nil {
			return err
		}
		fmt.Fprintln(os.Stdout, msg)
	}
	return nil
}
