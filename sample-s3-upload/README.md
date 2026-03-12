# S3 フォルダアップロードツール

このプログラムは、ローカルフォルダ内のすべてのファイルを AWS S3 バケットにアップロードする Go アプリケーションです。

## 機能

- ローカルディレクトリ内のすべてのファイルを再帰的にスキャン
- 各ファイルを S3 にアップロード
- 相対パス構造を S3 でも維持
- オプションで S3 プレフィックスを指定可能
- アップロードの進捗表示
- エラーハンドリングとサマリー表示

## 前提条件

- Go 1.21 以上
- AWS 認証情報が設定済み（AWS CLI、環境変数、IAM ロールなど）
- S3 バケットへの書き込み権限

## インストール

```bash
# 依存関係をインストール
go mod tidy
```

## 使用方法

### 基本的な使用方法

```bash
# test-filesフォルダ内のファイルをmy-bucketにアップロード
go run main.go -bucket my-bucket -dir ./test-files
```

### オプション付きの使用方法

```bash
# プレフィックス付きでアップロード
go run main.go -bucket my-bucket -dir ./test-files -prefix uploads/

# 別のリージョンとプロファイルを指定
go run main.go -bucket my-bucket -dir ./test-files -region us-west-2 -profile dev
```

### ヘルプの表示

```bash
go run main.go -help
```

## コマンドラインオプション

| オプション | 説明                                 | デフォルト値   | 必須 |
| ---------- | ------------------------------------ | -------------- | ---- |
| `-bucket`  | S3 バケット名                        | -              | ✅   |
| `-dir`     | アップロードするローカルディレクトリ | -              | ✅   |
| `-prefix`  | S3 でのプレフィックス（オプション）  | ""             | ❌   |
| `-region`  | AWS リージョン                       | ap-northeast-1 | ❌   |
| `-profile` | AWS プロファイル                     | default        | ❌   |
| `-help`    | ヘルプを表示                         | false          | ❌   |

## 例

### 例 1: 基本的なアップロード

```bash
go run main.go -bucket my-test-bucket -dir ./test-files
```

この例では、`./test-files`フォルダ内のすべてのファイルが`my-test-bucket`にアップロードされます。

### 例 2: プレフィックス付きアップロード

```bash
go run main.go -bucket my-test-bucket -dir ./test-files -prefix documents/2024/
```

この例では、ファイルが S3 の`documents/2024/`以下にアップロードされます。

### 例 3: 本番環境での使用

```bash
go run main.go -bucket production-bucket -dir ./production-files -region us-east-1 -profile prod
```

## フォルダ構造の例

ローカルフォルダ構造:

```
test-files/
├── sample1.txt
├── sample2.txt
├── data.json
└── subfolder/
    └── nested-file.txt
```

S3 でのアップロード結果（プレフィックスなし）:

```
s3://my-bucket/
├── sample1.txt
├── sample2.txt
├── data.json
└── subfolder/
    └── nested-file.txt
```

S3 でのアップロード結果（プレフィックス: `uploads/`）:

```
s3://my-bucket/uploads/
├── sample1.txt
├── sample2.txt
├── data.json
└── subfolder/
    └── nested-file.txt
```

## エラーハンドリング

- 存在しないディレクトリを指定した場合、エラーメッセージを表示して終了
- 個別ファイルのアップロードに失敗した場合、エラーを表示して次のファイルに進む
- 最終的にアップロード成功/失敗のサマリーを表示

## セキュリティ注意事項

- AWS 認証情報を適切に管理してください
- S3 バケットのアクセス権限を適切に設定してください
- 機密ファイルを誤ってアップロードしないよう注意してください

## トラブルシューティング

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

指定した S3 バケットへの書き込み権限があるか確認してください。

### ファイルが見つからない

```
Error: directory does not exist
```

指定したローカルディレクトリが存在するか確認してください。
