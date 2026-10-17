package model

// Instance is a struct that holds EC2 instance information
type Instance struct {
	ID           string
	Status       string
	InstanceType string
	PrivateIP    string
	PublicIP     string
	Name         string
}

// CostEntry holds a single service's cost amount for a time period
type CostEntry struct {
	Service string
	Amount  string
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
