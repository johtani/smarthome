# 家の家電をコマンドラインで操作するためのツール

以下のAPIを利用して家電などを操作するコマンド群。

* [OwnTone API](https://owntone.github.io/owntone-server/json-api/)
* [SwitchBot API](https://github.com/OpenWonderLabs/SwitchBotAPI)
* [Yamaha Extended Control](https://github.com/rsc-dev/pyamaha/blob/master/doc/YXC_API_Spec_Basic_v2.0.pdf)
* [OpenAI Chat Completion API](https://platform.openai.com/docs/api-reference/chat) (または互換API)
 
## 必要な設定

設定ファイルはデフォルトで `./config/` ディレクトリに格納します。
`-config-dir` フラグでディレクトリを変更できます。

```
smarthome -config-dir /path/to/config
```

### config.json

[config/config.json.sample](./config/config.json.sample)を`config.json`に変更して、各種設定を行います。

* owntone.url : 例："http://localhost:3689"
* llm.endpoint : （任意）OpenAI互換APIのエンドポイント（例: "https://api.openai.com/v1/chat/completions"）
* llm.model : （任意）使用するモデル名（例: "gpt-4o"）
* yamaha.url : 例："http://IPアドレス"
* influxdb
  * url : 例: "http://IPアドレス:ポート"
  * bucket : 例："バケット名"

> **注意**: トークン・シークレット等の秘匿情報（`switchbot.token`, `switchbot.secret`, `llm.api_key`, `influxdb.token`）は設定ファイルへの記載を非推奨とします。
> 後述の「[環境変数によるシークレットの上書き](#環境変数によるシークレットの上書き)」を使って設定してください。

`llm.endpoint` と `llm.model` は自然言語解決を使う場合のみ設定してください。両方未設定の場合はLLM機能を無効化し、既存のコマンド一致/あいまい一致のみで動作します。

### SwitchbotのデバイスのID

いくつかのデバイスのIDを指定するコマンドを用意しています。
それらのIDは以下のコマンドを実行すると一覧で取得できるので、必要なデバイスのIDをconfigに設定します。

```shell
smarthome device list
```

※デバイスの追加や家電の買い替えの後にIDの追加や入れ替えが必要になります。

### macros.json（マクロ定義）

`macros.json` に複数のアクションをまとめたマクロ（シーケンス）を定義できます。
このファイルは任意で、存在しない場合はスキップされます。

```json
[
  {
    "name": "start meeting",
    "description": "会議開始：音楽停止 → ヤマハ電源ON",
    "shortnames": ["sm"],
    "ignore_error": false,
    "actions": [
      { "type": "owntone_pause" },
      { "type": "yamaha_power_on" }
    ]
  }
]
```

#### 注意：マクロ名の重複

マクロの `name` に既存の組み込みコマンド名、または他のマクロと同じ名前を指定した場合、そのマクロは**スキップ**されます（起動時にwarningログが出力されます）。

```
WARN macro skipped: name already registered macro_name=help
```

組み込みコマンドの一覧は `help` コマンドで確認できます。

#### アクションタイプ一覧

| タイプ | 説明 | パラメータ |
|--------|------|-----------|
| `owntone_pause` | 音楽を一時停止 | - |
| `owntone_play` | 音楽を再生 | - |
| `owntone_clear_queue` | キューをクリア | - |
| `owntone_display_outputs` | 出力デバイスを表示 | `only_selected`: "true"/"false" |
| `yamaha_power_on` | ヤマハ電源ON | - |
| `yamaha_power_off` | ヤマハ電源OFF | - |
| `yamaha_set_input` | 入力切替 | `input`: 入力名 |
| `yamaha_set_volume` | 音量設定 | `volume`: 数値 |
| `yamaha_set_scene` | シーン設定 | `scene`: 数値 |
| `switchbot_send_command` | SwitchBotデバイス操作 | `device_id`, `command`: "turn_on"/"turn_off" |
| `switchbot_execute_scene` | SwitchBotシーン実行 | `scene_id` |
| `wait` | 待機 | `seconds`: 秒数（小数点可） |

`device_id` や `scene_id` には `$LightDeviceID`、`$LightSceneID`、`$AirConditionerID` のように `config.json` の値を参照できます。


### SlackのSocket Mode

SlackのSocket Modeを利用したサーバー機能も用意しています。
[config/slack.json.sample](./config/slack.json.sample)を`slack.json`に変更して値を設定します。

* debug : デバッグログ出力のtrue/false

> **注意**: `bot_token` および `app_token` は設定ファイルへの記載を非推奨とします。
> 後述の「[環境変数によるシークレットの上書き](#環境変数によるシークレットの上書き)」を使って設定してください。

`smarthome -server`で起動します。
Slackボットに対するメンションのみに対応しています。 
`@slackbot start music`のようにメンションすることでサブコマンドが実行されます。
また、自然言語（例：「エアコンをつけて」）での指示にも対応しており、既存のコマンドに一致しない場合はLLMが意図を解釈して適切なコマンドを実行します。

#### Slash Command対応

SlackのボットのSlash Commandにも対応しています。
Subcommand名をもとに、以下のようにSlash Command名として登録することで、Slackでの呼び出しが楽になります。

* 空白を"-"に
* 先頭に"/"に（Slash Commandが自動的に付与する）

[Slash Commandの詳細についてはSlackの公式ガイドを参考に](https://api.slack.com/interactivity/slash-commands)。

### 環境変数によるシークレットの上書き

設定ファイルに記載したシークレット情報は、以下の環境変数で上書きできます。
環境変数が設定されている場合は、設定ファイルの値より優先されます。

| 環境変数名 | 対応するフィールド |
|---|---|
| `SMARTHOME_SWITCHBOT_TOKEN` | `config.json` の `switchbot.token` |
| `SMARTHOME_SWITCHBOT_SECRET` | `config.json` の `switchbot.secret` |
| `SMARTHOME_LLM_API_KEY` | `config.json` の `llm.api_key` |
| `SMARTHOME_INFLUXDB_TOKEN` | `config.json` の `influxdb.token` |
| `SMARTHOME_SLACK_APP_TOKEN` | `slack.json` の `app_token` |
| `SMARTHOME_SLACK_BOT_TOKEN` | `slack.json` の `bot_token` |

その他の環境変数（URLやタイムアウト等）については [subcommand/config.go](./subcommand/config.go) を参照してください。

### Bitwarden Secrets Manager を使った秘匿情報管理

[Bitwarden Secrets Manager](https://bitwarden.com/products/secrets-manager/) と `bws` CLI を使うことで、トークン等の秘匿情報をサーバー上に平文で保存せずに管理できます。

#### 仕組み

```
Bitwarden Secrets Manager
        ↓ (BWS_ACCESS_TOKEN で認証)
    bws run
        ↓ (SMARTHOME_* を環境変数として展開)
    smarthome server
        ↓ (環境変数オーバーライドで各クライアントに設定)
    SwitchBot / Slack / LLM / InfluxDB ...
```

#### 設定手順

1. Bitwarden Secrets Manager にシークレットを登録（`SMARTHOME_SWITCHBOT_TOKEN` 等のキー名で）
2. Machine Account を作成し、アクセストークン（`BWS_ACCESS_TOKEN`）を取得
3. サーバーに [`bws` CLI](https://bitwarden.com/help/secrets-manager-cli/) をインストール
4. アクセストークンを専用ファイルに保存し、権限を制限:
   - systemd `--user` の場合（推奨）:
   ```bash
   mkdir -p ~/.config/smarthome
   echo "BWS_ACCESS_TOKEN=<your_access_token>" > ~/.config/smarthome/bws.env
   chmod 600 ~/.config/smarthome/bws.env
   ```
   - systemd システムサービス（`--system`）の場合:
   ```bash
   echo "BWS_ACCESS_TOKEN=<your_access_token>" > /etc/smarthome/bws.env
   chmod 600 /etc/smarthome/bws.env
   chown root:root /etc/smarthome/bws.env
   ```
5. systemd service ファイルを設定:
   - systemd `--user` の例:
   ```ini
   [Service]
   EnvironmentFile=%h/.config/smarthome/bws.env
   ExecStart=bws run -- /path/to/smarthome -server
   ExecReload=/bin/kill -HUP $MAINPID
   ```
   - systemd システムサービス（`--system`）の例:
   ```ini
   [Service]
   EnvironmentFile=/etc/smarthome/bws.env
   ExecStart=bws run -- /path/to/smarthome -server
   ExecReload=/bin/kill -HUP $MAINPID
   ```

`bws run` はシークレットを環境変数として展開した状態でコマンドを起動します。
`smarthome` は起動時にこれらの環境変数を読み込み、設定ファイルの値を上書きします。

### 設定の動的読み込み (Hot Reload)

アプリケーション（Slackボット、MCPサーバー、Cronジョブ）の実行中に、設定ファイル（`config.json`、`macros.json`）を編集して反映させることができます。
UNIX系環境では、以下のコマンドを実行して `SIGHUP` シグナルを送信することで再読み込みが行われます。

```bash
kill -HUP <pid>
```

#### systemd を使用している場合

systemd でサービス化している場合は、以下のコマンドで再読み込みを行うことができます。

1. **`systemctl reload` を使用する方法** (推奨)
   ユニットファイル（例：`/etc/systemd/system/smarthome.service`）に `ExecReload` が設定されている場合に使用できます。
   ```ini
   [Service]
   ...
   ExecReload=/bin/kill -HUP $MAINPID
   ...
   ```
   設定されている場合は、以下のコマンドを実行します。
   ```bash
   sudo systemctl reload smarthome
   ```

2. **直接シグナルを送る方法**
   ```bash
   sudo systemctl kill -s HUP smarthome
   ```

再読み込みに成功すると、標準エラー出力（または `journalctl -u smarthome`）にログが表示され、新しい設定が即座に反映されます。

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

組み込みのサブコマンドのほか、`macros.json` に定義したマクロもサブコマンドとして使用できます。

## 定期実行処理

Switchbotの温湿度計のデータを取得し、10分おきにInfluxDBに保存します。
対象となる機器は[record-temperature.go](https://github.com/johtani/smarthome/blob/master/server/cron/record-temperature.go#L11)で指定しています。

## OpenTelemetry

本ツールはOpenTelemetryに対応しています。トレースデータを送信するには、以下の環境変数を設定してください。

* `OTEL_EXPORTER_OTLP_ENDPOINT` : OTLPコレクターのエンドポイント（例: `http://localhost:4318`）
* `OTEL_SERVICE_NAME` : サービス名（デフォルト: `smarthome`）

現在は OTLP/HTTP エクスポーターを使用しています。

ログにはコンテキストから抽出された `trace_id` が付加されるため、特定のトレースに関連するログを容易に検索できます（`slog.InfoContext` 等を使用）。

## 開発ガイドライン

* **ブランチ作成**: 作業を開始する前に、必ず新しいブランチを作成してください。
* **一時ファイル**: 作業用の一時ファイルは `tmp/` ディレクトリの下に作成するか、作業が完了したら削除してください。 `tmp/` ディレクトリは `.gitignore` に含まれています。

## ライセンス

MITライセンス
