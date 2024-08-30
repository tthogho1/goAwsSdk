package types

import "encoding/json"

type NetWorkIF struct {
	Data map[string]interface{}
}

// String メソッドの実装
func (f *NetWorkIF) String() string {
	bytes, _ := json.Marshal(f.Data)
	return string(bytes)
}

// Set メソッドの実装
func (f *NetWorkIF) Set(value string) error {
	return json.Unmarshal([]byte(value), &f.Data)
}

type NetWorkIfKey struct {
	Device_index      string
	Public_ip_address string
	Subnet_id         string
	Groups            string
}

func NewNetWorkIfKey() *NetWorkIfKey {
	return &NetWorkIfKey{
		Device_index:      "device_index",
		Public_ip_address: "public_ip_address",
		Subnet_id:         "subnet_id",
		Groups:            "groups",
	}
}
