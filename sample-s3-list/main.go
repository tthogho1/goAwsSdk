package main

import (
	"flag"
	"fmt"
	"os"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
)

type Config struct {
	Profile    string
	Region     string
	BucketName string
	S3Prefix   string
	MaxKeys    int64
	Recursive  bool
}

func main() {
	// コマンドライン引数の定義
	profile := flag.String("profile", "myregion", "AWS profile name")
	region := flag.String("region", "ap-northeast-1", "AWS region")
	bucket := flag.String("bucket", "audio4input", "S3 bucket name (required)")
	s3Prefix := flag.String("prefix", "", "S3 prefix (optional)")
	maxKeys := flag.Int64("max", 1000, "Maximum number of keys to list")
	recursive := flag.Bool("recursive", true, "List files recursively")
	help := flag.Bool("help", false, "Show help")

	flag.Parse()

	if *help {
		fmt.Println("S3 File List Tool")
		fmt.Println()
		fmt.Println("Usage:")
		fmt.Println("  go run main.go -bucket <bucket-name> [options]")
		fmt.Println()
		fmt.Println("Options:")
		flag.PrintDefaults()
		fmt.Println()
		fmt.Println("Examples:")
		fmt.Println("  go run main.go -bucket my-bucket")
		fmt.Println("  go run main.go -bucket my-bucket -prefix uploads/")
		fmt.Println("  go run main.go -bucket my-bucket -max 100 -recursive=false")
		fmt.Println("  go run main.go -bucket my-bucket -region us-west-2 -profile dev")
		return
	}

	// 必須パラメータのチェック
	if *bucket == "" {
		fmt.Println("Error: bucket name is required")
		fmt.Println("Use -help for usage information")
		os.Exit(1)
	}

	config := Config{
		Profile:    *profile,
		Region:     *region,
		BucketName: *bucket,
		S3Prefix:   *s3Prefix,
		MaxKeys:    *maxKeys,
		Recursive:  *recursive,
	}

	fmt.Printf("Configuration:\n")
	fmt.Printf("  Profile: %s\n", config.Profile)
	fmt.Printf("  Region: %s\n", config.Region)
	fmt.Printf("  Bucket: %s\n", config.BucketName)
	fmt.Printf("  S3 Prefix: %s\n", config.S3Prefix)
	fmt.Printf("  Max Keys: %d\n", config.MaxKeys)
	fmt.Printf("  Recursive: %t\n", config.Recursive)
	fmt.Println()

	err := listS3Files(config)
	if err != nil {
		fmt.Printf("Error listing S3 files: %v\n", err)
		os.Exit(1)
	}

	fmt.Println("Listing completed successfully!")
}

func listS3Files(config Config) error {
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

	// S3サービスクライアントの作成
	svc := s3.New(sess)

	var delimiter string
	if !config.Recursive {
		delimiter = "/"
	}

	// リストオブジェクトの入力パラメータ
	input := &s3.ListObjectsV2Input{
		Bucket:  aws.String(config.BucketName),
		MaxKeys: aws.Int64(config.MaxKeys),
	}

	if config.S3Prefix != "" {
		input.Prefix = aws.String(config.S3Prefix)
	}

	if delimiter != "" {
		input.Delimiter = aws.String(delimiter)
	}

	fmt.Printf("Listing objects in bucket '%s'", config.BucketName)
	if config.S3Prefix != "" {
		fmt.Printf(" with prefix '%s'", config.S3Prefix)
	}
	if !config.Recursive {
		fmt.Printf(" (non-recursive)")
	}
	fmt.Println("...")
	fmt.Println()

	totalSize := int64(0)
	fileCount := 0
	folderCount := 0

	// ページネーション対応でファイル一覧を取得
	err = svc.ListObjectsV2Pages(input, func(page *s3.ListObjectsV2Output, lastPage bool) bool {
		// フォルダ（Common Prefixes）の表示
		for _, prefix := range page.CommonPrefixes {
			fmt.Printf("📁 %s\n", aws.StringValue(prefix.Prefix))
			folderCount++
		}

		// ファイル（Objects）の表示
		for _, obj := range page.Contents {
			key := aws.StringValue(obj.Key)
			size := aws.Int64Value(obj.Size)
			lastModified := aws.TimeValue(obj.LastModified)
			storageClass := aws.StringValue(obj.StorageClass)

			// サイズを人間が読みやすい形式に変換
			sizeStr := formatSize(size)
			
			// 日時をフォーマット
			timeStr := lastModified.Format("2006-01-02 15:04:05")

			fmt.Printf("📄 %-50s %10s %s %s\n", key, sizeStr, timeStr, storageClass)
			
			totalSize += size
			fileCount++
		}

		return true // 続行する
	})

	if err != nil {
		return fmt.Errorf("failed to list objects: %v", err)
	}

	fmt.Println()
	fmt.Printf("Summary:\n")
	fmt.Printf("  Files: %d\n", fileCount)
	if !config.Recursive {
		fmt.Printf("  Folders: %d\n", folderCount)
	}
	fmt.Printf("  Total Size: %s\n", formatSize(totalSize))

	return nil
}

func formatSize(bytes int64) string {
	const unit = 1024
	if bytes < unit {
		return fmt.Sprintf("%d B", bytes)
	}
	div, exp := int64(unit), 0
	for n := bytes / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}
	return fmt.Sprintf("%.1f %cB", float64(bytes)/float64(div), "KMGTPE"[exp])
}
