// utils.go
package utils

import (
	"fmt"
	"log"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/apprunner"
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/aws/aws-sdk-go/service/ecs"
)

func Describe(svc *ec2.EC2, pattern *string) {
	result, err := svc.DescribeInstances(nil)
	if err != nil {
		log.Fatalf("Unable to describe instances, %v", err)
	}

	// インスタンス情報の表
	for _, reservation := range result.Reservations {
		for _, instance := range reservation.Instances {

			if !IsPatternMatch(instance.Tags, pattern) {
				continue
			}

			publicIp := "None" // instance.PublicIpAddress
			if instance.PublicIpAddress != nil {
				publicIp = *instance.PublicIpAddress
			}
			fmt.Printf(" ID: %s,  %s, %s, %s, %s", *instance.InstanceId,
				*instance.State.Name, *instance.InstanceType, *instance.PrivateIpAddress, publicIp)

			for _, tag := range instance.Tags {
				fmt.Printf("  %s: %s", *tag.Key, *tag.Value)
			}
			fmt.Printf("\n")
		}
	}
}

func DescribeAppRunner(svc *apprunner.AppRunner, pattern *string) {

	input := &apprunner.ListServicesInput{}
	result, err := svc.ListServices(input)

	if err != nil {
		log.Fatalf("Unable to list services, %v", err)
		return
	}

	for _, service := range result.ServiceSummaryList {

		if !IsServicePatternMatch(service.ServiceName, pattern) {
			continue
		}

		fmt.Printf("Service Name:%s ID:%s ARN:%s Status:%s\n",
			*service.ServiceName, *service.ServiceId, *service.ServiceArn, *service.Status)
	}

}

func DescribeECS(svc *ecs.ECS, pattern *string) {
	listTasksInput := &ecs.ListTasksInput{
		Cluster: pattern,
	}

	listTasksOutput, err := svc.ListTasks(listTasksInput)
	if err != nil {
		log.Fatalf("Failed to list tasks: %v", err)
	}

	for _, taskArn := range listTasksOutput.TaskArns {
		fmt.Println(*taskArn)
	}
}

func DescribeAMI(svc *ec2.EC2, pattern *string) {

	var filter = "amazon"
	if *pattern != "" {
		filter = *pattern
	}

	params := &ec2.DescribeImagesInput{
		Owners:     []*string{aws.String(filter)},
		MaxResults: aws.Int64(10),
	}

	res, err := svc.DescribeImages(params)
	if err != nil {
		fmt.Println(err.Error())
	}
	for _, image := range res.Images {
		fmt.Println(*image.ImageId, *image.Name, *image.CreationDate)
	}
}

func SecurityGroupInfo(result *ec2.DescribeSecurityGroupsOutput) {

	if len(result.SecurityGroups) > 0 {
		sg := result.SecurityGroups[0]
		fmt.Printf("Security Group Details:\n")
		fmt.Printf("  ID: %s\n", *sg.GroupId)
		fmt.Printf("  Name: %s\n", *sg.GroupName)
		fmt.Printf("  Description: %s\n", *sg.Description)
		fmt.Printf("  VPC ID: %s\n", *sg.VpcId)

		fmt.Println("Inbound Rules:")
		for _, rule := range sg.IpPermissions {
			fmt.Printf("  Protocol: %s\n", *rule.IpProtocol)
			fmt.Printf("  Port Range: %d - %d\n", *rule.FromPort, *rule.ToPort)
			for _, ipRange := range rule.IpRanges {
				fmt.Printf("    CIDR: %s\n", *ipRange.CidrIp)
			}
		}

		fmt.Println("Outbound Rules:")
		for _, rule := range sg.IpPermissionsEgress {
			fmt.Printf("  Protocol: %s\n", *rule.IpProtocol)
			fmt.Printf("  Port Range: %d - %d\n", *rule.FromPort, *rule.ToPort)
			for _, ipRange := range rule.IpRanges {
				fmt.Printf("    CIDR: %s\n", *ipRange.CidrIp)
			}
		}
	} else {
		fmt.Println("No security group found with the specified ID")
	}
}

func DescribeSecurityGroup(svc *ec2.EC2, groupID *string) {

	var params *ec2.DescribeSecurityGroupsInput
	if *groupID != "" {
		params = &ec2.DescribeSecurityGroupsInput{
			GroupIds: []*string{
				aws.String(*groupID),
			},
		}
	} else {
		params = nil
	}

	res, err := svc.DescribeSecurityGroups(params)
	if err != nil {
		fmt.Println(err.Error())
	}

	for _, group := range res.SecurityGroups {
		fmt.Println(*group.GroupId, *group.GroupName, *group.Description)
	}
}
