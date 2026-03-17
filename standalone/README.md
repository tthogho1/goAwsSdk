# EC2 Instances Viewer

AWS EC2インスタンス情報をGUIテーブル形式で表示するデスクトップアプリケーション。

## 機能

- AWSプロファイルを画面から選択・入力
- `awsctrl` コマンドでEC2インスタンス情報を取得
- テーブル形式で表示（ID, Status, Type, PrivateIP, PublicIP, Cost, Name）

## 前提条件

- Go (module-enabled) — 本リポジトリは Go toolchain 1.23 以降で開発されています。
- `awsctrl` 実行バイナリ（またはパス）が利用可能であること（後述の環境変数で指定）。
- ネイティブGUIは Gio (`gioui.org`) を使用しています（C コンパイラは不要）。

## セットアップ

1. 必要に応じて `.env` ファイルを編集し、`AWSCTRL_PATH` に `awsctrl` 実行ファイルのパスを設定：

```env
AWSCTRL_PATH=C:\path\to\awsctrl.exe
```

2. 依存関係のインストール：

```bash
cd standalone
go mod tidy
```

3. 実行（開発時）：

```bash
cd standalone
go run .
```

または `run.bat` をダブルクリックして起動できます。

## ビルド

```bash
cd standalone
go build -o ec2viewer.exe .
```

## 使い方

1. アプリを起動
2. プロファイル欄でAWSプロファイルを選択または入力（初期値: `default`）
3. 「取込」ボタンを押すと EC2 インスタンス一覧がテーブルに表示されます

## 備考

- クリップボード機能は `github.com/atotto/clipboard` を利用しています。プラットフォームによっては追加のツール（例: Linux の `xclip`/`xsel`）が必要です。
- Windows 上でバイナリを上書きする際は、実行中のプロセスを停止するか、別名でビルドしてください（例: `-o ec2viewer_alt.exe`）。

もし他に追記してほしい操作手順やスクリーンショットがあれば教えてください。
