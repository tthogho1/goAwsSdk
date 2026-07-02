package model

// Instance はEC2インスタンス情報を保持する構造体
type Instance struct {
	ID           string
	Status       string
	InstanceType string
	PrivateIP    string
	PublicIP     string
	Name         string
}

// MapStatus はAWSステータスをon/off/-に変換する
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
