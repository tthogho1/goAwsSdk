package main

import (
	"strings"

	"ec2viewer/model"
)

// parseOutput parses the awsctrl output text and converts it into instance information
func parseOutput(output string) []model.Instance {
	var result []model.Instance
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

// parseInstanceLine parses a single line of instance information
// Input format: "ID: <id>,  <status>, <type>, <privateIP>, <publicIP>  <TagKey>: <TagVal>  ..."
func parseInstanceLine(line string) model.Instance {
	line = strings.TrimPrefix(line, "ID: ")

	parts := strings.SplitN(line, ", ", 5)
	if len(parts) < 5 {
		return model.Instance{}
	}

	id := strings.TrimSpace(parts[0])
	status := strings.TrimSpace(parts[1])
	instanceType := strings.TrimSpace(parts[2])
	privateIP := strings.TrimSpace(parts[3])
	lastPart := parts[4]

	publicIP := lastPart
	name := ""

	// Separate the tag portion after publicIP (tags are delimited by "  ")
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

	return model.Instance{
		ID:           id,
		Status:       status,
		InstanceType: instanceType,
		PrivateIP:    privateIP,
		PublicIP:     publicIP,
		Name:         name,
	}
}
