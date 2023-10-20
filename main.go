package main

import (
	"flag"
	"fmt"
	"github.com/johtani/smarthome/server/slack"
	"github.com/johtani/smarthome/subcommand"
	"os"
	"strings"
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
	dymMsg := ""
	d, err := config.Commands.Find(name, false)
	if err != nil {
		candidates, cmds := config.Commands.DidYouMean(name, true)
		if len(candidates) == 0 {
			fmt.Fprintf(os.Stderr, "command[%v] is not found.\n", name)
			printHelp(config.Commands.Help())
			return nil
		} else {
			d = candidates[0]
			dymMsg = fmt.Sprintf("Did you mean \"%v\"?", cmds[0])
		}
	}
	c := d.Init(config)
	msg, err := c.Exec()
	if err != nil {
		return err
	}
	if len(dymMsg) > 0 {
		msg = strings.Join([]string{dymMsg, msg}, "\n")
	}
	fmt.Fprintln(os.Stdout, msg)

	return nil
}
