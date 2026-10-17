package main

import (
	"fmt"
	"strings"
	"time"

	"ec2viewer/model"
	"ec2viewer/ui"
)

// refreshInstances runs awsctrl to fetch and update the instance list
func refreshInstances(state *ui.AppState, profile string) error {
	output, err := executeAwsCtrl(profile)
	if err != nil {
		return err
	}
	state.Instances = parseOutput(output)
	state.InitStatusSlices()
	return nil
}

// handleFetch handles the processing when the "Fetch" button is pressed
func handleFetch(state *ui.AppState) {
	profile := strings.TrimSpace(state.SelectedProfile)
	if profile == "" {
		state.ErrMsg = "Please select a profile"
		state.InfoMsg = ""
		return
	}
	state.ErrMsg = ""
	state.InfoMsg = ""
	if err := refreshInstances(state, profile); err != nil {
		state.ErrMsg = fmt.Sprintf("awsctrl execution error: %v", err)
		return
	}
	if len(state.Instances) == 0 {
		state.InfoMsg = "No instances found"
	}
}

// handleExecute handles the processing when the "Execute" button is pressed
func handleExecute(state *ui.AppState) {
	profile := strings.TrimSpace(state.SelectedProfile)
	var errs []string
	for i := range state.Instances {
		if state.DesiredStatus[i] == state.OriginalStatus[i] || state.OriginalStatus[i] == "-" {
			continue
		}
		action := "up"
		if state.DesiredStatus[i] == "off" {
			action = "down"
		}
		if err := executeAwsCtrlAction(profile, action, state.Instances[i].ID); err != nil {
			errs = append(errs, fmt.Sprintf("%s: %v", state.Instances[i].ID, err))
		}
	}
	if len(errs) > 0 {
		state.ErrMsg = "Execution error: " + strings.Join(errs, "; ")
		state.InfoMsg = ""
		return
	}
	// On success: refetch and update status
	if err := refreshInstances(state, profile); err != nil {
		state.ErrMsg = fmt.Sprintf("Refetch error: %v", err)
		return
	}
	state.InfoMsg = "Execution complete"
	state.ErrMsg = ""
}

// handleCost handles the processing when the "Cost" button is pressed.
// It fetches cost-and-usage data for the current month (1st day through today).
func handleCost(state *ui.AppState) {
	profile := strings.TrimSpace(state.SelectedProfile)
	if profile == "" {
		state.ErrMsg = "Please select a profile"
		state.InfoMsg = ""
		return
	}

	now := time.Now()
	start := time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, now.Location())
	// Cost Explorer's End date is exclusive, so use tomorrow to include today
	end := now.AddDate(0, 0, 1)

	startStr := start.Format("2006-01-02")
	endStr := end.Format("2006-01-02")

	output, err := executeAwsCtrlCost(profile, startStr, endStr)
	if err != nil {
		state.ErrMsg = fmt.Sprintf("cost execution error: %v", err)
		state.InfoMsg = ""
		return
	}

	entries, total := parseCostOutput(output)
	state.CostEntries = entries
	state.CostTotal = total
	state.CostPeriod = fmt.Sprintf("%s - %s", startStr, endStr)
	state.CostPanelOpen = true
	state.ErrMsg = ""
	if len(entries) == 0 {
		state.InfoMsg = "No cost data found for this period"
	} else {
		state.InfoMsg = ""
	}
}

// handleDetail handles fetching and displaying details (VPC, Subnet,
// Security Groups, Tags) for a single EC2 instance.
func handleDetail(state *ui.AppState, instanceID string) {
	profile := strings.TrimSpace(state.SelectedProfile)
	if profile == "" {
		state.ErrMsg = "Please select a profile"
		state.InfoMsg = ""
		return
	}
	if instanceID == "" {
		return
	}

	output, err := executeAwsCtrlDetail(profile, instanceID)
	if err != nil {
		state.DetailErr = fmt.Sprintf("detail fetch error: %v", err)
		state.Detail = model.InstanceDetail{}
		state.DetailPanelOpen = true
		return
	}

	state.Detail = parseInstanceDetail(output)
	state.DetailErr = ""
	state.DetailPanelOpen = true
}
