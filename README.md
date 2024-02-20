# 家の家電をコマンドラインで操作するためのツール

以下のAPIを利用して家電などを操作するコマンド群。

* [OwnTone API](https://owntone.github.io/owntone-server/json-api/)
* [SwitchBot API](https://github.com/OpenWonderLabs/SwitchBotAPI)
* [Yamaha Extended Control](https://github.com/rsc-dev/pyamaha/blob/master/doc/YXC_API_Spec_Basic_v2.0.pdf)
 
## 必要な設定

[config/config.json.sample](./config/config.json.sample)を`config.json`に変更して、各種設定を行います。

* owntone.url : 例："http://localhost:3689"
* switchbot.token : [See detail on API doc](https://github.com/OpenWonderLabs/SwitchBotAPI#getting-started) 
* switchbot.secret : [See detail on API doc](https://github.com/OpenWonderLabs/SwitchBotAPI#getting-started)
* yamaha.url : 例："http://IPアドレス"

### SlackのSocket Mode

SlackのSocket Modeを利用したサーバー機能も用意しています。
[config/slack.json.sample](./config/slack.json.sample)を`slack.json`に変更して値を設定します。

* bot_token : "xoxb-"で始まるトークン
* app_token : "xapp-"で始まるトークン
* debug : デバッグログ出力のtrue/false

`smarthome -server`で起動します。
Slackボットに対するメンションのみに対応しています。 
`@slackbot start music`のようにメンションすることでサブコマンドが実行されます。

## ビルド

```
go build
```

## 実行

実行例：

```
smarthome start meeting
```

`start meeting`がサブコマンド。サブコマンドを指定しない場合は現在利用可能なサブコマンドの一覧が表示される。

## subcommand

サブコマンド単位で、いくつかの操作をまとめて実行することを想定しています。
利用できるサブコマンド一覧は`smarthome help`で表示されます。
サブコマンドは完全に自分好みに実装しています。。。

## ライセンス

MITライセンス