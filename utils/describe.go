// utils.go
package utils

import (
	"fmt"
	"log"

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
