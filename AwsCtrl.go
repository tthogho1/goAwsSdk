package main

import (
	"awsctrl/types"
	"awsctrl/utils"
	"flag"
	"fmt"
	"strings"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/apprunner"
	"github.com/aws/aws-sdk-go/service/costexplorer"
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/aws/aws-sdk-go/service/ecs"
	"github.com/aws/aws-sdk-go/service/s3"
)

func Usage() {
	fmt.Println()

	fmt.Println("Usage: goAwsSdk [options]")
	fmt.Println()
	fmt.Println("Examples:")
	fmt.Println("awsctrl -c describe -t EC2 -i <instanceid> <-p> <pattern>")
	fmt.Println("awsctrl -c describe -t AMI ")
	fmt.Println("awsctrl -c describe -t SecurityGroup ")
	fmt.Println("awsctrl -c appRunner -t EC2 -i <instanceid>")
	fmt.Println("awsctrl -c up -t EC2 -i <instanceid>")
	fmt.Println("awsctrl -c up -t appRunner -s <service arn>")
	fmt.Println("awsctrl -c S3download -b <bucketName> -t <localdir>")
	fmt.Println("awsctrl -c S3upload -b <bucketName> -f <localfile>")
	fmt.Println("awsctrl -c S3upload -b <bucketName> -t <localdir>")
	fmt.Println("awsctrl -c describe -t ECS -n <cluster>")
	fmt.Println("awsctrl -c cost -start <YYYY-MM-DD> -end <YYYY-MM-DD>")
	fmt.Println("awsctrl -c create -t EC2 -ami <ami id> -type <instance type>")

	fmt.Println()
	fmt.Println("options detail:")
	flag.PrintDefaults()
}

type Options struct {
	profile       *string
	region        *string
	cmd           *string
	target        *string
	help          *string
	serviceArn    *string
	instansString *string
	bucketName    *string
	pattern       *string
	name          *string
	file          *string
	start         *string
	end           *string
	ec2type       *string
	amiString     *string
	keyPair       *string
	groupID       *string
	network       *types.NetWorkIF
}

func OptionParse() Options {

	Options := Options{}
	Options.profile = flag.String("profile", "default", "Specifiy Credential profile")
	Options.region = flag.String("region", "ap-northeast-1", "Specify AWS region")
	Options.cmd = flag.String("c", "describe", "command : describe | up |down | S3download | AMI ")
	Options.target = flag.String("t", "EC2", "target : EC2 | appRunner | local dir | ECS | SecurityGroup")
	Options.pattern = flag.String("p", "", "regression pattern for Names of Tag")
	Options.file = flag.String("f", "", "upload file name")
	Options.name = flag.String("n", "", "cluster name")
	Options.help = flag.String("h", "", "help")

	Options.start = flag.String("start", "", "start time YYYY-MM-DD")
	Options.end = flag.String("end", "", "end time YYYY-MM-DD")

	Options.serviceArn = flag.String("s", "", "service arn")
	Options.instansString = flag.String("i", "", "instance id")
	Options.bucketName = flag.String("b", "", "bucket name")

	Options.ec2type = flag.String("ec2type", "t2.micro", "instance type")
	Options.keyPair = flag.String("key", "", "key pair name")
	Options.amiString = flag.String("ami", "", "ami id")
	Options.groupID = flag.String("g", "", "group id")

	fmt.Printf("profile: %s, region: %s\n", *Options.profile, *Options.region)

	// json部分
	netWorkFlag := &types.NetWorkIF{Data: make(map[string]interface{})}
	flag.Var(netWorkFlag, "network-interfaces", "JSON形式の入力")
	flag.Parse()

	Options.network = netWorkFlag

	//
	fmt.Printf("Parsed JSON: %+v\n", netWorkFlag.Data)

	return Options
}

func main() {

	Options := OptionParse()

	instanceIds := strings.Split(*Options.instansString, ",")
	sess := session.Must(session.NewSessionWithOptions(session.Options{
		Profile: *Options.profile, // specify profile
	}))

	if *Options.help == "help" {
		flag.PrintDefaults()
		Usage()
		return
	}

	switch *Options.cmd {
	case "create":
		if *Options.target == "EC2" {
			svc := ec2.New(sess, aws.NewConfig().WithRegion(*Options.region))
			utils.CreateInstance(svc, *Options.amiString, *Options.ec2type, *Options.keyPair, *Options.network)
		} else {
			fmt.Println("create target " + *Options.target + " is Invalid")
		}
	case "cost":
		svc := costexplorer.New(sess)
		utils.Cost(svc, *Options.start, *Options.end)
	case "describe":
		//  create a EC2 service client
		if *Options.target == "EC2" {
			svc := ec2.New(sess, aws.NewConfig().WithRegion(*Options.region))
			utils.Describe(svc, Options.pattern)

		} else if *Options.target == "appRunner" {
			svc := apprunner.New(sess, aws.NewConfig().WithRegion(*Options.region))
			utils.DescribeAppRunner(svc, Options.pattern)

		} else if *Options.target == "ECS" {
			svc := ecs.New(sess, aws.NewConfig().WithRegion(*Options.region))
			utils.DescribeECS(svc, Options.name)

		} else if *Options.target == "AMI" {
			svc := ec2.New(sess, aws.NewConfig().WithRegion(*Options.region))
			utils.DescribeAMI(svc, Options.pattern)

		} else if *Options.target == "SecurityGroup" {
			svc := ec2.New(sess, aws.NewConfig().WithRegion(*Options.region))
			utils.DescribeSecurityGroup(svc, Options.groupID)

		} else {
			fmt.Println("describe target " + *Options.target + " is Invalid")
		}
	case "up":
		if *Options.target == "EC2" {
			svc := ec2.New(sess, aws.NewConfig().WithRegion(*Options.region))
			utils.Up(svc, instanceIds)
		} else if *Options.target == "appRunner" {
			svc := apprunner.New(sess, aws.NewConfig().WithRegion(*Options.region))
			utils.UpAppRunner(svc, Options.serviceArn)
		}
	case "down":
		if *Options.target == "EC2" {
			svc := ec2.New(sess, aws.NewConfig().WithRegion(*Options.region))
			utils.Down(svc, instanceIds)
		} else if *Options.target == "appRunner" {
			svc := apprunner.New(sess, aws.NewConfig().WithRegion(*Options.region))
			utils.DownAppRunner(svc, Options.serviceArn)
		} else if *Options.target == "ECS" {
			svc := ecs.New(sess, aws.NewConfig().WithRegion(*Options.region))
			utils.DownECS(svc, Options.name)
		}
	case "S3download":
		sess, errs3 := session.NewSession(&aws.Config{
			Region: aws.String(*Options.region),
		})
		if errs3 != nil {
			fmt.Println(errs3)
		}

		svc := s3.New(sess)
		utils.Download(svc, *Options.bucketName, *Options.target)
	case "S3upload":
		sess, errs3 := session.NewSessionWithOptions(session.Options{
			Config: aws.Config{
				Region: aws.String(*Options.region),
			},
			Profile: *Options.profile,
		})

		if errs3 != nil {
			fmt.Println(errs3)
		}

		svc := s3.New(sess)
		if *Options.file != "" {
			utils.UploadFile(svc, *Options.bucketName, *Options.file)
		} else {
			utils.Upload(svc, *Options.bucketName, *Options.target)
		}

	default:
		Usage()
	}
}
