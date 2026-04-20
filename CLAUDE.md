# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## 言語設定

回答・コード説明・コミットメッセージの提案はすべて**日本語**で行う。

## コマンド

```bash
# 依存関係のインストールとツールセットアップ
make init          # .env 初期化 + ツール一式インストール
make deps          # go mod download/tidy

# コード生成
make di            # Wire による DI コード生成
make mock          # mockgen によるモック生成
make generate      # di + mock を一括実行

# 静的解析
make lint          # golangci-lint 実行
make vuln-check    # govulncheck 実行

# 実行
make start-rec START=19:00 STOP=21:00   # 録画スケジュール起動（デフォルト: 08:44〜10:00）

# テスト（標準 go コマンド）
go test ./...
go test ./internal/usecase/...   # 特定パッケージのみ
```

## アーキテクチャ

OBS の録画を指定時刻に自動開始・停止する CLI ツール。OBS WebSocket v5 プロトコルで OBS を制御する。

### 層の責務

- **`cmd/obs-scheduler/`** — エントリーポイント。Wire による依存注入、起動・終了処理。
- **`internal/config/`** — `.env` ファイルと CLI フラグ（`-start`, `-stop`, `-skip-launch`）から設定をロード。
- **`internal/domain/`** — `Recorder`・`AppLifecycle` インターフェース定義。外部依存ゼロ。
- **`internal/usecase/`** — `Scheduler`：指定時刻まで待機し `Recorder` を操作するコアロジック。リトライ・グレースフルシャットダウンを含む。
- **`internal/infrastructure/obs/`** — `domain.Recorder` の実装。WebSocket v5 の Hello/Identify 認証、StartRecord/StopRecord を実装。
- **`internal/infrastructure/lifecycle/`** — プラットフォーム別（darwin/windows/linux）の OBS 起動・終了実装。ビルドタグで切り替え。

### 実行フロー

1. `.env` + CLI フラグから設定読み込み
2. Wire DI で OBS クライアント・ライフサイクルマネージャを初期化
3. 開始時刻の 10 秒前に OBS 起動（プラットフォーム別）
4. OBS WebSocket に接続（15 回リトライ、2 秒間隔）
5. 開始時刻に録画開始（10 回リトライ）
6. 終了時刻に録画停止 → OBS 終了
7. SIGTERM/SIGINT でグレースフルシャットダウン

## コーディング規則

- `context.Context` を必ず伝播させる（長時間待機・ネットワーク通信）
- ロギングは `log/slog` の構造化ログ
- エラーは `fmt.Errorf("操作名: %w", err)` でラップして返す
- マジックナンバー禁止（OpCode・定数は `const` で定義）
- テストはテーブル駆動テスト形式、`domain` インターフェースのモックを `usecase` テストで使用

## 環境設定

`.env.example` をコピーして `.env` を作成し、`OBS_ADDR` と `OBS_PASSWORD` を設定する。
