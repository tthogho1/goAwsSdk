package utils

import (
	"fmt"
	"strings"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/ec2"
)

// Security group editing commands. Output is line-oriented so Ec2Viewer
// (standalone/parser.go) can parse it; keep the formats in sync:
//
//	SG: sg-xxx (name)
//	Group: sg-xxx (name)
//	Rule: <ingress|egress>, <proto>, <from>, <to>, <cidr>, <description>
//	RuleOther: <ingress|egress>, <proto>, <from>, <to>, <source>
//
// All functions return an error instead of printing it, so the caller can
// exit with a non-zero status.

func strOr(p *string, def string) string {
	if p == nil {
		return def
	}
	return *p
}

func portStr(p *int64) string {
	if p == nil {
		return "-1"
	}
	return fmt.Sprintf("%d", *p)
}

// ListSecurityGroups prints the security groups in a VPC (all groups if vpcID is empty).
func ListSecurityGroups(svc *ec2.EC2, vpcID string) error {
	input := &ec2.DescribeSecurityGroupsInput{}
	if vpcID != "" {
		input.Filters = []*ec2.Filter{{Name: aws.String("vpc-id"), Values: []*string{aws.String(vpcID)}}}
	}
	return svc.DescribeSecurityGroupsPages(input, func(page *ec2.DescribeSecurityGroupsOutput, _ bool) bool {
		for _, g := range page.SecurityGroups {
			fmt.Printf("SG: %s (%s)\n", strOr(g.GroupId, "None"), strOr(g.GroupName, ""))
		}
		return true
	})
}

// DescribeSecurityGroupRules prints the inbound and outbound rules of a security group.
func DescribeSecurityGroupRules(svc *ec2.EC2, groupID string) error {
	if groupID == "" {
		return fmt.Errorf("security group id (-g) is required")
	}
	res, err := svc.DescribeSecurityGroups(&ec2.DescribeSecurityGroupsInput{GroupIds: []*string{aws.String(groupID)}})
	if err != nil {
		return err
	}
	if len(res.SecurityGroups) == 0 {
		return fmt.Errorf("security group %s not found", groupID)
	}
	g := res.SecurityGroups[0]
	fmt.Printf("Group: %s (%s)\n", strOr(g.GroupId, "None"), strOr(g.GroupName, ""))
	printPerms("ingress", g.IpPermissions)
	printPerms("egress", g.IpPermissionsEgress)
	return nil
}

func printPerms(dir string, perms []*ec2.IpPermission) {
	for _, p := range perms {
		proto := strOr(p.IpProtocol, "-1")
		from, to := portStr(p.FromPort), portStr(p.ToPort)
		for _, r := range p.IpRanges {
			fmt.Printf("Rule: %s, %s, %s, %s, %s, %s\n", dir, proto, from, to, strOr(r.CidrIp, ""), strOr(r.Description, ""))
		}
		for _, r := range p.Ipv6Ranges {
			fmt.Printf("RuleOther: %s, %s, %s, %s, %s\n", dir, proto, from, to, strOr(r.CidrIpv6, ""))
		}
		for _, r := range p.UserIdGroupPairs {
			fmt.Printf("RuleOther: %s, %s, %s, %s, %s\n", dir, proto, from, to, strOr(r.GroupId, ""))
		}
		for _, r := range p.PrefixListIds {
			fmt.Printf("RuleOther: %s, %s, %s, %s, %s\n", dir, proto, from, to, strOr(r.PrefixListId, ""))
		}
	}
}

// SGRuleSpec describes a single IPv4 CIDR rule.
type SGRuleSpec struct {
	GroupID     string
	Direction   string // ingress | egress
	Protocol    string // tcp | udp | icmp | -1
	FromPort    int64
	ToPort      int64
	Cidr        string
	Description string
}

func (r SGRuleSpec) validate() error {
	if r.GroupID == "" {
		return fmt.Errorf("security group id (-g) is required")
	}
	if r.Direction != "ingress" && r.Direction != "egress" {
		return fmt.Errorf("-dir must be ingress or egress, got %q", r.Direction)
	}
	if r.Cidr == "" {
		return fmt.Errorf("-cidr is required")
	}
	return nil
}

