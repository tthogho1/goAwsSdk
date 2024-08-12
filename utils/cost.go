// utils.go
package utils

import (
	"fmt"
	"log"
	"strconv"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/costexplorer"
)

// get cost from Aws betweern start and end
func Cost(svc *costexplorer.CostExplorer, startDate string, endDate string) {
	// Create the input for GetCostAndUsage
	input := &costexplorer.GetCostAndUsageInput{
		TimePeriod: &costexplorer.DateInterval{
			Start: aws.String(startDate),
			End:   aws.String(endDate),
		},
		Granularity: aws.String("MONTHLY"),
		Metrics:     []*string{aws.String("UnblendedCost")},
		GroupBy: []*costexplorer.GroupDefinition{
			{
				Type: aws.String("DIMENSION"),
				Key:  aws.String("SERVICE"),
			},
		},
	}

	// Retrieve the cost and usage data
	result, err := svc.GetCostAndUsage(input)
	if err != nil {
		log.Fatalf("failed to get cost and usage, %v", err)
	}

	var total float64 = 0
	for _, resultByTime := range result.ResultsByTime {
		fmt.Printf("Time Period: %s - %s\n", *resultByTime.TimePeriod.Start, *resultByTime.TimePeriod.End)
		var amount float64 = 0
		for _, group := range resultByTime.Groups {
			amount, _ = strconv.ParseFloat(*group.Metrics["UnblendedCost"].Amount, 32)
			total = total + amount
			fmt.Printf("Service: %s, Amount: %s\n", *group.Keys[0], *group.Metrics["UnblendedCost"].Amount)
		}
	}
	fmt.Printf("Total: %f\n", total)

}
