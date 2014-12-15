# Present

はてなブックマークから集めた人気の新着URLを、定期的にお知らせするSlack用のbotプログラムです。

- 指定したタグの新着URLをはてなブックマークを定期チェックして取得
- 適度に人気で新しいURLを選んで定期的にSlackのチャンネルに発言
- Slackのチャンネルで人間が会話している時には遠慮して発言しない
- Slackのチャンネルで人間が会話していない時には発言しすぎない

## 起動

1. MySQLに新しいデータベースを作る
2. SlackのIncomming WebHooksを設定する
  - "Post to channel"にURLを発言するチャンネルを設定
3. SlackのOutgoing WebHooksを設定する
  - "Channel"にURLを発言するチャンネルを設定
  - "URL(s)"にこのbotのhook APIのURLを設定 (例: http://example.com:8080/hook)
4. 環境変数に設定を与えて起動する

```sh
$ go get github.com/hakobe/present
$ PORT="8080" \
  PRESENT_SLACK_INCOMMING_URL="https://hooks.slack.com/services/ABCD1234/EFGH5679/abcdefghijk123456" \
  PRESENT_DB_DSN="id:pass@tcp(mysqldhost:3306)/dbname?parseTime=true&charset=utf8" \
  PRESENT_TAGS=perl,ruby,javascript \ # チェック対象のタグ(カンマ区切り)
  PRESENT_WAIT=900 \                  # URLを発言する頻度(秒)
  PRESENT_NOOP_LIMIT=3 \              # この回数だけ連続して発言したら一時停止する
  $GOPATH/bin/present
```

## 使い方

- 放っておくと`PRESENT_WAIT`で指定した秒数ごとにURLをSlackのチャンネルに発言します
- 人間がチャンネルに`present plz`と発言することでURLの発言を促すことができます
