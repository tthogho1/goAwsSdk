package main

import (
	"strings"

	"gioui.org/widget"
)

// Instance はEC2インスタンス情報を保持する構造体
type Instance struct {
	ID           string
	Status       string
	InstanceType string
	PrivateIP    string
	PublicIP     string
	Name         string
}

// mapStatus はAWSステータスをon/off/-に変換する
func mapStatus(awsStatus string) string {
	switch awsStatus {
	case "running":
		return "on"
	case "stopped":
		return "off"
	default:
		return "-"
	}
}

// initStatusSlices はインスタンス取得後にステータス管理用スライスを初期化する
func initStatusSlices() {
	n := len(instances)
	originalStatus = make([]string, n)
	desiredStatus = make([]string, n)
	toggleBtns = make([]widget.Clickable, n)
	// create per-cell clickables: one per instance * number of columns
	cellClickables = make([]widget.Clickable, n*len(headers))
	for i, inst := range instances {
		s := mapStatus(inst.Status)
		originalStatus[i] = s
		desiredStatus[i] = s
	}
	// mark visible indices dirty so UI recomputes filter after a fetch
	visibleDirty = true
}

// hasStatusChanges はon/offが変更されたインスタンスがあるかを返す
func hasStatusChanges() bool {
	for i := range originalStatus {
		if originalStatus[i] != "-" && desiredStatus[i] != originalStatus[i] {
			return true
		}
	}
	return false
}

// parseOutput は awsctrl の出力テキストをパースしてインスタンス情報に変換する
func parseOutput(output string) []Instance {
	var result []Instance
	lines := strings.Split(output, "\n")
	for _, line := range lines {
		trimmed := strings.TrimSpace(line)
		if !strings.HasPrefix(trimmed, "ID:") {
			continue
		}
		inst := parseInstanceLine(trimmed)
		if inst.ID != "" {
			result = append(result, inst)
		}
	}
	return result
}

// parseInstanceLine は1行のインスタンス情報をパースする
// 入力形式: "ID: <id>,  <status>, <type>, <privateIP>, <publicIP>  <TagKey>: <TagVal>  ..."
func parseInstanceLine(line string) Instance {
	line = strings.TrimPrefix(line, "ID: ")

	parts := strings.SplitN(line, ", ", 5)
	if len(parts) < 5 {
		return Instance{}
	}

	id := strings.TrimSpace(parts[0])
	status := strings.TrimSpace(parts[1])
	instanceType := strings.TrimSpace(parts[2])
	privateIP := strings.TrimSpace(parts[3])
	lastPart := parts[4]

	publicIP := lastPart
	name := ""

	// publicIP 以降のタグ部分を分離（タグは "  " で区切られている）
	tagIdx := strings.Index(lastPart, "  ")
	if tagIdx >= 0 {
		publicIP = strings.TrimSpace(lastPart[:tagIdx])
		tagStr := lastPart[tagIdx:]
		tagEntries := strings.Split(tagStr, "  ")
		for _, entry := range tagEntries {
			entry = strings.TrimSpace(entry)
			if entry == "" {
				continue
			}
			if kv := strings.SplitN(entry, ": ", 2); len(kv) == 2 {
				if strings.TrimSpace(kv[0]) == "Name" {
					name = strings.TrimSpace(kv[1])
				}
			}
		}
	} else {
		publicIP = strings.TrimSpace(publicIP)
	}

	return Instance{
		ID:           id,
		Status:       status,
		InstanceType: instanceType,
		PrivateIP:    privateIP,
		PublicIP:     publicIP,
		Name:         name,
	}
}
