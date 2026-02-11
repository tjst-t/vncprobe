# 設計: セッション持続モード（デーモンモード）

## 概要

`session` サブコマンドを新設し、VNC 接続を維持したまま UNIX ドメインソケット経由で複数コマンドを受け付けるモードを提供する。

## CLI 設計

```
# セッション開始（フォアグラウンドでリッスン。バックグラウンド化は呼び出し側が & で行う）
vncprobe session start -s 10.0.0.1:5900 --socket /tmp/vncprobe.sock [--idle-timeout 300]

# セッション経由でコマンド実行
vncprobe capture --socket /tmp/vncprobe.sock -o screen.png
vncprobe key --socket /tmp/vncprobe.sock enter
vncprobe type --socket /tmp/vncprobe.sock "show interfaces"
vncprobe wait change --socket /tmp/vncprobe.sock

# セッション終了
vncprobe session stop --socket /tmp/vncprobe.sock
```

### 設計判断

- `session start` はフォアグラウンドで動作する。デーモン化はユーザーが `&` や `nohup` で行う
  - Go でのデーモン化は複雑（fork 問題）で、メリットに対してコストが大きい
  - Claude Code からは `&` で十分
- `--socket` がない場合は従来どおり直接 VNC 接続（後方互換性100%維持）
- `--socket` はグローバルオプションとして `cmd/root.go` の `ParseGlobalFlags` に追加

## 通信プロトコル

UNIX ドメインソケット上のシンプルなラインプロトコル（JSON Lines）。

### リクエスト（1行のJSON）

```json
{"command":"capture","args":["-o","screen.png"]}
{"command":"key","args":["enter"]}
{"command":"type","args":["show interfaces"]}
{"command":"wait","args":["change","--timeout","10"]}
{"command":"session","args":["stop"]}
```

### レスポンス（1行のJSON）

```json
{"ok":true}
{"ok":false,"error":"capture failed: timeout waiting for framebuffer update"}
```

- 1リクエスト = 1レスポンスの同期プロトコル
- 改行（`\n`）がメッセージ区切り
- コマンドの exit code はレスポンスに含めない（ok/error で十分）

## アーキテクチャ

### 新規ファイル

```
session/server.go       — セッションサーバ（ソケットリッスン、コマンド実行）
session/server_test.go  — サーバのユニットテスト
session/client.go       — セッションクライアント（ソケット接続、コマンド送信）
session/client_test.go  — クライアントのユニットテスト
session/protocol.go     — Request/Response 型定義
cmd/session.go          — session start/stop の CLI ハンドラ
```

### パッケージ構成の理由

`session/` を独立パッケージにする理由:
- `vnc/` はVNCプロトコルのロジックに専念
- セッション管理はVNCとは直交する関心事（ソケット通信、プロセスライフサイクル）
- `cmd/` からは `session.NewServer()` / `session.NewClient()` を呼ぶだけ

### プロトコル型定義（`session/protocol.go`）

```go
type Request struct {
    Command string   `json:"command"`
    Args    []string `json:"args"`
}

type Response struct {
    OK    bool   `json:"ok"`
    Error string `json:"error,omitempty"`
}
```

### セッションサーバ（`session/server.go`）

```go
type Server struct {
    client     vnc.VNCClient
    socketPath string
    listener   net.Listener
    idleTimeout time.Duration
}

func NewServer(client vnc.VNCClient, socketPath string, idleTimeout time.Duration) *Server
func (s *Server) ListenAndServe() error  // ブロッキング。SIGINT/SIGTERM でクリーンシャットダウン
func (s *Server) Shutdown() error         // グレースフルシャットダウン
```

#### 動作フロー

1. UNIX ドメインソケットを作成してリッスン
2. 接続を accept（同時接続は1つだけ。排他的にコマンドを処理する）
3. 1行ずつ JSON を読み、`Request` にデコード
4. `command` に応じて `cmd.RunCapture` / `cmd.RunKey` 等を呼び出す
5. 結果を `Response` として1行 JSON で返す
6. `session stop` を受けたら、レスポンスを返してからシャットダウン

#### 同時接続の扱い

- 同時接続は1クライアントのみ許可（mutex で排他制御）
- 2つ目の接続は即座にエラーレスポンスを返して切断
- VNC クライアントは単一接続なのでコマンドの並行実行は意味がない

