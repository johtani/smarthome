package main

import (
	"flag"
	"fmt"
	"os"
	"smart_home/server"
	"smart_home/subcommand"
)

func printHelp(smap map[string]subcommand.Definition) string {
	fmt.Println("SlackBot用Serverを起動する場合は-serverオプションをつけてください")
	fmt.Println("コマンドモードで利用可能なコマンドは次の通りです。")
	for _, command := range smap {
		fmt.Printf("  %s: %s\n", command.Name, command.Description)
	}
	return `コマンドを指定してください。
smart_home <コマンド名>`
}

func main() {
	if err := run(); err != nil {
		fmt.Fprintf(os.Stderr, "%v", err)
		os.Exit(1)
	}
}

func run() error {
	config, err := subcommand.NewConfig()
	if err != nil {
		return err
	}
	smap := subcommand.Map()
	var serverFlag bool
	flag.BoolVar(&serverFlag, "server", false, "SlackBot用Serverを起動するかどうか")
	flag.Parse()
	if serverFlag {
		return server.Run(config, smap)
	} else {
		return runCmd(config, smap)
	}
	return nil
}

func runCmd(config subcommand.Config, smap map[string]subcommand.Definition) error {
	if len(os.Args) < 2 {
		return fmt.Errorf(printHelp(smap))
	}
	name := os.Args[1]
	d, ok := subcommand.Map()[name]
	if ok {
		c := d.Init(config)
		err := c.Exec()
		if err != nil {
			return err
		}
	} else {
		fmt.Fprintf(os.Stderr, "command[%v] is not found.\n", name)
		printHelp(smap)
	}
	return nil
}
