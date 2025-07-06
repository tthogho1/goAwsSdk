package main

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3/s3manager"
)

type Config struct {
	Profile    string
	Region     string
	BucketName string
	LocalDir   string
	S3Prefix   string
}

func main() {
	// コマンドライン引数の定義
	profile := flag.String("profile", "myregion", "AWS profile name")
	region := flag.String("region", "ap-northeast-1", "AWS region")
	bucket := flag.String("bucket", "audio4input", "S3 bucket name (required)")
	localDir := flag.String("dir", "C:\\temp\\mp4", "Local directory to upload (required)")
	s3Prefix := flag.String("prefix", "", "S3 prefix (optional)")
	help := flag.Bool("help", false, "Show help")

	flag.Parse()

	if *help {
		fmt.Println("S3 Folder Upload Tool")
		fmt.Println()
		fmt.Println("Usage:")
		fmt.Println("  go run main.go -bucket <bucket-name> -dir <local-directory> [options]")
		fmt.Println()
		fmt.Println("Options:")
		flag.PrintDefaults()
		fmt.Println()
		fmt.Println("Examples:")
		fmt.Println("  go run main.go -bucket my-bucket -dir ./test-files")
		fmt.Println("  go run main.go -bucket my-bucket -dir ./test-files -prefix uploads/")
		fmt.Println("  go run main.go -bucket my-bucket -dir ./test-files -region us-west-2 -profile dev")
		return
	}

	// 必須パラメータのチェック
	if *bucket == "" {
		fmt.Println("Error: bucket name is required")
		fmt.Println("Use -help for usage information")
		os.Exit(1)
	}

	if *localDir == "" {
		fmt.Println("Error: local directory is required")
		fmt.Println("Use -help for usage information")
		os.Exit(1)
	}

	// ディレクトリの存在確認
	if _, err := os.Stat(*localDir); os.IsNotExist(err) {
		fmt.Printf("Error: directory '%s' does not exist\n", *localDir)
		os.Exit(1)
	}

	config := Config{
		Profile:    *profile,
		Region:     *region,
		BucketName: *bucket,
		LocalDir:   *localDir,
		S3Prefix:   *s3Prefix,
	}

	fmt.Printf("Configuration:\n")
	fmt.Printf("  Profile: %s\n", config.Profile)
	fmt.Printf("  Region: %s\n", config.Region)
	fmt.Printf("  Bucket: %s\n", config.BucketName)
	fmt.Printf("  Local Directory: %s\n", config.LocalDir)
	fmt.Printf("  S3 Prefix: %s\n", config.S3Prefix)
	fmt.Println()

	err := uploadFolder(config)
	if err != nil {
		fmt.Printf("Error uploading folder: %v\n", err)
		os.Exit(1)
	}

	fmt.Println("Upload completed successfully!")
}

func uploadFolder(config Config) error {
	// AWS セッションの作成
	sess, err := session.NewSessionWithOptions(session.Options{
		Config: aws.Config{
			Region: aws.String(config.Region),
		},
		Profile: config.Profile,
	})
	if err != nil {
		return fmt.Errorf("failed to create AWS session: %v", err)
	}

	// S3アップローダーの作成
	uploader := s3manager.NewUploader(sess)

	// フォルダ内のファイルを取得
	files, err := getFilesInDirectory(config.LocalDir)
	if err != nil {
		return fmt.Errorf("failed to get files in directory: %v", err)
	}

	if len(files) == 0 {
		fmt.Println("No files found in the specified directory")
		return nil
	}

	fmt.Printf("Found %d files to upload:\n", len(files))

	// 各ファイルをアップロード
	successCount := 0
	for _, filePath := range files {
		err := uploadFile(uploader, config, filePath)
		if err != nil {
			fmt.Printf("  ❌ Failed to upload %s: %v\n", filePath, err)
		} else {
			successCount++
		}
	}

	fmt.Printf("\nUpload Summary:\n")
	fmt.Printf("  Success: %d/%d files\n", successCount, len(files))
	fmt.Printf("  Failed: %d/%d files\n", len(files)-successCount, len(files))

	return nil
}

func getFilesInDirectory(dir string) ([]string, error) {
	var files []string

	err := filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		// ディレクトリは除外
		if !info.IsDir() {
			files = append(files, path)
		}

		return nil
	})

	return files, err
}

func uploadFile(uploader *s3manager.Uploader, config Config, filePath string) error {
	// ファイルを開く
	file, err := os.Open(filePath)
	if err != nil {
		return fmt.Errorf("failed to open file: %v", err)
	}
	defer file.Close()

	// 相対パスを取得（ローカルディレクトリからの相対パス）
	relPath, err := filepath.Rel(config.LocalDir, filePath)
	if err != nil {
		return fmt.Errorf("failed to get relative path: %v", err)
	}

	// S3キーを構築
	s3Key := relPath
	if config.S3Prefix != "" {
		// プレフィックスがある場合は結合
		s3Key = strings.TrimSuffix(config.S3Prefix, "/") + "/" + relPath
	}

	// WindowsのパスセパレータをS3用に変換
	s3Key = strings.ReplaceAll(s3Key, "\\", "/")

	fmt.Printf("  📁 Uploading: %s -> s3://%s/%s\n", relPath, config.BucketName, s3Key)

	// ファイルをアップロード
	result, err := uploader.Upload(&s3manager.UploadInput{
		Bucket: aws.String(config.BucketName),
		Key:    aws.String(s3Key),
		Body:   file,
	})

	if err != nil {
		return err
	}

	fmt.Printf("  ✅ Uploaded to: %s\n", result.Location)
	return nil
}
