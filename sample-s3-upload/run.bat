@echo off
echo S3 フォルダアップロードツール
echo ==============================
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
echo サンプル実行（test-filesフォルダをアップロード）
echo バケット名を入力してください:
set /p bucket_name="Bucket Name: "

if "%bucket_name%"=="" (
    echo バケット名が入力されていません
    pause
    exit /b 1
)

echo.
echo アップロード中...
go run main.go -bucket %bucket_name% -dir ./test-files

echo.
echo 処理が完了しました
pause
