package main

import (
	"fmt"
	"strings"

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
