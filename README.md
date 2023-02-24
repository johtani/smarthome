# 家の家電をコマンドラインで操作するためのツール

以下のAPIを利用して家電などを操作するコマンド群。

* [OwnTone API](https://owntone.github.io/owntone-server/json-api/)
* [SwitchBot API](https://github.com/OpenWonderLabs/SwitchBotAPI)
 
## 必要な設定

* OWNTONE_URL : 例："http://localhost:3689"
* SWITCHBOT_TOKEN : [See detail on API doc](https://github.com/OpenWonderLabs/SwitchBotAPI#getting-started) 
* SWITCHBOT_SECRET : [See detail on API doc](https://github.com/OpenWonderLabs/SwitchBotAPI#getting-started)
  

## ビルド

```
go build
```

## 実行

実行例：

```
smart_home start-meeting
```

`start-meeting`がサブコマンド。サブコマンドを指定しない場合は現在利用可能なサブコマンドの一覧が表示される。

## subcommand

サブコマンド単位で、いくつかの操作をまとめて実行することを想定しています。
利用できるサブコマンドは[ここ](https://github.com/johtani/smart_home/blob/master/subcommand/subcommand.go#L43)で登録しています。
サブコマンドは完全に自分好みに実装しています。。。

## ライセンス

MITライセンス