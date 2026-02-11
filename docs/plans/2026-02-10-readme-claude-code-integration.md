# 設計: README に Claude Code 連携ワークフロー例を追加

## 概要

README に `## Claude Code Integration` セクションを追加し、Claude Code から vncprobe をどう使うかのワークフローを示す。wait コマンド・session モードの説明も含める。

## 変更対象

- `README.md`（英語セクション・日本語セクション両方）

## 追加する内容

### 1. CLAUDE.md ツール定義の例

プロジェクトの CLAUDE.md に vncprobe をツールとして定義するサンプルを示す。全コマンド（capture, key, type, click, move, wait, session）を含む。

### 2. ステップバイステップのワークフロー例

#### ユースケースA: BIOS/UEFI 設定画面の操作

1. `vncprobe session start` でセッション開始
2. `vncprobe capture --socket` で画面を撮る
3. Claude Code が PNG を読み、現在の画面状態を認識
4. `vncprobe click/key --socket` で操作
5. `vncprobe wait change --socket` で画面変化を待つ
6. `vncprobe capture --socket` で再度撮って確認
7. 繰り返し → `vncprobe session stop` で終了

#### ユースケースB: ネットワーク機器の CLI 操作

1. `vncprobe session start` でセッション開始
2. `vncprobe capture --socket` でプロンプトを確認
3. `vncprobe type --socket` + `vncprobe key --socket enter` でコマンド実行
4. `vncprobe wait stable --socket --duration 2` で出力が安定するまで待つ
5. `vncprobe capture --socket` で結果を読み取る
6. 繰り返し → `vncprobe session stop` で終了

### 3. コマンドリファレンスの更新

README の Usage セクションに以下を追加:
- `wait` コマンド（`wait change`, `wait stable`）の説明と使用例
- `session` コマンド（`session start`, `session stop`）の説明と使用例
- `--socket` グローバルオプションの説明

### 4. Project structure の更新

```
├── session/          # Session server/client
│   ├── server.go
│   ├── client.go
│   └── protocol.go
```

vnc/ に `compare.go`, `wait.go` を追加。

## 実装手順

1. README.md の英語セクション Usage に wait/session コマンドの使い方を追加
2. `## Claude Code Integration` セクションを追加（Testing の前）
3. Project structure を更新
4. 日本語セクションにも同等の内容を追加
