package main

import (
	"fmt"
	"os"
	"smart_home/subcommand"
)

func printHelp(smap map[string]subcommand.Definition) string {
	fmt.Println("利用可能なコマンドは次の通りです。")
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
	if len(os.Args) < 2 {
		return fmt.Errorf(printHelp(smap))
	}
	name := os.Args[1]
	d, ok := subcommand.Map()[name]
	if ok {
		c := d.Init(config)
		err = c.Exec()
		if err != nil {
			return err
		}
	} else {
		fmt.Fprintf(os.Stderr, "command[%v] is not found.\n", name)
		printHelp(smap)
	}

	return nil
}
