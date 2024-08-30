package utils

import (
	"awsctrl/types"
	"fmt"
	"log"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/ec2"
)

func CreateInstance(svc *ec2.EC2, amiString string, typeString string, keyPair string, network types.NetWorkIF) {
	networkIfKey := types.NewNetWorkIfKey()

	groups := network.Data[networkIfKey.Groups].([]interface{})
	secGroups := []*string{}
	for i := 0; i < len(groups); i++ {
		secGroups = append(secGroups, aws.String(groups[i].(string)))
	}

	result, err := svc.RunInstances(&ec2.RunInstancesInput{
		ImageId:      aws.String(amiString),
		InstanceType: aws.String(typeString),
		MinCount:     aws.Int64(1),
		MaxCount:     aws.Int64(1),
		KeyName:      aws.String(keyPair),
		//SecurityGroupIds: []*string{aws.String("sg-05c48f49dbf81efeb")},
		NetworkInterfaces: []*ec2.InstanceNetworkInterfaceSpecification{
			{
				DeviceIndex:              aws.Int64(int64(network.Data[networkIfKey.Device_index].(float64))), // required
				AssociatePublicIpAddress: aws.Bool(network.Data[networkIfKey.Public_ip_address].(bool)),
				SubnetId:                 aws.String(network.Data[networkIfKey.Subnet_id].(string)),
				Groups:                   secGroups,
			},
		},
	})

	if err != nil {
		log.Fatalf("Unable to create instance, %v", err)
	}
	fmt.Println(result)
}
