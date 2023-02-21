package main

import (
	"fmt"
	"os"
	"smart_home/subcommand"
)

func printSubcommands() {
	fmt.Println("利用可能なコマンドは次の通りです。")
	for _, command := range subcommand.SubcommandMap() {
		fmt.Printf("  %s: %s\n", command.Name, command.Description)
	}
}

func printHelp() string {
	// TODO もうちょっときれいに？
	printSubcommands()
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
	// 第1引数がない場合はヘルプを出して終了
	if len(os.Args) < 2 {
		return fmt.Errorf(printHelp())
	}
	// 第1引数の文字列から、実行するサブコマンドを決定する
	name := os.Args[1]
	// コマンドのインスタンスを探す（全部じゃなくて、呼ばれたやつだけインスタンス化してもよさそう？）
	c, ok := subcommand.SubcommandMap()[name]
	if ok {
		// TODO configを読み込む(コマンドごとにする？それともここで全部読み取る？コマンドで設定を読み込むのと十分かをチェックするメソッドを用意すればいいか？)
		err := c.CheckConfig()
		if err != nil {
			return err
		}
		err = c.Exec()
		if err != nil {
			return err
		}
	}

	return nil
}
