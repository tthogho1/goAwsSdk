delete_s3_object.go — Usage

This script deletes S3 objects listed in a file (one key per line).

Examples:

- Dry run (no deletes):

  go run delete_s3_object.go -bucket my-bucket -keyfile keys.txt -region ap-northeast-1 -dryrun

- Perform deletes:

  go run delete_s3_object.go -bucket my-bucket -keyfile keys.txt -region ap-northeast-1

- Use a specific AWS profile:

  go run delete_s3_object.go -bucket my-bucket -keyfile keys.txt -profile dev

Notes:

- `keys.txt` should contain one object key per line. Keys may include `/` characters.
- Ensure your AWS credentials are available (environment, `~/.aws/credentials`, or specified profile).
- Test with `-dryrun` before performing actual deletes.
