@echo off
echo S3 ファイル一覧表示ツール
echo =============================
echo.

REM 依存関係のインストール
echo 依存関係をインストール中...
go mod tidy
if %errorlevel% neq 0 (
    echo エラー: 依存関係のインストールに失敗しました
    pause
    exit /b 1
)
echo.

REM サンプル実行
echo S3バケットのファイル一覧を表示します
echo バケット名を入力してください:
set /p bucket_name="Bucket Name: "

if "%bucket_name%"=="" (
    echo バケット名が入力されていません
    pause
    exit /b 1
)

echo.
echo プレフィックスを指定しますか？ (オプション)
echo 例: uploads/, logs/2024/, など
set /p prefix_name="Prefix (空の場合はEnter): "

echo.
echo ファイル一覧を取得中...

if "%prefix_name%"=="" (
    go run main.go -bucket %bucket_name%
) else (
    go run main.go -bucket %bucket_name% -prefix %prefix_name%
)

echo.
echo 処理が完了しました
pause
