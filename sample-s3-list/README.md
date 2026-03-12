# S3 ファイル一覧表示ツール

このプログラムは、指定した AWS S3 バケット内のファイル一覧を表示する Go アプリケーションです。

## 機能

- S3 バケット内のファイル一覧を表示
- プレフィックス（フォルダパス）によるフィルタリング
- 再帰的/非再帰的な検索モード
- ファイルサイズ、最終更新日時、ストレージクラスの表示
- 人間が読みやすいファイルサイズ表示
- 合計ファイル数とサイズのサマリー表示

## 前提条件

- Go 1.21 以上
- AWS 認証情報が設定済み（AWS CLI、環境変数、IAM ロールなど）
- S3 バケットへの読み取り権限

## インストール

```bash
# 依存関係をインストール
go mod tidy
```

## 使用方法

### 基本的な使用方法

```bash
# バケット全体のファイル一覧を表示
go run main.go -bucket my-bucket
```

### オプション付きの使用方法

```bash
# 特定のプレフィックス（フォルダ）のファイル一覧を表示
go run main.go -bucket my-bucket -prefix uploads/

# 最大100ファイルまで表示
go run main.go -bucket my-bucket -max 100

# 非再帰的（フォルダ直下のみ）で表示
go run main.go -bucket my-bucket -recursive=false

# 別のリージョンとプロファイルを指定
go run main.go -bucket my-bucket -region us-west-2 -profile dev
```

### ヘルプの表示

```bash
go run main.go -help
```

## コマンドラインオプション

| オプション   | 説明                              | デフォルト値   | 必須 |
| ------------ | --------------------------------- | -------------- | ---- |
| `-bucket`    | S3 バケット名                     | -              | ✅   |
| `-prefix`    | S3 プレフィックス（フォルダパス） | ""             | ❌   |
| `-max`       | 表示する最大ファイル数            | 1000           | ❌   |
| `-recursive` | 再帰的に検索するか                | true           | ❌   |
| `-region`    | AWS リージョン                    | ap-northeast-1 | ❌   |
| `-profile`   | AWS プロファイル                  | myregion       | ❌   |
| `-help`      | ヘルプを表示                      | false          | ❌   |

## 出力例

### 基本的な出力

```
Configuration:
  Profile: myregion
  Region: ap-northeast-1
  Bucket: audio4input
  S3 Prefix:
  Max Keys: 1000
  Recursive: true

Listing objects in bucket 'audio4input'...

📄 document1.pdf                                   2.5 MB 2024-01-15 10:30:45 STANDARD
📄 image1.jpg                                      1.2 MB 2024-01-14 15:22:33 STANDARD
📄 video1.mp4                                     45.8 MB 2024-01-13 09:15:22 STANDARD
📄 uploads/data.json                               15.3 KB 2024-01-12 14:45:12 STANDARD
📄 uploads/backup/archive.zip                    125.6 MB 2024-01-11 08:30:00 GLACIER

Summary:
  Files: 5
  Total Size: 175.1 MB
```

### 非再帰的な出力（フォルダ表示）

```
📁 uploads/
📁 images/
📁 documents/
📄 root-file.txt                                  1.5 KB 2024-01-15 12:00:00 STANDARD

Summary:
  Files: 1
  Folders: 3
  Total Size: 1.5 KB
```

## 機能詳細

### プレフィックスフィルタリング

`-prefix`オプションを使用して、特定のフォルダ内のファイルのみを表示できます：

```bash
# uploads/ フォルダ内のファイルのみ表示
go run main.go -bucket my-bucket -prefix uploads/

# 2024年のログファイルのみ表示
go run main.go -bucket my-bucket -prefix logs/2024/
```

### 再帰的検索の制御

- `-recursive=true`（デフォルト）：すべてのサブフォルダ内のファイルも表示
- `-recursive=false`：指定したレベルのフォルダとファイルのみ表示

### ファイルサイズ表示

ファイルサイズは人間が読みやすい形式で表示されます：

- B（バイト）
- KB（キロバイト）
- MB（メガバイト）
- GB（ギガバイト）
- TB（テラバイト）

### ストレージクラス

各ファイルのストレージクラス（STANDARD、GLACIER、DEEP_ARCHIVE など）も表示されます。

## エラーハンドリング

### AWS 認証エラー

```
Error: failed to create AWS session
```

AWS 認証情報が正しく設定されているか確認してください：

```bash
aws configure list
```

### バケットアクセスエラー

```
Error: Access Denied
```

指定した S3 バケットへの読み取り権限があるか確認してください。

### バケットが存在しない

```
Error: The specified bucket does not exist
```

バケット名が正しいか、指定したリージョンにバケットが存在するか確認してください。

## パフォーマンス

- 大量のファイルがある場合は、`-max`オプションで表示するファイル数を制限することを推奨
- プレフィックスを使用してスコープを絞ることで、処理時間を短縮可能
- ページネーション対応により、メモリ使用量を最適化

## セキュリティ注意事項

- AWS 認証情報を適切に管理してください
- 必要最小限の権限（S3 読み取り権限）のみを付与してください
- 機密バケットへのアクセス時は特に注意してください
