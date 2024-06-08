package main

import (
	"awsctrl/utils"
	"flag"
	"fmt"
	"strings"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/ec2"
)

func main() {

	profile := flag.String("profile", "default", "profile")
	region := flag.String("region", "ap-northeast-1", "region")
	cmd := flag.String("c", "describe", "command")

	instansString := flag.String("i", "", "instance id")

	flag.Parse()
	fmt.Printf("profile: %s, region: %s\n", *profile, *region)

	instanceIds := strings.Split(*instansString, ",")
	sess := session.Must(session.NewSessionWithOptions(session.Options{
		Profile: *profile, // specify profile
	}))

	switch *cmd {
	case "describe":
		//  create a EC2 service client
		svc := ec2.New(sess, aws.NewConfig().WithRegion(*region))
		utils.Describe(svc)
	case "up":
		svc := ec2.New(sess, aws.NewConfig().WithRegion(*region))
		utils.Up(svc, instanceIds)
	case "down":
		svc := ec2.New(sess, aws.NewConfig().WithRegion(*region))
		utils.Down(svc, instanceIds)
	default:
		fmt.Println("unknown command")
	}

}
