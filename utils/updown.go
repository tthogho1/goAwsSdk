package utils

import (
	"fmt"
	"log"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/apprunner"
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/aws/aws-sdk-go/service/ecs"
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

func DownECS(svc *ecs.ECS, pattern *string) {
	listTasksInput := &ecs.ListTasksInput{
		Cluster: pattern,
	}
	listTasksOutput, err := svc.ListTasks(listTasksInput)
	if err != nil {
		log.Fatalf("Failed to list tasks: %v", err)
	}

	for _, taskArn := range listTasksOutput.TaskArns {
		input := &ecs.StopTaskInput{
			Cluster: pattern,
			Task:    taskArn,
			Reason:  aws.String("Stopped by Go SDK"),
		}

		result, err := svc.StopTask(input)
		if err != nil {
			log.Fatalf("Failed to stop task: %v", err)
		}

		fmt.Printf("downed task: %s\n", *result.Task.TaskArn)
		fmt.Printf("status: %s\n", *result.Task.LastStatus)
	}
}

func UpAppRunner(svc *apprunner.AppRunner, serviceArn *string) {
	// up instance with specified instance id
	input := &apprunner.ResumeServiceInput{
		ServiceArn: aws.String(*serviceArn),
	}
	result, err := svc.ResumeService(input)
	if err != nil {
		log.Fatalf("Unable to start Service, %v", err)
	}
	fmt.Println(result)
}

func DownAppRunner(svc *apprunner.AppRunner, serviceArn *string) {
	// up instance with specified instance id
	input := &apprunner.PauseServiceInput{
		ServiceArn: aws.String(*serviceArn),
	}

	result, err := svc.PauseService(input)
	if err != nil {
		log.Fatalf("Unable to stop Service, %v", err)
	}
	fmt.Println(result)
}
