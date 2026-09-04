# Cntrol AWS BY GO AWS SDK

## Usage
Usage: goAwsSdk [options]

Examples:  
awsctrl -c describe -t EC2 -i <instanceid>  
awsctrl -c appRunner -t EC2 -i <instanceid>  
awsctrl -c up -t EC2 -i <instanceid>  
awsctrl -c up -t appRunner -s <service arn>  
awsctrl -c S3download -b <bucketName> -t <localdir>  
awsctrl -c describe -t ECS  
awsctrl -c create -t EC2 -key <keypair> -ec2type <ec2type> -network-interfaces <json String>  
awsctrl -c exec -exec <cmdstring> -t EC2 -i <instanceid> 
```
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
    	target : EC2 | appRunner | localdir (default "EC2")  
  -ec2type string  
      ec2type : like t2.micro  
  -nerwork-interfaces string  
    need 4 value  
       public_ip_address : true|flase  // when true public ipadress added  
       device_index : 0     // only one network-interface  
       subne_id :  <subnet-nnnnn>  
       gropus: [<security_group_id>]  // array of security group id  

     example  
      {"public_ip_address":true,"device_index":0,"subnet_id":"subnet-05f2a9a81d40d433d","groups":["sg-05c48f49dbf81efeb"]} 
      
   -key string  
      key pairename  
```


## Example
1.Describe EC2 Instance
```
awsCtrl
```

2.Up EC2 instance
```
awsctrl -c up -i <instance_id>
```

3.Down EC2 instance
```
awsctrl -c down -i <instance_id>
```
