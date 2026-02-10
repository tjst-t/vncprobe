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

Global Options:
  -s, --server    VNC server address (required)
  -p, --password  VNC password
  --timeout       Connection timeout in seconds (default: 10)
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

## Exit codes

| Code | Meaning |
|------|---------|
| 0 | Success |
| 1 | Argument/usage error |
| 2 | Connection error |
| 3 | Operation error |

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
│   └── move.go       # move command
├── vnc/              # VNC client logic
│   ├── client.go     # VNCClient interface
│   ├── realclient.go # kward/go-vnc implementation
│   ├── keymap.go     # Key name to keysym mapping
│   ├── input.go      # Key/mouse input helpers
│   └── capture.go    # Screenshot capture + PNG save
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

Global Options:
  -s, --server    VNCサーバアドレス（必須）
  -p, --password  VNCパスワード
  --timeout       接続タイムアウト秒数（デフォルト: 10）
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

## 終了コード

| コード | 意味 |
|--------|------|
| 0 | 成功 |
| 1 | 引数・使用方法エラー |
| 2 | 接続エラー |
| 3 | 操作エラー |

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
│   └── move.go       # moveコマンド
├── vnc/              # VNCクライアントロジック
│   ├── client.go     # VNCClientインターフェース
│   ├── realclient.go # kward/go-vnc実装
│   ├── keymap.go     # キー名→keysymマッピング
│   ├── input.go      # キー・マウス入力ヘルパー
│   └── capture.go    # スクリーンキャプチャ・PNG保存
├── testutil/         # テストインフラ
│   └── fakeserver.go # フェイクRFB 003.008サーバ
└── testdata/
    └── expected.png  # テスト用画像（64x64）
```