// copy_flatten_and_delete.go
//
// Reads a file with S3 object keys (one per line) like:
//   8hSLxqFG0EM/8hSLxqFG0EM.m4a
// For each key the program performs:
//   1) Copy: s3://<bucket>/<key> -> s3://<bucket>/<prefix>.<ext>
//      where <prefix> is the part before the first '/' and <ext> is the file extension
//      Example: 8hSLxqFG0EM/8hSLxqFG0EM.m4a -> 8hSLxqFG0EM.m4a
//   2) Delete the original object
//
// Usage examples:
//   go run copy_flatten_and_delete.go -bucket audio4input -keyfile keys.txt -dryrun
//   go run copy_flatten_and_delete.go -bucket audio4input -keyfile keys.txt -profile dev

package main

import (
	"bufio"
	"flag"
	"fmt"
	"log"
	"net/url"
	"os"
	"path"
	"strings"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
)

func main() {
	bucket := flag.String("bucket", "audio4input", "S3 bucket name (required)")
	keyFile := flag.String("keyfile", "keys.txt", "Path to file containing object keys to process (one per line) (required)")
	region := flag.String("region", "ap-northeast-1", "AWS region")
	profile := flag.String("profile", "", "AWS profile name (optional)")
	dryRun := flag.Bool("dryrun", false, "If set, will only print actions without performing them")
	flag.Parse()

	if *bucket == "" || *keyFile == "" {
		fmt.Println("Usage:")
		flag.PrintDefaults()
		os.Exit(1)
	}

	sessOpts := session.Options{
		Config: aws.Config{
			Region: aws.String(*region),
		},
	}
	if *profile != "" {
		sessOpts.Profile = *profile
	}

	sess, err := session.NewSessionWithOptions(sessOpts)
	if err != nil {
		log.Fatalf("failed to create AWS session: %v", err)
	}
	svc := s3.New(sess)

	f, err := os.Open(*keyFile)
	if err != nil {
		log.Fatalf("failed to open key file: %v", err)
	}
	defer f.Close()

	scanner := bufio.NewScanner(f)
	line := 0
	for scanner.Scan() {
		line++
		key := strings.TrimSpace(scanner.Text())
		if key == "" {
			continue
		}

		// expect at least one '/'
		parts := strings.SplitN(key, "/", 2)
		if len(parts) < 2 {
			log.Printf("skipping unexpected key (no '/'): %s (line %d)", key, line)
			continue
		}
		prefix := parts[0]
		ext := path.Ext(key)
		if ext == "" {
			// fallback: use extension of filename part
			ext = path.Ext(parts[1])
		}
		destKey := prefix + ext

		if *dryRun {
			fmt.Printf("[dryrun] aws s3 cp s3://%s/%s s3://%s/%s\n", *bucket, key, *bucket, destKey)
			fmt.Printf("[dryrun] aws s3 rm s3://%s/%s\n", *bucket, key)
			continue
		}

		// Copy object. CopySource must have the source key URL-encoded.
		encodedKey := url.PathEscape(key)
		copySource := fmt.Sprintf("%s/%s", *bucket, encodedKey)
		_, err := svc.CopyObject(&s3.CopyObjectInput{
			Bucket:     aws.String(*bucket),
			CopySource: aws.String(copySource),
			Key:        aws.String(destKey),
		})
		if err != nil {
			log.Printf("copy failed for %s (line %d): %v", key, line, err)
			continue
		}

		// Optionally you may wait for the object to exist before deleting the source.
		// For simplicity we proceed to delete; in unstable networks add a HeadObject waiter.
		_, err = svc.DeleteObject(&s3.DeleteObjectInput{
			Bucket: aws.String(*bucket),
			Key:    aws.String(key),
		})
		if err != nil {
			log.Printf("delete failed for %s (line %d): %v", key, line, err)
			continue
		}

		fmt.Printf("Processed: %s -> %s\n", key, destKey)
	}
	if err := scanner.Err(); err != nil {
		log.Fatalf("error reading key file: %v", err)
	}
	fmt.Println("Completed.")
}
