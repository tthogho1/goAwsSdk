package utils

import (
	"fmt"
	"log"
	"os"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/s3"

	"path/filepath"

	"github.com/aws/aws-sdk-go/service/s3/s3manager"
)

func Download(svc *s3.S3, bucket string, localdir string) {
	var continuationToken *string
	downloader := s3manager.NewDownloaderWithClient(svc)

	for {
		resp, err := svc.ListObjectsV2(&s3.ListObjectsV2Input{
			Bucket:            aws.String(bucket),
			ContinuationToken: continuationToken,
		})
		if err != nil {
			fmt.Println("failed to list objects", err)
			return
		}

		for _, obj := range resp.Contents {
			fmt.Println("Downloading", *obj.Key)

			downloadObject(downloader, bucket, *obj.Key, localdir)
		}

		if !aws.BoolValue(resp.IsTruncated) {
			break
		}
		continuationToken = resp.NextContinuationToken
	}
}

func Upload(svc *s3.S3, bucket string, localdir string) {

	files, err := os.ReadDir(localdir)
	if err != nil {
		log.Fatal(err)
	}

	for _, file := range files {
		UploadFile(svc, bucket, filepath.Join(localdir, file.Name()))
	}

}

func UploadFile(svc *s3.S3, bucket string, localfile string) {
	uploader := s3manager.NewUploaderWithClient(svc)

	file, err := os.Open(localfile)
	if err != nil {
		fmt.Printf("Could not open file: %s\n", err)
		return
	}
	defer file.Close()

	fileName := filepath.Base(localfile)
	// ファイルをアップロード
	result, err := uploader.Upload(&s3manager.UploadInput{
		Bucket: aws.String(bucket),
		Key:    aws.String(fileName),
		Body:   file,
	})

	if err != nil {
		fmt.Printf("Failed to Upload: %s\n", err)
	} else {
		fmt.Printf("Success to Upload: %s\n", result.Location)
	}
}

func downloadObject(downloader *s3manager.Downloader, bucket string, key string, localdir string) {
	localFilePath := filepath.Join(localdir, key)
	file, err := os.Create(localFilePath)
	if err != nil {
		fmt.Println("failed to create file", err)
		return
	}
	defer file.Close()

	_, err2 := downloader.Download(file, &s3.GetObjectInput{
		Bucket: aws.String(bucket),
		Key:    aws.String(key),
	})
	if err2 != nil {
		fmt.Println("failed to get object", err)
		return
	}
}
