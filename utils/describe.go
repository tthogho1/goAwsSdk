// utils.go
package utils

import (
	"fmt"
	"log"

	"github.com/aws/aws-sdk-go/service/ec2"
)

func Describe(svc *ec2.EC2) {
	result, err := svc.DescribeInstances(nil)
	if err != nil {
		log.Fatalf("Unable to describe instances, %v", err)
	}

	// インスタンス情報の表
	for _, reservation := range result.Reservations {
		for _, instance := range reservation.Instances {

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
