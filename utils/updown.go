package utils

import (
	"fmt"
	"log"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/ec2"
)

func Up(svc *ec2.EC2, instanceId []string) {
	// up instance with specified instance id

	instanceIds := make([]*string, 0)
	for _, idStr := range instanceId {
		instanceIds = append(instanceIds, aws.String(idStr))
	}

	input := &ec2.StartInstancesInput{
		InstanceIds: instanceIds,
	}

	result, err := svc.StartInstances(input)
	if err != nil {
		log.Fatalf("Unable to start instance, %v", err)
	}

	fmt.Println(result)
}

func Down(svc *ec2.EC2, instanceId []string) {

	instanceIds := make([]*string, 0)
	for _, idStr := range instanceId {
		instanceIds = append(instanceIds, aws.String(idStr))
	}

	input := &ec2.StopInstancesInput{
		InstanceIds: instanceIds,
	}
	result, err := svc.StopInstances(input)

	if err != nil {
		log.Fatalf("Unable to stop instance, %v", err)
	}
	fmt.Println(result)
}
