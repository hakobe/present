# Present

はてなブックマークから集めた人気の新着URLを、定期的にお知らせするSlack用のbotプログラムです。

- はてなブックマークを検索して指定したタグの新着URLを取得
- 適度に人気で新しいURLを選んで定期的にSlackのチャンネルに発言
- Slackのチャンネルで人間が会話している時には遠慮して発言しない
- Slackのチャンネルで人間が会話していない時には発言しすぎない

## 起動

1. MySQLに新しいデータベースを作る
2. botが常駐するチャンネルを選んでSlackのintegrationを設定する
  - SlackのIncomming WebHooksを設定する
    - "Post to channel"をbotが常駐するチャンネルを設定
  - SlackのOutgoing WebHooksを設定する
    - "Channel"にbotが常駐するチャンネルを設定
    - "URL(s)"にこのbotのhook APIのURLを設定 (例: http://example.com:8080/hook)
3. 環境変数に設定を与えて起動する

```sh
$ go get github.com/hakobe/present
$ PRESENT_SLACK_INCOMMING_URL="https://hooks.slack.com/services/ABCD1234/EFGH5679/abcdefghijk123456" \
  PRESENT_DB_DSN="id:pass@tcp(mysqldhost:3306)/dbname?parseTime=true&charset=utf8" \
  PRESENT_NAME=engineerkun \          # コマンドを実行するときに呼ぶbotの名前
  PRESENT_WAIT=900 \                  # URLを発言する頻度(秒)
  PRESENT_NOOP_LIMIT=3 \              # この回数だけ連続して発言したら一時停止する
  PORT="8080" \                       # WebHooksを待ち受けるHTTPサーバのポート
  $GOPATH/bin/present
```

## 使い方

- 予めタグを設定したあと、ほうっておくと設定したタグではてなブックマークを検索し、`PRESENT_WAIT`で設定した秒数ごとにURLを発言します
- `PRESENT_NAME`に設定した文字列のあとにコマンドを続けて発言することで、botに指示を与えることができます

### コマンド一覧
`PRESENT_NAME`に`engineerkun`と設定した場合

- `engineerkun tags`
  - 現在検索対象に設定されているタグの一覧を表示する
- `engineerkun add <tag>`
  - &lt;tag&gt; を検索対象に追加する
- `engineerkun del <tag>`
  - &lt;tag&gt; を検索対象から削除する
- `engineerkun plz`
  - URLを強制的に発言させる
- `engineerkun help`
  - ヘルプを表示する


