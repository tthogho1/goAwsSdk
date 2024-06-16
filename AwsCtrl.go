package main

import (
	"awsctrl/utils"
	"flag"
	"fmt"
	"strings"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/apprunner"
	"github.com/aws/aws-sdk-go/service/ec2"
)

func Usage() {
	fmt.Println()

	fmt.Println("Usage: goAwsSdk [options]")
	fmt.Println()
	fmt.Println("Examples:")
	fmt.Println("AwsCtrl -c describe -t EC2 -i <instanceid>")
	fmt.Println("AwsCtrl -c appRunner -t EC2 -i <instanceid>")
	fmt.Println("AwsCtrl -c up -t EC2 -i <instanceid>")
	fmt.Println("AwsCtrl -c up -t appRunner -s <service arn>")

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
}

func OptionParse() Options {

	Options := Options{}
	Options.profile = flag.String("profile", "default", "Specifiy Credential profile")
	Options.region = flag.String("region", "ap-northeast-1", "Specify AWS region")
	Options.cmd = flag.String("c", "describe", "command : describe | up |down")
	Options.target = flag.String("t", "EC2", "target : EC2 | appRunner")
	Options.help = flag.String("h", "help", "help")

	Options.serviceArn = flag.String("s", "", "service arn")
	Options.instansString = flag.String("i", "", "instance id")

	flag.Parse()
	fmt.Printf("profile: %s, region: %s\n", *Options.profile, *Options.region)

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
	case "describe":
		//  create a EC2 service client
		if *Options.target == "EC2" {
			svc := ec2.New(sess, aws.NewConfig().WithRegion(*Options.region))
			utils.Describe(svc)
		} else if *Options.target == "appRunner" {
			svc := apprunner.New(sess, aws.NewConfig().WithRegion(*Options.region))
			utils.DescribeAppRunner(svc)
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
		}
	default:
		Usage()
	}
}
