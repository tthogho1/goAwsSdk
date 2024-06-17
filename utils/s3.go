package utils

import (
	"fmt"
	"os"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/s3"

	"path/filepath"

	"github.com/aws/aws-sdk-go/service/s3/s3manager"
)

func Download(svc *s3.S3, bucket string, localdir string) {
	resp, err := svc.ListObjects(&s3.ListObjectsInput{Bucket: aws.String(bucket)})
	if err != nil {
		fmt.Println("failed to list objects", err)
		return
	}

	for _, item := range resp.Contents {
		fmt.Println(*item.Key)

		localFilePath := filepath.Join(localdir, *item.Key)
		file, err := os.Create(localFilePath)
		if err != nil {
			fmt.Println("failed to create file", err)
			return
		}
		defer file.Close()

		downloader := s3manager.NewDownloaderWithClient(svc)
		_, err2 := downloader.Download(file, &s3.GetObjectInput{
			Bucket: aws.String(bucket),
			Key:    item.Key,
		})

		if err2 != nil {
			fmt.Println("failed to get object", err)
			return
		}

		fmt.Println("downloaded", localFilePath)
	}

}
