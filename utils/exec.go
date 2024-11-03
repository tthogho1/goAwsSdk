// utils.go
package utils

import (
	"fmt"
	"log"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/ssm"
)

// Tag名に指定されたPatternに一致するものがあればTrueを返す
func Exec(sess *session.Session, instanceId string, cmdstring string) bool {

	ssmClient := ssm.New(sess)

	output, err := ssmClient.SendCommand(&ssm.SendCommandInput{
		InstanceIds:  []*string{aws.String(instanceId)},
		DocumentName: aws.String("AWS-RunShellScript"),
		Parameters: map[string][]*string{
			"commands": {
				aws.String(cmdstring),
			},
		},
	})

	if err != nil {
		log.Fatalf("Failed to send command: %v", err)
		return false
	}

	commandID := output.Command.CommandId

	result, err := ssmClient.GetCommandInvocation(&ssm.GetCommandInvocationInput{
		CommandId:  commandID,
		InstanceId: aws.String(instanceId),
	})
	if err != nil {
		log.Fatalf("Failed to get command result: %v", err)
		return false
	}
	fmt.Printf("Command output: %s\n", *result.StandardOutputContent)

	return true
}
