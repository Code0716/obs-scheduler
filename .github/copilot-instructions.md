````instructions
# GitHub Copilot Instructions for obs-scheduler

## 基本情報

### プロジェクト概要

OBS (Open Broadcaster Software) の録画開始・終了を指定した時間に自動で行う CLI ツール。
OBS WebSocket v5 プロトコルを使用して OBS を制御する。

### 技術スタック

- **言語**: Go 1.25.3
- **OBS クライアント**: `github.com/andreykaipov/goobs`
- **設定読み込み**: `github.com/joho/godotenv`
- **ロギング**: `log/slog` (推奨)
- **アーキテクチャ**: Clean Architecture (Layered Architecture)

### ディレクトリ構造

```
obs-scheduler/
├── cmd/
│   └── obs-scheduler/      # アプリケーションのエントリーポイント (main.go)
├── internal/
│   ├── config/             # 設定読み込み (環境変数, フラグ)
│   ├── domain/             # ドメインロジック・インターフェース定義
│   ├── usecase/            # ビジネスロジック (スケジューリング)
│   └── infrastructure/     # 外部インターフェースの実装
│       └── obs/            # OBS WebSocket クライアント実装
├── go.mod
└── Makefile
```

## Go 言語コード生成・レビュー規則

### 基本原則

- **シンプルさ優先**: 必要以上に複雑な設計を避け、標準ライブラリを最大限活用する。
- **Context の活用**: キャンセル信号やタイムアウトの伝播には必ず `context.Context` を使用する。特に長時間待機する処理やネットワーク通信には必須。
- **構造化ロギング**: ログ出力には `log/slog` を使用し、解析可能な形式で出力する。
- **マジックナンバー禁止**: OBS の OpCode やリクエスト ID などの定数は、コード内に直接記述せず定数として定義する。

### エラーハンドリング

- エラーは無視せず、必ずハンドリングする。
- 呼び出し元にエラーを返す際は、コンテキスト情報を付与してラップする（`fmt.Errorf("...: %w", err)`）。

### 並行性

- **Graceful Shutdown**: `os.Interrupt` や `syscall.SIGTERM` を検知し、実行中の録画や接続を適切に終了・切断する処理を実装する。

## アーキテクチャ層の責務

- **cmd/obs-scheduler**:
  - 依存関係の注入 (Wiring)。
  - アプリケーションの起動と終了処理。
- **internal/config**:
  - 環境変数 (`.env`) やコマンドライン引数からの設定値のロード。
- **internal/domain**:
  - `Recorder` インターフェースの定義。
  - 外部依存を持たない純粋なドメイン定義。
- **internal/usecase**:
  - `Scheduler`: 指定時刻までの待機と `Recorder` の操作を行うビジネスロジック。
- **internal/infrastructure/obs**:
  - `domain.Recorder` の実装。
  - `goobs` ライブラリを使用した OBS との通信詳細。

## テスト生成指針

- **テーブル駆動テスト**: テストケースはテーブル駆動テスト (Table Driven Tests) で記述する。
- **モック**: `domain` パッケージで定義されたインターフェースのモックを作成し、`usecase` のテストで使用する。

## Copilot の振る舞い設定

### 言語設定

- **回答・レビュー言語**: すべての応答、コードの説明、プルリクエストのレビューコメント、コミットメッセージの提案は **日本語** で行ってください。
- 英語で質問された場合でも、文脈から日本人の開発者であると判断できる場合は日本語で回答してください。
````
