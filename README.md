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
* influxdb
  * url : 例: "http://IPアドレス:ポート"
  * token : [See detail on API doc](https://docs.influxdata.com/influxdb/v2/admin/tokens/)
  * bucket : 例："バケット名"

### SwitchbotのデバイスのID

いくつかのデバイスのIDを指定するコマンドを用意しています。
それらのIDは以下のコマンドを実行すると一覧で取得できるので、必要なデバイスのIDをconfigに設定します。

```shell
smarthome device list
```

※デバイスの追加や家電の買い替えの後にIDの追加や入れ替えが必要になります。


### SlackのSocket Mode

SlackのSocket Modeを利用したサーバー機能も用意しています。
[config/slack.json.sample](./config/slack.json.sample)を`slack.json`に変更して値を設定します。

* bot_token : "xoxb-"で始まるトークン
* app_token : "xapp-"で始まるトークン
* debug : デバッグログ出力のtrue/false

`smarthome -server`で起動します。
Slackボットに対するメンションのみに対応しています。 
`@slackbot start music`のようにメンションすることでサブコマンドが実行されます。

#### Slash Command対応

SlackのボットのSlash Commandにも対応しています。
Subcommand名をもとに、以下のようにSlash Command名として登録することで、Slackでの呼び出しが楽になります。

* 空白を"-"に
* 先頭に"/"に（Slash Commandが自動的に付与する）

[Slash Commandの詳細についてはSlackの公式ガイドを参考に](https://api.slack.com/interactivity/slash-commands)。

## ビルド

```
go build
```

## 実行

実行例：

```
smarthome start meeting
```

`start meeting`がサブコマンド。サブコマンドを指定しない（もしくは`help`サブコマンドを指定した）場合は現在利用可能なサブコマンドの一覧が表示される。

## subcommand

サブコマンド単位で、いくつかの操作をまとめて実行することを想定しています。
利用できるサブコマンド一覧は`smarthome help`で表示されます。
サブコマンドは完全に自分好みに実装しています。。。

## 定期実行処理

Switchbotの温湿度計のデータを取得し、10分おきにInfluxDBに保存します。
対象となる機器は[record-temperature.go](https://github.com/johtani/smarthome/blob/master/server/cron/record-temperature.go#L11)で指定しています。

## ライセンス

MITライセンス