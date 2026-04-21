# OBS Scheduler

OBS (Open Broadcaster Software) の録画開始・終了を指定した時間に自動で行う CLI ツールです。
OBS WebSocket v5 プロトコルを使用して OBS を制御します。

## 前提条件

- **Go**: 1.26.2 以上
- **OBS Studio**:
  - WebSocket サーバー設定が有効になっていること
  - WebSocket v5 プロトコルを使用

## セットアップ

1. リポジトリをクローンします。
2. 依存関係を解決します。

```bash
go mod download
```

## 設定

プロジェクトルートに `.env` ファイルを作成し、OBS WebSocket の接続情報を設定してください。

`.env` の例:

```env
OBS_ADDR=ws://localhost:4455
OBS_PASSWORD=your_password
```

| 変数名         | 説明                               | 例                    |
| :------------- | :--------------------------------- | :-------------------- |
| `OBS_ADDR`     | OBS WebSocket サーバーのアドレス   | `ws://localhost:4455` |
| `OBS_PASSWORD` | OBS WebSocket サーバーのパスワード | `secret`              |

## 使い方

### コマンドライン引数

以下のフラグを使用して、録画の開始・終了時刻を指定できます。時刻フォーマットは `HH:MM` (24 時間表記) です。

| フラグ   | デフォルト値 | 説明               |
| :------- | :----------- | :----------------- |
| `-start` | `08:44`      | 録画を開始する時刻 |
| `-stop`  | `10:00`      | 録画を終了する時刻 |

### 実行例

**Go コマンドで実行:**

```bash
# デフォルト設定で実行
go run cmd/obs-scheduler/main.go

# 時間を指定して実行
go run cmd/obs-scheduler/main.go -start 19:00 -stop 21:00
```

**Make コマンドで実行:**

```bash
# デフォルト設定で実行
make start-rec

# 時間を指定して実行
make start-rec START=19:00 STOP=21:00
```

## ディレクトリ構造

```
obs-scheduler/
├── cmd/
│   └── obs-scheduler/      # アプリケーションのエントリーポイント
├── internal/
│   ├── config/             # 設定読み込み
│   ├── domain/             # ドメインロジック・インターフェース
│   ├── usecase/            # ビジネスロジック (スケジューリング)
│   └── infrastructure/     # 外部インターフェース (OBS WebSocket)
├── go.mod
└── Makefile
```
