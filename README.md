# vncprobe

A single-binary CLI tool for VNC screen capture, keyboard input, and mouse operations. Designed for automating VM consoles via Claude Code.

## Install

Download a prebuilt binary from [GitHub Releases](https://github.com/tjst-t/vncprobe/releases/latest):

```bash
# Example: Linux amd64
curl -Lo vncprobe "https://github.com/tjst-t/vncprobe/releases/latest/download/vncprobe-$(curl -sL https://api.github.com/repos/tjst-t/vncprobe/releases/latest | grep -oP '\"tag_name\": \"\K[^\"]+' )-linux-amd64"
chmod +x vncprobe
sudo mv vncprobe /usr/local/bin/
```

Or install with `go install`:

```bash
go install github.com/tjst-t/vncprobe@latest
```

Or build from source:

```bash
git clone https://github.com/tjst-t/vncprobe.git
cd vncprobe
go build -o vncprobe .
```

## Usage

```
vncprobe <command> [options]

Commands:
  capture   Capture screen to PNG
  key       Send key input
  type      Type a string
  click     Mouse click
  move      Mouse move
  wait      Wait for screen change or stability
  session   Manage persistent VNC sessions

Global Options:
  -s, --server    VNC server address (required unless --socket is used)
  -p, --password  VNC password
  --timeout       Connection timeout in seconds (default: 10)
  --socket        Use session socket instead of direct connection
```

### Capture screenshot

```bash
vncprobe capture -s 10.0.0.1:5900 -o screen.png
```

### Send key input

```bash
# Single key
vncprobe key -s 10.0.0.1:5900 enter

# Modifier combinations
vncprobe key -s 10.0.0.1:5900 ctrl-c
vncprobe key -s 10.0.0.1:5900 alt-f4
vncprobe key -s 10.0.0.1:5900 ctrl-alt-delete
```

Supported keys: `enter`, `tab`, `escape`, `backspace`, `delete`, `space`, `up`, `down`, `left`, `right`, `home`, `end`, `pageup`, `pagedown`, `insert`, `f1`-`f12`

Modifiers: `ctrl`, `alt`, `shift`, `super`, `meta`

### Type a string

```bash
vncprobe type -s 10.0.0.1:5900 "show interfaces"
```

### Mouse click

```bash
# Left click (default)
vncprobe click -s 10.0.0.1:5900 400 300

# Right click
vncprobe click -s 10.0.0.1:5900 --button 3 400 300
```

Button numbers: `1` = left, `2` = middle, `3` = right

### Mouse move

```bash
vncprobe move -s 10.0.0.1:5900 400 300
```

### Wait for screen change

Wait until the screen changes from its initial state:

```bash
vncprobe wait change -s 10.0.0.1:5900 --max-wait 30 --interval 1 --threshold 0.01
```

Wait until the screen stays unchanged for a given duration (useful for animations or progress bars):

```bash
vncprobe wait stable -s 10.0.0.1:5900 --duration 3 --max-wait 60 --interval 1
```

Options for `wait change` and `wait stable`:

| Option | Default | Description |
|--------|---------|-------------|
| `--max-wait` | 30 | Maximum wait time in seconds |
| `--interval` | 1 | Polling interval in seconds |
| `--threshold` | 0.01 | Pixel difference ratio (0.0-1.0) |
| `--duration` | (required for `stable`) | Required stable duration in seconds |

### Session mode

Keep a VNC connection open and reuse it across multiple commands:

```bash
# Start a session (runs in foreground; use & for background)
vncprobe session start -s 10.0.0.1:5900 --socket /tmp/vncprobe.sock &

# Run commands via the session (no -s needed)
vncprobe capture --socket /tmp/vncprobe.sock -o screen.png
vncprobe key --socket /tmp/vncprobe.sock enter
vncprobe type --socket /tmp/vncprobe.sock "show interfaces"
vncprobe wait change --socket /tmp/vncprobe.sock

# Stop the session
vncprobe session stop --socket /tmp/vncprobe.sock
```

Options for `session start`:

| Option | Default | Description |
|--------|---------|-------------|
| `--socket` | (required) | UNIX socket path |
| `--idle-timeout` | 300 | Auto-shutdown after N seconds of inactivity (0 to disable) |

## Claude Code Integration

### Tool definition for CLAUDE.md

Add the following to your project's `CLAUDE.md` so Claude Code knows how to use vncprobe. Replace the server address and socket path with your environment's values; `<file>`, `<key>`, etc. are parameters that Claude Code fills in at each invocation:

```markdown
## Tools

vncprobe is available at /usr/local/bin/vncprobe.
Use it to interact with VM consoles via VNC.

- `vncprobe capture -s 10.0.0.1:5900 -o <file>` — Take a screenshot (PNG)
- `vncprobe key -s 10.0.0.1:5900 <key>` — Send key (e.g. enter, ctrl-c, f2)
- `vncprobe type -s 10.0.0.1:5900 "<text>"` — Type a string
- `vncprobe click -s 10.0.0.1:5900 <x> <y>` — Left click at coordinates
- `vncprobe move -s 10.0.0.1:5900 <x> <y>` — Move mouse
- `vncprobe wait change -s 10.0.0.1:5900` — Wait until screen changes
- `vncprobe wait stable -s 10.0.0.1:5900 --duration <sec>` — Wait until screen stops changing
- `vncprobe session start -s 10.0.0.1:5900 --socket /tmp/vnc.sock` — Start persistent session
- `vncprobe session stop --socket /tmp/vnc.sock` — Stop session

When using a session, pass `--socket /tmp/vnc.sock` instead of `-s 10.0.0.1:5900` for all commands.
```

### Workflow: BIOS/UEFI setup

1. `vncprobe session start -s 10.0.0.1:5900 --socket /tmp/vnc.sock &`
2. `vncprobe capture --socket /tmp/vnc.sock -o screen.png` — take a screenshot
3. Claude Code reads the PNG and recognizes the current screen (e.g. BIOS main menu)
4. `vncprobe key --socket /tmp/vnc.sock f2` — navigate to a menu item
5. `vncprobe wait change --socket /tmp/vnc.sock` — wait for the screen to update
6. `vncprobe capture --socket /tmp/vnc.sock -o screen.png` — verify the result
7. Repeat steps 3-6 until the desired configuration is applied
8. `vncprobe session stop --socket /tmp/vnc.sock`

### Workflow: Network device CLI

1. `vncprobe session start -s 10.0.0.1:5900 --socket /tmp/vnc.sock &`
2. `vncprobe capture --socket /tmp/vnc.sock -o screen.png` — read current prompt
3. `vncprobe type --socket /tmp/vnc.sock "show interfaces"` then `vncprobe key --socket /tmp/vnc.sock enter`
4. `vncprobe wait stable --socket /tmp/vnc.sock --duration 2` — wait for output to finish
5. `vncprobe capture --socket /tmp/vnc.sock -o screen.png` — read the output
6. Claude Code analyzes the output and decides the next command
7. Repeat steps 3-6
8. `vncprobe session stop --socket /tmp/vnc.sock`

## Exit codes

| Code | Meaning |
|------|---------|
| 0 | Success |
| 1 | Argument/usage error |
| 2 | Connection error |
| 3 | Operation error (includes timeout for `wait`) |

## Testing

All tests run offline using a built-in fake VNC server. No external VNC server required.

```bash
go test ./...
```

## Project structure

```
vncprobe/
├── main.go           # Entry point, subcommand dispatch
├── e2e_test.go       # End-to-end tests
├── cmd/              # CLI command handlers
│   ├── root.go       # Global flag parsing, usage
│   ├── capture.go    # capture command
│   ├── key.go        # key command
│   ├── typecmd.go    # type command
│   ├── click.go      # click command
│   ├── move.go       # move command
│   ├── wait.go       # wait command
│   └── session.go    # session command
├── vnc/              # VNC client logic
│   ├── client.go     # VNCClient interface
│   ├── realclient.go # kward/go-vnc implementation
│   ├── keymap.go     # Key name to keysym mapping
│   ├── input.go      # Key/mouse input helpers
│   ├── capture.go    # Screenshot capture + PNG save
│   ├── compare.go    # Image comparison (DiffRatio)
│   └── wait.go       # WaitForChange, WaitForStable
├── session/          # Session server/client
│   ├── protocol.go   # Request/Response types
│   ├── server.go     # UNIX socket server
│   └── client.go     # UNIX socket client
├── testutil/         # Test infrastructure
│   └── fakeserver.go # Fake RFB 003.008 server
└── testdata/
    └── expected.png  # Test image (64x64)
```

---

# vncprobe

VNC経由で画面キャプチャ・キー入力・マウス操作を行う単独バイナリのCLIツール。Claude CodeによるVMコンソールなどの自動操作用。

## インストール

[GitHub Releases](https://github.com/tjst-t/vncprobe/releases/latest) からビルド済みバイナリをダウンロード:

```bash
# 例: Linux amd64
curl -Lo vncprobe "https://github.com/tjst-t/vncprobe/releases/latest/download/vncprobe-$(curl -sL https://api.github.com/repos/tjst-t/vncprobe/releases/latest | grep -oP '\"tag_name\": \"\K[^\"]+' )-linux-amd64"
chmod +x vncprobe
sudo mv vncprobe /usr/local/bin/
```

`go install` でインストール:

```bash
go install github.com/tjst-t/vncprobe@latest
```

ソースからビルドする場合:

```bash
git clone https://github.com/tjst-t/vncprobe.git
cd vncprobe
go build -o vncprobe .
```

## 使い方

```
vncprobe <command> [options]

Commands:
  capture   画面キャプチャしてPNGで保存
  key       キー入力を送信
  type      文字列をタイプ
  click     マウスクリック
  move      マウス移動
  wait      画面変化の待機
  session   VNCセッション管理

Global Options:
  -s, --server    VNCサーバアドレス（--socket未使用時は必須）
  -p, --password  VNCパスワード
  --timeout       接続タイムアウト秒数（デフォルト: 10）
  --socket        セッションソケット経由で接続
```

### 画面キャプチャ

```bash
vncprobe capture -s 10.0.0.1:5900 -o screen.png
```

### キー入力送信

```bash
# 単一キー
vncprobe key -s 10.0.0.1:5900 enter

# 修飾キーの組み合わせ
vncprobe key -s 10.0.0.1:5900 ctrl-c
vncprobe key -s 10.0.0.1:5900 alt-f4
vncprobe key -s 10.0.0.1:5900 ctrl-alt-delete
```

対応キー: `enter`, `tab`, `escape`, `backspace`, `delete`, `space`, `up`, `down`, `left`, `right`, `home`, `end`, `pageup`, `pagedown`, `insert`, `f1`-`f12`

修飾キー: `ctrl`, `alt`, `shift`, `super`, `meta`

### 文字列入力

```bash
vncprobe type -s 10.0.0.1:5900 "show interfaces"
```

### マウスクリック

```bash
# 左クリック（デフォルト）
vncprobe click -s 10.0.0.1:5900 400 300

# 右クリック
vncprobe click -s 10.0.0.1:5900 --button 3 400 300
```

ボタン番号: `1` = 左, `2` = 中, `3` = 右

### マウス移動

```bash
vncprobe move -s 10.0.0.1:5900 400 300
```

### 画面変化の待機

画面が変化するまで待機:

```bash
vncprobe wait change -s 10.0.0.1:5900 --max-wait 30 --interval 1 --threshold 0.01
```

画面が一定時間変化しなくなるまで待機（アニメーションやプログレスバーの完了待ち）:

```bash
vncprobe wait stable -s 10.0.0.1:5900 --duration 3 --max-wait 60 --interval 1
```

`wait change` / `wait stable` 共通オプション:

| オプション | デフォルト | 説明 |
|-----------|-----------|------|
| `--max-wait` | 30 | 最大待機時間（秒） |
| `--interval` | 1 | ポーリング間隔（秒） |
| `--threshold` | 0.01 | 差分ピクセル割合の閾値（0.0〜1.0） |
| `--duration` | （`stable`では必須） | 安定と判定する連続時間（秒） |

### セッションモード

VNC接続を維持して複数コマンドで再利用:

```bash
# セッション開始（フォアグラウンド実行。バックグラウンドにするには & を付ける）
vncprobe session start -s 10.0.0.1:5900 --socket /tmp/vncprobe.sock &

# セッション経由でコマンド実行（-s 不要）
vncprobe capture --socket /tmp/vncprobe.sock -o screen.png
vncprobe key --socket /tmp/vncprobe.sock enter
vncprobe type --socket /tmp/vncprobe.sock "show interfaces"
vncprobe wait change --socket /tmp/vncprobe.sock

# セッション終了
vncprobe session stop --socket /tmp/vncprobe.sock
```

`session start` のオプション:

| オプション | デフォルト | 説明 |
|-----------|-----------|------|
| `--socket` | （必須） | UNIXソケットパス |
| `--idle-timeout` | 300 | 無操作時の自動終了秒数（0で無効） |

## Claude Code 連携

### CLAUDE.md へのツール定義例

プロジェクトの `CLAUDE.md` に以下を追加すると、Claude Code が vncprobe を使えるようになります。サーバアドレスとソケットパスは環境に合わせて書き換えてください。`<file>` や `<key>` 等は呼び出し時に Claude Code が埋めるパラメータです:

```markdown
## Tools

vncprobe is available at /usr/local/bin/vncprobe.
Use it to interact with VM consoles via VNC.

- `vncprobe capture -s 10.0.0.1:5900 -o <file>` — スクリーンショット（PNG）
- `vncprobe key -s 10.0.0.1:5900 <key>` — キー送信（例: enter, ctrl-c, f2）
- `vncprobe type -s 10.0.0.1:5900 "<text>"` — 文字列入力
- `vncprobe click -s 10.0.0.1:5900 <x> <y>` — 座標クリック
- `vncprobe move -s 10.0.0.1:5900 <x> <y>` — マウス移動
- `vncprobe wait change -s 10.0.0.1:5900` — 画面変化を待機
- `vncprobe wait stable -s 10.0.0.1:5900 --duration <sec>` — 画面安定を待機
- `vncprobe session start -s 10.0.0.1:5900 --socket /tmp/vnc.sock` — セッション開始
- `vncprobe session stop --socket /tmp/vnc.sock` — セッション終了

セッション使用時は -s 10.0.0.1:5900 の代わりに --socket /tmp/vnc.sock を指定。
```

### ワークフロー例: BIOS/UEFI 設定操作

1. `vncprobe session start -s 10.0.0.1:5900 --socket /tmp/vnc.sock &`
2. `vncprobe capture --socket /tmp/vnc.sock -o screen.png` — スクリーンショット取得
3. Claude Code が PNG を読み、画面の状態を認識（例: BIOSメインメニュー）
4. `vncprobe key --socket /tmp/vnc.sock f2` — メニュー操作
5. `vncprobe wait change --socket /tmp/vnc.sock` — 画面変化を待つ
6. `vncprobe capture --socket /tmp/vnc.sock -o screen.png` — 結果を確認
7. 目的の設定が完了するまでステップ3〜6を繰り返す
8. `vncprobe session stop --socket /tmp/vnc.sock`

### ワークフロー例: ネットワーク機器の CLI 操作

1. `vncprobe session start -s 10.0.0.1:5900 --socket /tmp/vnc.sock &`
2. `vncprobe capture --socket /tmp/vnc.sock -o screen.png` — 現在のプロンプトを確認
3. `vncprobe type --socket /tmp/vnc.sock "show interfaces"` → `vncprobe key --socket /tmp/vnc.sock enter`
4. `vncprobe wait stable --socket /tmp/vnc.sock --duration 2` — 出力完了を待つ
5. `vncprobe capture --socket /tmp/vnc.sock -o screen.png` — 出力結果を読み取る
6. Claude Code が出力を分析し、次のコマンドを決定
7. ステップ3〜6を繰り返す
8. `vncprobe session stop --socket /tmp/vnc.sock`

## 終了コード

| コード | 意味 |
|--------|------|
| 0 | 成功 |
| 1 | 引数・使用方法エラー |
| 2 | 接続エラー |
| 3 | 操作エラー（`wait` のタイムアウトを含む） |

## テスト

全テストは内蔵のフェイクVNCサーバを使用しオフラインで実行可能。外部VNCサーバ不要。

```bash
go test ./...
```

## プロジェクト構成

```
vncprobe/
├── main.go           # エントリポイント、サブコマンド振り分け
├── e2e_test.go       # E2Eテスト
├── cmd/              # CLIコマンドハンドラ
│   ├── root.go       # グローバルフラグ解析、ヘルプ表示
│   ├── capture.go    # captureコマンド
│   ├── key.go        # keyコマンド
│   ├── typecmd.go    # typeコマンド
│   ├── click.go      # clickコマンド
│   ├── move.go       # moveコマンド
│   ├── wait.go       # waitコマンド
│   └── session.go    # sessionコマンド
├── vnc/              # VNCクライアントロジック
│   ├── client.go     # VNCClientインターフェース
│   ├── realclient.go # kward/go-vnc実装
│   ├── keymap.go     # キー名→keysymマッピング
│   ├── input.go      # キー・マウス入力ヘルパー
│   ├── capture.go    # スクリーンキャプチャ・PNG保存
│   ├── compare.go    # 画像比較（DiffRatio）
│   └── wait.go       # WaitForChange, WaitForStable
├── session/          # セッションサーバ/クライアント
│   ├── protocol.go   # Request/Response型定義
│   ├── server.go     # UNIXソケットサーバ
│   └── client.go     # UNIXソケットクライアント
├── testutil/         # テストインフラ
│   └── fakeserver.go # フェイクRFB 003.008サーバ
└── testdata/
    └── expected.png  # テスト用画像（64x64）
```
