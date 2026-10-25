package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/atotto/clipboard"

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

// startCfn handles the "Generate CFn" button. Template generation can take
// minutes, so awsctrl runs in a goroutine; the result is delivered on done and
// applied to state by the event loop via applyCfnResult (the goroutine never
// touches state). invalidate wakes the window so the result is picked up.
func startCfn(state *ui.AppState, done chan<- model.CfnResult, invalidate func()) {
	if state.CfnRunning {
		return
	}
	profile := strings.TrimSpace(state.SelectedProfile)
	if profile == "" {
		state.ErrMsg = "Please select a profile"
		state.InfoMsg = ""
		return
	}
	ids := state.SelectedInstanceIDs()
	if len(ids) == 0 {
		state.ErrMsg = "Select one or more instances (Select column) to generate a template"
		state.InfoMsg = ""
		return
	}
	related := state.CfnRelated.Value

	state.CfnRunning = true
	state.ErrMsg = ""
	state.InfoMsg = fmt.Sprintf("Generating CloudFormation template for %d instance(s)... this can take a few minutes", len(ids))

	go func() {
		done <- generateCfn(profile, ids, related)
		invalidate()
	}()
}

// generateCfn runs awsctrl into a temp file and returns the template body.
func generateCfn(profile string, ids []string, related bool) model.CfnResult {
	tmp, err := os.CreateTemp("", "ec2viewer-cfn-*.yaml")
	if err != nil {
		return model.CfnResult{Err: fmt.Sprintf("cannot create temp file: %v", err)}
	}
	tmpPath := tmp.Name()
	_ = tmp.Close()
	defer os.Remove(tmpPath)

	output, err := executeAwsCtrlCfn(profile, ids, related, tmpPath)
	resources, warnings := parseCfnOutput(output)
	if err != nil {
		logErrorf("cfn generation failed: %v", err)
		return model.CfnResult{Resources: resources, Warnings: warnings, Err: fmt.Sprintf("cfn generation error: %v", err)}
	}
	body, err := os.ReadFile(tmpPath)
	if err != nil {
		return model.CfnResult{Resources: resources, Warnings: warnings, Err: fmt.Sprintf("cannot read generated template: %v", err)}
	}
	return model.CfnResult{Template: string(body), Resources: resources, Warnings: warnings}
}

// applyCfnResult stores a finished generation result and opens the CFn panel.
func applyCfnResult(state *ui.AppState, r model.CfnResult) {
	state.CfnRunning = false
	state.Cfn = r
	state.CfnEditor.SetText(r.Template)
	state.CfnPanelOpen = true
	if r.Err != "" {
		state.ErrMsg = "CloudFormation template generation failed"
		state.InfoMsg = ""
		return
	}
	state.ErrMsg = ""
	state.InfoMsg = fmt.Sprintf("Template generated (%d resources, %d warnings) - review before deploying", len(r.Resources), len(r.Warnings))
}

// handleCfnSave writes the generated template to the user's Downloads folder
// (or the home directory if there is none).
func handleCfnSave(state *ui.AppState) {
	if state.Cfn.Template == "" {
		return
	}
	dir, err := os.UserHomeDir()
	if err != nil {
		state.ErrMsg = fmt.Sprintf("cannot find home directory: %v", err)
		state.InfoMsg = ""
		return
	}
	if dl := filepath.Join(dir, "Downloads"); isDir(dl) {
		dir = dl
	}
	path := filepath.Join(dir, "cfn-template-"+time.Now().Format("20060102-150405")+".yaml")
	if err := os.WriteFile(path, []byte(state.Cfn.Template), 0644); err != nil {
		state.ErrMsg = fmt.Sprintf("failed to save template: %v", err)
		state.InfoMsg = ""
		return
	}
	state.ErrMsg = ""
	state.InfoMsg = "Saved: " + path
}

// handleCfnCopy copies the generated template to the clipboard.
func handleCfnCopy(state *ui.AppState) {
	if state.Cfn.Template == "" {
		return
	}
	if err := clipboard.WriteAll(state.Cfn.Template); err != nil {
		state.ErrMsg = "Failed to copy to clipboard: " + err.Error()
		state.InfoMsg = ""
		return
	}
	state.ErrMsg = ""
	state.InfoMsg = "Copied template to clipboard"
}

func isDir(path string) bool {
	fi, err := os.Stat(path)
	return err == nil && fi.IsDir()
}
