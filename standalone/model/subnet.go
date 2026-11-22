package model

// Subnet holds one VPC subnet as printed by "awsctrl -c describe -t Subnet"
type Subnet struct {
	ID           string
	VpcID        string
	Cidr         string
	AZ           string
	State        string // available | pending
	AvailableIPs string
	Default      bool // default subnet for its AZ
	PublicIP     bool // auto-assigns public IPv4 addresses on launch
	Name         string
}
