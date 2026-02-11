# 設計: 画面変化の待機・ポーリング機能（wait コマンド）

## 概要

`wait` サブコマンドを新設し、VNC 画面の変化を検知して待機する機能を提供する。

## CLI 設計

`wait` の下に `change` と `stable` の2つのサブコマンドを持つ。

```
vncprobe wait change  -s <addr> [--max-wait 30] [--interval 1] [--threshold 0.01]
vncprobe wait stable  -s <addr> --duration <sec> [--max-wait 30] [--interval 1] [--threshold 0.01]
```

### wait change

画面が前回のキャプチャから変化するまで待機する。

- 初回キャプチャを撮り、以降 `--interval` 間隔で再キャプチャして比較
- 差分ピクセル割合が `--threshold` を超えたら「変化した」と判定して終了（exit 0）
- `--max-wait` に達したら exit 3

### wait stable

画面が `--duration` 秒間変化しなくなるまで待機する（アニメーション/プログレスバー完了待ち用）。

- `--interval` 間隔でキャプチャし、前回と比較
- 差分が threshold 以下の状態が `--duration` 秒間続いたら「安定した」と判定して終了（exit 0）
- `--max-wait` に達したら exit 3
- `--duration` は必須引数

### 共通オプション

| オプション | デフォルト | 説明 |
|-----------|-----------|------|
| `--max-wait` | 30 | 最大待機時間（秒） |
| `--interval` | 1 | ポーリング間隔（秒） |
| `--threshold` | 0.01 | 変化判定の閾値（差分ピクセル割合、0.0〜1.0） |

### 将来の拡張

`wait reference -s <addr> --file ref.png` を後から追加できる構造にしておく。`wait` のサブコマンド分岐に1ケース追加するだけで対応可能。

## アーキテクチャ

### 新規ファイル

```
vnc/compare.go       — 画像比較ロジック
vnc/compare_test.go  — 画像比較のユニットテスト
vnc/wait.go          — WaitForChange, WaitForStable 関数
vnc/wait_test.go     — wait ロジックのユニットテスト（モッククライアント使用）
cmd/wait.go          — CLI ハンドラ（サブコマンド分岐 + フラグ解析）
```

### 画像比較ロジック（`vnc/compare.go`）

```go
// DiffRatio returns the fraction of pixels that differ between two images.
// Images must have the same dimensions; returns error otherwise.
// A pixel is "different" if any channel (R, G, B) differs.
func DiffRatio(a, b image.Image) (float64, error)
```

- 両画像の全ピクセルを走査し、RGBA 値が異なるピクセルの割合を返す
- 画像サイズが異なる場合はエラー
- 比較は厳密な一致（チャンネル単位の完全一致）。ノイズ耐性は threshold で調整する設計

### wait ロジック（`vnc/wait.go`）

```go
type WaitOptions struct {
    Timeout   time.Duration
    Interval  time.Duration
    Threshold float64
}

// WaitForChange captures repeatedly until screen differs from initial capture.
// Returns nil on change detected, error on timeout.
func WaitForChange(client VNCClient, opts WaitOptions) error

// WaitForStable captures repeatedly until screen stays unchanged for duration.
// Returns nil when stable, error on timeout.
func WaitForStable(client VNCClient, opts WaitOptions, stableDuration time.Duration) error
```

- `WaitForChange`: 初回キャプチャを基準に、interval ごとに再キャプチャ → `DiffRatio` で比較 → threshold 超えたら return nil
- `WaitForStable`: interval ごとにキャプチャ → 前回と比較 → threshold 以下が stableDuration 続いたら return nil
- タイムアウトは `time.After` で管理
- タイムアウト時は専用のエラー型を返す（`cmd` 層で exit code 3 に変換するため）

### CMD ハンドラ（`cmd/wait.go`）

```go
func RunWait(client vnc.VNCClient, args []string) error
```

- `args[0]` でサブコマンドを判定（`change` / `stable`）
- サブコマンドなし or 不明の場合はエラー + ヘルプ表示
- `flag.FlagSet` で各サブコマンドのフラグを解析

### main.go の変更

- `run()` の command switch に `"wait"` を追加
- `cmd.RunWait(client, cmdArgs)` を呼び出す

### Usage の更新

- `cmd/root.go` の `Usage()` に `wait` コマンドの説明を追加

## テスト戦略

### ユニットテスト: vnc/compare_test.go

- 同一画像 → DiffRatio = 0.0
- 完全に異なる画像 → DiffRatio = 1.0
- 一部だけ異なる画像 → 正しい割合
- サイズ不一致 → エラー

### ユニットテスト: vnc/wait_test.go

モッククライアント（`client_test.go` の既存パターン）を使用:

- `WaitForChange`: 3回目のキャプチャで変化する → 成功
- `WaitForChange`: 変化しない → タイムアウト
- `WaitForStable`: 最初は変化し続け、途中から安定 → 成功
- `WaitForStable`: 安定しない → タイムアウト

### E2E テスト: e2e_test.go

フェイクサーバを拡張して、途中で画面が変わるシナリオをテスト:

- `FakeVNCServer` に `SetImage(img image.Image)` メソッドを追加
  - mutex で保護し、`sendFramebufferUpdate` が参照する `s.img` を差し替え可能にする
- テスト内で goroutine から `time.Sleep` 後に `SetImage` を呼ぶことで画面変化をシミュレート

テストケース:
- `TestE2EWaitChange`: サーバ画像を途中で変更 → exit 0
- `TestE2EWaitChangeTimeout`: サーバ画像を変更しない → exit 3
- `TestE2EWaitStable`: サーバ画像を2回変更後に安定 → exit 0

## エラーハンドリング

- タイムアウト: `vnc.ErrTimeout` を返し、`cmd` 層で exit 3 にマップ
- VNC 接続エラー（キャプチャ失敗）: そのままエラーを伝搬 → exit 3
- 引数エラー（サブコマンドなし、不正な値）: `cmd` 層でエラー → main.go で exit 1 にするため `cmd.RunWait` の前にバリデーション

### exit code の注意

現在の `main.go` は `cmd.Run*` のエラーを全て exit 3 にマップしている。`wait` の引数エラー（サブコマンド不正など）は exit 1 にすべき。

対応方法: `cmd` パッケージに `UsageError` 型を定義し、`main.go` の `run()` で `errors.As` で判定して exit 1 を返す。既存コマンドの `flag.FlagSet` のエラーも同様に扱えるが、スコープを限定して `wait` だけに適用する。

## 実装順序

1. `vnc/compare.go` + テスト
2. `vnc/wait.go` + テスト（モック使用）
3. `cmd/wait.go`（CLI ハンドラ）
4. `main.go` + `cmd/root.go` 更新
5. `testutil/fakeserver.go` に `SetImage` 追加
6. E2E テスト
7. README 更新
