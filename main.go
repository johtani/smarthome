package main

import (
	"fmt"
	"os"
	"smart_home/subcommand"
)

func printHelp(smap map[string]subcommand.Subcommand) string {
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
	smap := subcommand.Map(config)
	// 第1引数がない場合はヘルプを出して終了
	if len(os.Args) < 2 {
		return fmt.Errorf(printHelp(smap))
	}
	// 第1引数の文字列から、実行するサブコマンドを決定する
	name := os.Args[1]
	// コマンドのインスタンスを探す（全部じゃなくて、呼ばれたやつだけインスタンス化してもよさそう？）
	c, ok := subcommand.Map(config)[name]
	if ok {
		err = c.Exec()
		if err != nil {
			return err
		}
	}

	return nil
}
