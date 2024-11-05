// utils.go
package utils

import (
	"fmt"
	"log"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/ssm"
)

func getCommandResult(ssmClient *ssm.SSM, instanceId string, commandId *string) (*ssm.GetCommandInvocationOutput, error) {

	fmt.Printf("Command ID: %s\n", *commandId)
	var result *ssm.GetCommandInvocationOutput
	var err error
	for {
		time.Sleep(1 * time.Second) // Wait for 5 seconds before checking
		result, err = ssmClient.GetCommandInvocation(&ssm.GetCommandInvocationInput{
			CommandId:  commandId,
			InstanceId: aws.String(instanceId),
		})
		if err != nil {
			return nil, err
		}

		switch *result.Status {
		case "Success":
			return result, nil
		case "Failed", "Cancelled", "TimedOut":
			return nil, fmt.Errorf("command execution failed with status: %s", *result.Status)
		case "InProgress", "Pending":
			// Continue polling
		default:
			return nil, fmt.Errorf("unknown command status: %s", *result.Status)
		}
	}
}

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
	result, err := getCommandResult(ssmClient, instanceId, commandID)
	fmt.Printf("Command output: %s\n", *result.StandardOutputContent)

	return true
}
