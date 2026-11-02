package model

import "strings"

// Instance is a struct that holds EC2 instance information
type Instance struct {
	ID           string
	Status       string
	InstanceType string
	PrivateIP    string
	PublicIP     string
	Name         string
}

// PrivateIPList splits the raw ';'-joined PrivateIP field into individual
// IP addresses (multiple ENIs / secondary IPs are joined with ';').
func (i Instance) PrivateIPList() []string {
	if i.PrivateIP == "" || i.PrivateIP == "None" {
		return nil
	}
	return strings.Split(i.PrivateIP, ";")
}

// DisplayPrivateIP returns a human-readable, comma-separated representation
// of all private IP addresses for display in the UI.
func (i Instance) DisplayPrivateIP() string {
	ips := i.PrivateIPList()
	if len(ips) == 0 {
		return i.PrivateIP
	}
	return strings.Join(ips, ", ")
}

// CostEntry holds a single service's cost amount for a time period
type CostEntry struct {
	Service string
	Amount  string
}

// Tag holds a single EC2 tag key/value pair
type Tag struct {
	Key   string
	Value string
}

// InstanceDetail holds detailed EC2 instance information:
// VPC, Subnet, attached Security Groups and all Tags.
type InstanceDetail struct {
	InstanceId     string
	VpcId          string
	SubnetId       string
	SecurityGroups []string
	Tags           []Tag
}

// MapStatus converts an AWS status into on/off/-
func MapStatus(awsStatus string) string {
	switch awsStatus {
	case "running":
		return "on"
	case "stopped":
		return "off"
	default:
		return "-"
	}
}

// CfnResult holds the outcome of generating a CloudFormation template
// from selected instances via "awsctrl -c cfn".
type CfnResult struct {
	Template  string   // YAML template body
	Resources []string // "<type> <key>=<id>" lines included in the template
	Warnings  []string // property warnings reported by the IaC generator
	Err       string   // non-empty when generation failed
}

// SGSummary identifies a security group (from "awsctrl -c sglist").
type SGSummary struct {
	ID   string
	Name string
}

// Label returns "sg-xxx (name)", the same format as InstanceDetail.SecurityGroups.
func (s SGSummary) Label() string {
	if s.Name == "" {
		return s.ID
	}
	return s.ID + " (" + s.Name + ")"
}

// SGRule is one security group rule (from "awsctrl -c sgrules").
// Only IPv4 CIDR rules are Editable; for other rules (IPv6, SG reference,
// prefix list) Cidr holds the source and Description is empty.
type SGRule struct {
	Direction   string // ingress | egress
	Protocol    string // tcp | udp | icmp | -1
	FromPort    string
	ToPort      string
	Cidr        string
	Description string
	Editable    bool
}

// SGIDFromDetail extracts the group ID from an "sg-xxx (name)" string.
func SGIDFromDetail(s string) string {
	s = strings.TrimSpace(s)
	if i := strings.Index(s, " "); i >= 0 {
		return s[:i]
	}
	return s
}