func (r SGRuleSpec) permission(withDesc bool) []*ec2.IpPermission {
	ipr := &ec2.IpRange{CidrIp: aws.String(r.Cidr)}
	if withDesc && r.Description != "" {
		ipr.Description = aws.String(r.Description)
	}
	p := &ec2.IpPermission{
		IpProtocol: aws.String(r.Protocol),
		IpRanges:   []*ec2.IpRange{ipr},
	}
	if r.Protocol != "-1" {
		p.FromPort = aws.Int64(r.FromPort)
		p.ToPort = aws.Int64(r.ToPort)
	}
	return []*ec2.IpPermission{p}
}

func (r SGRuleSpec) String() string {
	return fmt.Sprintf("%s, %s, %s, %d, %d, %s", r.GroupID, r.Direction, r.Protocol, r.FromPort, r.ToPort, r.Cidr)
}

// AddSecurityGroupRule authorizes an IPv4 CIDR rule.
func AddSecurityGroupRule(svc *ec2.EC2, r SGRuleSpec) error {
	if err := r.validate(); err != nil {
		return err
	}
	var err error
	if r.Direction == "ingress" {
		_, err = svc.AuthorizeSecurityGroupIngress(&ec2.AuthorizeSecurityGroupIngressInput{
			GroupId: aws.String(r.GroupID), IpPermissions: r.permission(true)})
	} else {
		_, err = svc.AuthorizeSecurityGroupEgress(&ec2.AuthorizeSecurityGroupEgressInput{
			GroupId: aws.String(r.GroupID), IpPermissions: r.permission(true)})
	}
	if err != nil {
		return err
	}
	fmt.Printf("Added: %s\n", r)
	return nil
}

// DeleteSecurityGroupRule revokes an IPv4 CIDR rule.
func DeleteSecurityGroupRule(svc *ec2.EC2, r SGRuleSpec) error {
	if err := r.validate(); err != nil {
		return err
	}
	var err error
	if r.Direction == "ingress" {
		_, err = svc.RevokeSecurityGroupIngress(&ec2.RevokeSecurityGroupIngressInput{
			GroupId: aws.String(r.GroupID), IpPermissions: r.permission(false)})
	} else {
		_, err = svc.RevokeSecurityGroupEgress(&ec2.RevokeSecurityGroupEgressInput{
			GroupId: aws.String(r.GroupID), IpPermissions: r.permission(false)})
	}
	if err != nil {
		return err
	}
	fmt.Printf("Deleted: %s\n", r)
	return nil
}

// SetInstanceSecurityGroups replaces the security groups attached to an instance.
// It fails for instances with multiple network interfaces (AWS restriction).
func SetInstanceSecurityGroups(svc *ec2.EC2, instanceID string, groupIDs []string) error {
	if instanceID == "" {
		return fmt.Errorf("instance id (-i) is required")
	}
	var ids []*string
	for _, g := range groupIDs {
		if g = strings.TrimSpace(g); g != "" {
			ids = append(ids, aws.String(g))
		}
	}
	if len(ids) == 0 {
		return fmt.Errorf("at least one security group id (-g) is required")
	}
	if _, err := svc.ModifyInstanceAttribute(&ec2.ModifyInstanceAttributeInput{
		InstanceId: aws.String(instanceID),
		Groups:     ids,
	}); err != nil {
		return err
	}
	fmt.Printf("SecurityGroupsSet: %s %s\n", instanceID, strings.Join(groupIDs, ","))
	return nil
}

// CreateSecurityGroup creates a security group in a VPC and prints its ID.
// AWS adds a default "allow all" egress rule and no ingress rules.
func CreateSecurityGroup(svc *ec2.EC2, vpcID, name, description string) error {
	if vpcID == "" {
		return fmt.Errorf("VPC id (-vpc) is required")
	}
	if name == "" {
		return fmt.Errorf("group name (-n) is required")
	}
	if description == "" {
		description = name
	}
	res, err := svc.CreateSecurityGroup(&ec2.CreateSecurityGroupInput{
		VpcId:       aws.String(vpcID),
		GroupName:   aws.String(name),
		Description: aws.String(description),
	})
	if err != nil {
		return err
	}
	fmt.Printf("Created: %s (%s)\n", strOr(res.GroupId, "None"), name)
	return nil
}
