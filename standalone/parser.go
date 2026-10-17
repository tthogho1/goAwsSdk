package main

import (
	"strconv"
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

// parseCostOutput parses the output of "awsctrl -c cost ..." into per-service
// cost entries and a grand total.
// Expected lines look like:
//
//	Time Period: 2026-10-01 - 2026-10-17
//	Service: Amazon EC2, Amount: 12.340000
//	Total: 123.450000
func parseCostOutput(output string) ([]model.CostEntry, float64) {
	var entries []model.CostEntry
	var total float64

	lines := strings.Split(output, "\n")
	for _, line := range lines {
		trimmed := strings.TrimSpace(line)
		switch {
		case strings.HasPrefix(trimmed, "Service:"):
			rest := strings.TrimPrefix(trimmed, "Service:")
			parts := strings.SplitN(rest, ", Amount:", 2)
			if len(parts) != 2 {
				continue
			}
			service := strings.TrimSpace(parts[0])
			amount := strings.TrimSpace(parts[1])
			entries = append(entries, model.CostEntry{Service: service, Amount: amount})
		case strings.HasPrefix(trimmed, "Total:"):
			amountStr := strings.TrimSpace(strings.TrimPrefix(trimmed, "Total:"))
			if v, err := strconv.ParseFloat(amountStr, 64); err == nil {
				total = v
			}
		}
	}
	return entries, total
}
