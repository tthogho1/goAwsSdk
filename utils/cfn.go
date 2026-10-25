package utils

import (
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/cloudformation"
	"github.com/aws/aws-sdk-go/service/ec2"
)

const (
	cfnPollInterval = 5 * time.Second
	cfnPollTimeout  = 10 * time.Minute
)

// GenerateCfnTemplate builds a CloudFormation template from existing EC2
// instances using the IaC generator (CreateGeneratedTemplate), writes the YAML
// to outPath and deletes the generated template from the account afterwards.
// When related is true, the VPC, Subnet and Security Groups of each instance
// are included as well.
//
// Progress is printed as simple "Key: Value" lines so it can be parsed by
// callers (e.g. Ec2Viewer):
//
//	Resource: AWS::EC2::Instance InstanceId=i-xxxx
//	Status: CREATE_IN_PROGRESS
//	Warning: AWS::EC2::Instance i-xxxx: <description>
//	Template: <outPath>
func GenerateCfnTemplate(ec2Svc *ec2.EC2, cfnSvc *cloudformation.CloudFormation, instanceIDs []string, related bool, outPath string) error {
	ids := nonEmpty(instanceIDs)
	if len(ids) == 0 {
		return fmt.Errorf("instance id is required (-i id1,id2,...)")
	}
	if outPath == "" {
		return fmt.Errorf("output file is required (-o template.yaml)")
	}

	resources, err := cfnResources(ec2Svc, ids, related)
	if err != nil {
		return err
	}
	for _, r := range resources {
		for k, v := range r.ResourceIdentifier {
			fmt.Printf("Resource: %s %s=%s\n", *r.ResourceType, k, *v)
		}
	}

	name := fmt.Sprintf("awsctrl-%s", time.Now().Format("20060102-150405"))
	_, err = cfnSvc.CreateGeneratedTemplate(&cloudformation.CreateGeneratedTemplateInput{
		GeneratedTemplateName: aws.String(name),
		Resources:             resources,
		TemplateConfiguration: &cloudformation.TemplateConfiguration{
			// Retain so that deleting a stack built from this template never
			// deletes the original resources.
			DeletionPolicy:      aws.String(cloudformation.GeneratedTemplateDeletionPolicyRetain),
			UpdateReplacePolicy: aws.String(cloudformation.GeneratedTemplateUpdateReplacePolicyRetain),
		},
	})
	if err != nil {
		return fmt.Errorf("CreateGeneratedTemplate: %v", err)
	}
	// The generated template counts against an account quota; remove it once we have the body.
	defer func() {
		if _, err := cfnSvc.DeleteGeneratedTemplate(&cloudformation.DeleteGeneratedTemplateInput{
			GeneratedTemplateName: aws.String(name),
		}); err != nil {
			fmt.Printf("Warning: failed to delete generated template %s: %v\n", name, err)
		}
	}()

	desc, err := waitGeneratedTemplate(cfnSvc, name)
	if err != nil {
		return err
	}
	for _, r := range desc.Resources {
		for _, w := range r.Warnings {
			for _, p := range w.Properties {
				fmt.Printf("Warning: %s %s: %s %s\n", aws.StringValue(r.ResourceType), identifierString(r.ResourceIdentifier),
					aws.StringValue(p.PropertyPath), aws.StringValue(p.Description))
			}
		}
	}

	out, err := cfnSvc.GetGeneratedTemplate(&cloudformation.GetGeneratedTemplateInput{
		GeneratedTemplateName: aws.String(name),
		Format:                aws.String(cloudformation.TemplateFormatYaml),
	})
	if err != nil {
		return fmt.Errorf("GetGeneratedTemplate: %v", err)
	}
	if err := os.WriteFile(outPath, []byte(aws.StringValue(out.TemplateBody)), 0644); err != nil {
		return fmt.Errorf("failed to write %s: %v", outPath, err)
	}
	fmt.Printf("Template: %s\n", outPath)
	return nil
}

// cfnResources resolves the instances (and optionally their VPC, Subnet and
// Security Groups) into IaC generator resource definitions, without duplicates.
func cfnResources(svc *ec2.EC2, instanceIDs []string, related bool) ([]*cloudformation.ResourceDefinition, error) {
	result, err := svc.DescribeInstances(&ec2.DescribeInstancesInput{
		InstanceIds: aws.StringSlice(instanceIDs),
	})
	if err != nil {
		return nil, fmt.Errorf("DescribeInstances: %v", err)
	}

	var resources []*cloudformation.ResourceDefinition
	seen := map[string]bool{}
	add := func(resourceType, key string, id *string) {
		if id == nil || *id == "" || seen[resourceType+*id] {
			return
		}
		seen[resourceType+*id] = true
		resources = append(resources, &cloudformation.ResourceDefinition{
			ResourceType:       aws.String(resourceType),
			ResourceIdentifier: map[string]*string{key: aws.String(*id)},
		})
	}

	// Identifier keys are the primaryIdentifier of each type in the CloudFormation registry.
	for _, reservation := range result.Reservations {
		for _, instance := range reservation.Instances {
			add("AWS::EC2::Instance", "InstanceId", instance.InstanceId)
			if !related {
				continue
			}
			add("AWS::EC2::VPC", "VpcId", instance.VpcId)
			add("AWS::EC2::Subnet", "SubnetId", instance.SubnetId)
			for _, sg := range instance.SecurityGroups {
				add("AWS::EC2::SecurityGroup", "Id", sg.GroupId)
			}
		}
	}
	if len(resources) == 0 {
		return nil, fmt.Errorf("no instances found for %s", strings.Join(instanceIDs, ","))
	}
	return resources, nil
}

// waitGeneratedTemplate polls DescribeGeneratedTemplate until generation completes or fails.
func waitGeneratedTemplate(svc *cloudformation.CloudFormation, name string) (*cloudformation.DescribeGeneratedTemplateOutput, error) {
	deadline := time.Now().Add(cfnPollTimeout)
	for {
		desc, err := svc.DescribeGeneratedTemplate(&cloudformation.DescribeGeneratedTemplateInput{
			GeneratedTemplateName: aws.String(name),
		})
		if err != nil {
			return nil, fmt.Errorf("DescribeGeneratedTemplate: %v", err)
		}
		status := aws.StringValue(desc.Status)
		fmt.Printf("Status: %s\n", status)
		switch status {
		case cloudformation.GeneratedTemplateStatusComplete:
			return desc, nil
		case cloudformation.GeneratedTemplateStatusFailed:
			return nil, fmt.Errorf("template generation failed: %s", aws.StringValue(desc.StatusReason))
		}
		if time.Now().After(deadline) {
			return nil, fmt.Errorf("timed out after %s waiting for template generation (last status %s)", cfnPollTimeout, status)
		}
		time.Sleep(cfnPollInterval)
	}
}

func identifierString(m map[string]*string) string {
	var parts []string
	for k, v := range m {
		parts = append(parts, k+"="+aws.StringValue(v))
	}
	return strings.Join(parts, ",")
}

func nonEmpty(values []string) []string {
	var out []string
	for _, v := range values {
		if v = strings.TrimSpace(v); v != "" {
			out = append(out, v)
		}
	}
	return out
}
