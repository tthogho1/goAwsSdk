// utils.go
package utils

import (
	"fmt"
	"log"
	"sort"
	"strings"

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
			// fmt.Println(instance.Tags)
			if !IsPatternMatch(instance.Tags, pattern) {
				continue
			}

			publicIp := "None" // instance.PublicIpAddress
			if instance.PublicIpAddress != nil {
				publicIp = *instance.PublicIpAddress
			}

			privateIp := "None"
			var ips []string
			for _, ni := range instance.NetworkInterfaces {
				for _, addr := range ni.PrivateIpAddresses {
					if addr.PrivateIpAddress != nil {
						ips = append(ips, *addr.PrivateIpAddress)
					}
				}
			}
			if len(ips) > 0 {
				privateIp = strings.Join(ips, ";")
			} else if instance.PrivateIpAddress != nil {
				privateIp = *instance.PrivateIpAddress
			}

			instanceId := "None"
			if instance.InstanceId != nil {
				instanceId = *instance.InstanceId
			}

			stateName := "None"
			if instance.State != nil && instance.State.Name != nil {
				stateName = *instance.State.Name
			}

			instanceType := "None"
			if instance.InstanceType != nil {
				instanceType = *instance.InstanceType
			}

			fmt.Printf(" ID: %s,  %s, %s, %s, %s", instanceId,
				stateName, instanceType, privateIp, publicIp)

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

// DescribeSubnets prints the subnets in a VPC (all subnets if vpcID is
// empty), sorted by VPC, AZ and CIDR. Output format (parsed by Ec2Viewer;
// the name is last because it may contain commas):
//
//	Subnet: subnet-xxx, vpc-xxx, 10.0.1.0/24, ap-northeast-1a, available, 251, default=false, publicIP=true, Name=web-a
func DescribeSubnets(svc *ec2.EC2, vpcID string) error {
	input := &ec2.DescribeSubnetsInput{}
	if vpcID != "" {
		input.Filters = []*ec2.Filter{{Name: aws.String("vpc-id"), Values: []*string{aws.String(vpcID)}}}
	}
	var subnets []*ec2.Subnet
	err := svc.DescribeSubnetsPages(input, func(page *ec2.DescribeSubnetsOutput, _ bool) bool {
		subnets = append(subnets, page.Subnets...)
		return true
	})
	if err != nil {
		return err
	}

	sort.Slice(subnets, func(i, j int) bool {
		a, b := subnets[i], subnets[j]
		if va, vb := strOr(a.VpcId, ""), strOr(b.VpcId, ""); va != vb {
			return va < vb
		}
		if za, zb := strOr(a.AvailabilityZone, ""), strOr(b.AvailabilityZone, ""); za != zb {
			return za < zb
		}
		return strOr(a.CidrBlock, "") < strOr(b.CidrBlock, "")
	})

	for _, s := range subnets {
		name := "-"
		for _, tag := range s.Tags {
			if tag.Key != nil && *tag.Key == "Name" && tag.Value != nil {
				name = *tag.Value
				break
			}
		}
		fmt.Printf("Subnet: %s, %s, %s, %s, %s, %d, default=%t, publicIP=%t, Name=%s\n",
			strOr(s.SubnetId, "None"), strOr(s.VpcId, "-"), strOr(s.CidrBlock, "-"),
			strOr(s.AvailabilityZone, "-"), strOr(s.State, "-"),
			aws.Int64Value(s.AvailableIpAddressCount),
			aws.BoolValue(s.DefaultForAz), aws.BoolValue(s.MapPublicIpOnLaunch), name)
	}
	return nil
}

// DescribeInstanceDetail prints detailed information for a single EC2 instance:
// VPC, Subnet, attached Security Groups and all Tags. Output is formatted as
// simple "Key: Value" lines so it can be parsed by callers (e.g. Ec2Viewer).
func DescribeInstanceDetail(svc *ec2.EC2, instanceID *string) {
	if instanceID == nil || *instanceID == "" {
		fmt.Println("instance id is required")
		return
	}

	params := &ec2.DescribeInstancesInput{
		InstanceIds: []*string{aws.String(*instanceID)},
	}

	result, err := svc.DescribeInstances(params)
	if err != nil {
		fmt.Println(err.Error())
		return
	}

	for _, reservation := range result.Reservations {
		for _, instance := range reservation.Instances {
			id := "None"
			if instance.InstanceId != nil {
				id = *instance.InstanceId
			}
			fmt.Printf("InstanceId: %s\n", id)

			vpcID := "None"
			if instance.VpcId != nil {
				vpcID = *instance.VpcId
			}
			fmt.Printf("VpcId: %s\n", vpcID)

			subnetID := "None"
			if instance.SubnetId != nil {
				subnetID = *instance.SubnetId
			}
			fmt.Printf("SubnetId: %s\n", subnetID)

			for _, sg := range instance.SecurityGroups {
				groupID := "None"
				if sg.GroupId != nil {
					groupID = *sg.GroupId
				}
				groupName := "None"
				if sg.GroupName != nil {
					groupName = *sg.GroupName
				}
				fmt.Printf("SecurityGroup: %s (%s)\n", groupID, groupName)
			}

			for _, tag := range instance.Tags {
				key := ""
				val := ""
				if tag.Key != nil {
					key = *tag.Key
				}
				if tag.Value != nil {
					val = *tag.Value
				}
				fmt.Printf("Tag: %s=%s\n", key, val)
			}
		}
	}
}
