# EC2 Instances Viewer

AWS EC2インスタンス情報をGUIテーブル形式で表示するデスクトップアプリケーション。

## 機能

- AWSプロファイルを画面から選択・入力
- `awsctrl` コマンドでEC2インスタンス情報を取得
- テーブル形式で表示（ID, Status, Type, PrivateIP, PublicIP, Cost, Name）

## 前提条件

- Go 1.22 以上
- `awsctrl.exe` がビルド済みであること
- Fyne v2 の動作要件（GCCなどCコンパイラ）

## セットアップ

1. `.env` ファイルを編集し、`AWSCTRL_PATH` に `awsctrl.exe` のパスを設定：

```env
AWSCTRL_PATH=C:\path\to\awsctrl.exe
```

2. 依存関係のインストール：

```bash
cd standalone
go mod tidy
```

3. 実行：

```bash
go run main.go
```

または `run.bat` をダブルクリック。

## ビルド

```bash
go build -o ec2viewer.exe main.go
```

## 使い方

1. アプリを起動
2. プロファイル欄でAWSプロファイルを選択または入力（初期値: `default`）
3. 「取込」ボタンを押すとEC2インスタンス一覧がテーブルに表示される
