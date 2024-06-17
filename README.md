# Cntrol EC2 by go Aws Sdk
Get EC2 information and up/down ec2 instance

## Usage
Usage: goAwsSdk [options]

Examples:  
AwsCtrl -c describe -t EC2 -i <instanceid>  
AwsCtrl -c appRunner -t EC2 -i <instanceid>  
AwsCtrl -c up -t EC2 -i <instanceid>  
AwsCtrl -c up -t appRunner -s <service arn>  
AwsCtrl -c S3download -b <bucketName> -t <localdir>  

options detail:  
  -b string  
    	bucket name  
  -c string  
    	command : describe | up |down (default "describe")  
  -h string  
    	help
  -i string  
    	instance id  
  -profile string  
    	Specifiy Credential profile (default "default")  
  -region string  
    	Specify AWS region (default "ap-northeast-1")  
  -s string  
    	service arn  
  -t string  
    	target : EC2 | appRunner | lodaldir (default "EC2")  

## Example
1.Describe EC2 Instance
```
AwsCtrl
```

2.Up EC2 instance
```
AwsCtrl -c up -i <instance_id>
```

3.Down EC2 instance
```
AwsCtrl -c down -i <instance_id>
```