#### アイドルタイムアウト

- 最後のコマンド実行から `--idle-timeout` 秒間コマンドがなければ自動シャットダウン
- デフォルト: 300秒（5分）
- 0 で無効化

#### ソケットファイルのクリーンアップ

- 起動時: ソケットファイルが既に存在する場合、接続を試みて生存確認
  - 接続できない（= 前のプロセスが死んでいる）→ 削除して続行
  - 接続できる（= 別のセッションが生きている）→ エラーで終了
- シャットダウン時: ソケットファイルを削除
- シグナルハンドリング: SIGINT/SIGTERM でクリーンシャットダウン（ソケット削除含む）

### セッションクライアント（`session/client.go`）

```go
type Client struct {
    socketPath string
}

func NewClient(socketPath string) *Client
func (c *Client) Execute(command string, args []string) error
```

- ソケットに接続 → Request を JSON で送信 → Response を読む → 切断
- 1コマンド1接続（コネクションプーリングは不要。オーバーヘッドは UNIX ソケットなので無視できる）

### main.go の変更

`run()` を以下のように拡張:

```go
func run(args []string) int {
    // ... command parsing ...

    switch command {
    case "session":
        return runSession(remaining)  // session は VNC 接続前に分岐
    }

    // Parse global flags
    opts, cmdArgs, err := cmd.ParseGlobalFlags(remaining)

    // --socket がある場合はセッション経由
    if opts.Socket != "" {
        return runViaSession(opts.Socket, command, cmdArgs)
    }

    // 従来どおり直接 VNC 接続
    client := vnc.NewRealClient()
    // ...
}
```

ポイント:
- `session start/stop` は VNC 接続の前に分岐する（`session start` が自分で接続する）
- `--socket` 付きの通常コマンドは VNC 接続せず、セッションクライアント経由で実行
- `--socket` なしは従来どおり

### GlobalOpts の拡張

```go
type GlobalOpts struct {
    Server   string
    Password string
    Timeout  int
    Socket   string  // 追加
}
```

- `--socket` がある場合、`-s` は不要（セッションが VNC 接続を持っている）
- `ParseGlobalFlags` の `-s` 必須チェックを `--socket` がない場合のみに変更

### cmd/session.go

```go
func RunSessionStart(args []string) (serverAddr, password, socketPath string, idleTimeout int, err error)
func RunSessionStop(socketPath string) error
```

- `session start` のフラグ解析: `-s`, `-p`, `--socket`, `--idle-timeout`
- `session stop`: `--socket` のみ

## テスト戦略

### ユニットテスト: session/server_test.go, session/client_test.go

- サーバ起動 → クライアントからコマンド送信 → レスポンス検証
- VNCClient はモックを使用
- テストごとに `t.TempDir()` でソケットパスを生成
- アイドルタイムアウトのテスト（短いタイムアウトを設定して自動シャットダウンを確認）
- 同時接続の排他制御テスト

### E2E テスト

フェイク VNC サーバ + セッションサーバ + セッションクライアントの組み合わせ:

- `TestE2ESessionCapture`: session start → capture via socket → session stop
- `TestE2ESessionMultipleCommands`: 複数コマンドを順に実行
- `TestE2ESessionIdleTimeout`: アイドルタイムアウトで自動終了

## エラーハンドリング

| エラー状況 | 動作 |
|-----------|------|
| ソケットファイルが既に使用中 | exit 2 + エラーメッセージ |
| セッションに接続できない | exit 2 + エラーメッセージ |
| セッション経由のコマンドがエラー | exit 3 |
| アイドルタイムアウト | サーバが自動終了（ログ出力） |

## 実装順序

1. `session/protocol.go` — 型定義
2. `session/server.go` + テスト — サーバ実装（モック VNCClient 使用）
3. `session/client.go` + テスト — クライアント実装
4. `cmd/session.go` — CLI ハンドラ
5. `cmd/root.go` — `--socket` グローバルオプション追加、`-s` バリデーション条件変更
6. `main.go` — `session` コマンド分岐と `--socket` 分岐
7. E2E テスト
8. README 更新

## 後方互換性

- `--socket` を指定しなければ全ての既存コマンドは今までどおり動作する
- グローバルオプションに `--socket` が追加されるだけで、既存のフラグには影響なし
- exit code の体系はそのまま維持
