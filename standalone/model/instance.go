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
