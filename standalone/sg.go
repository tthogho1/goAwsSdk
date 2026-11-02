package main

import (
	"fmt"
	"net"
	"sort"
	"strconv"
	"strings"

	"gioui.org/widget"

	"ec2viewer/model"
	"ec2viewer/ui"
)

// Security group editing handlers. They run awsctrl synchronously, like
// handleDetail / handleCost, and only touch state from the event loop.

// handleSGOpen opens the SG panel for the instance shown in the Detail panel:
// it lists the SGs of the instance's VPC (pre-checking the attached ones) and
// loads the rules of the selected attached SG.
func handleSGOpen(state *ui.AppState) {
	profile := strings.TrimSpace(state.SelectedProfile)
	d := state.Detail
	if profile == "" || d.InstanceId == "" {
		return
	}

	if state.SGInstanceID != d.InstanceId {
		state.SGSelected = ""
		state.SGFromEd.SetText("")
		state.SGToEd.SetText("")
		state.SGCidrEd.SetText("")
		state.SGDescEd.SetText("")
	}
	state.SGInstanceID = d.InstanceId
	state.SGVpcID = d.VpcId
	state.SGAttached = state.SGAttached[:0]
	for _, sg := range d.SecurityGroups {
		state.SGAttached = append(state.SGAttached, model.SGIDFromDetail(sg))
	}
	state.SGSelectBtns = make([]widget.Clickable, len(state.SGAttached))
	if state.SGNewDir == "" {
		state.SGNewDir = "ingress"
	}
	if state.SGNewProto == "" {
		state.SGNewProto = "tcp"
	}
	state.SGErr = ""
	state.SGPanelOpen = true

	state.SGAll = nil
	if out, err := executeAwsCtrlSGList(profile, d.VpcId); err != nil {
		state.SGErr = fmt.Sprintf("security group list error: %v", err)
	} else {
		state.SGAll = parseSGList(out)
	}
	state.SGAttachBoxes = make([]widget.Bool, len(state.SGAll))
	for i, sg := range state.SGAll {
		state.SGAttachBoxes[i].Value = contains(state.SGAttached, sg.ID)
	}

	sel := state.SGSelected
	if !contains(state.SGAttached, sel) {
		sel = ""
		if len(state.SGAttached) > 0 {
			sel = state.SGAttached[0]
		}
	}
	if sel == "" {
		state.SGSelected = ""
		state.SGRules = nil
		state.SGRuleDelBtns = nil
		state.SGDelPending = -1
		return
	}
	loadSGRules(state, profile, sel)
}

// handleSGSelect shows the rules of another attached security group.
func handleSGSelect(state *ui.AppState, groupID string) {
	profile := strings.TrimSpace(state.SelectedProfile)
	if profile == "" || groupID == "" {
		return
	}
	state.SGErr = ""
	loadSGRules(state, profile, groupID)
}

func loadSGRules(state *ui.AppState, profile, groupID string) {
	state.SGSelected = groupID
	state.SGDelPending = -1
	state.SGRules = nil
	out, err := executeAwsCtrlSGRules(profile, groupID)
	if err != nil {
		state.SGErr = fmt.Sprintf("rule fetch error: %v", err)
	} else {
		_, state.SGRules = parseSGRules(out)
	}
	state.SGRuleDelBtns = make([]widget.Clickable, len(state.SGRules))
}

// handleSGApply attaches exactly the checked security groups to the instance.
func handleSGApply(state *ui.AppState) {
	profile := strings.TrimSpace(state.SelectedProfile)
	if profile == "" || state.SGInstanceID == "" {
		return
	}
	var ids []string
	for i, sg := range state.SGAll {
		if i < len(state.SGAttachBoxes) && state.SGAttachBoxes[i].Value {
			ids = append(ids, sg.ID)
		}
	}
	if len(ids) == 0 {
		state.SGErr = "Check at least one security group (an instance must have one)"
		return
	}
	if sameSet(ids, state.SGAttached) {
		state.SGErr = ""
		state.InfoMsg = "Attached security groups unchanged"
		return
	}
	if _, err := executeAwsCtrlSetSG(profile, state.SGInstanceID, ids); err != nil {
		state.SGErr = fmt.Sprintf("attach error: %v", err)
		return
	}

	// Refresh the Detail panel, then rebuild the SG panel from it
	handleDetail(state, state.SGInstanceID)
	if state.DetailErr != "" {
		state.SGErr = "Security groups changed, but refreshing details failed: " + state.DetailErr
		return
	}
	handleSGOpen(state)
	state.ErrMsg = ""
	state.InfoMsg = fmt.Sprintf("Security groups of %s set to %s", state.SGInstanceID, strings.Join(ids, ", "))
}

// handleSGCreate creates a security group in the instance's VPC and, when
// "attach" is checked, adds it to the instance's attached groups. The new
// group is then selected so rules can be added to it right away.
func handleSGCreate(state *ui.AppState) {
	profile := strings.TrimSpace(state.SelectedProfile)
	if profile == "" || state.SGInstanceID == "" {
		return
	}
	name := strings.TrimSpace(state.SGNewNameEd.Text())
	desc := strings.TrimSpace(state.SGNewGroupDesc.Text())
	if name == "" {
		state.SGErr = "Enter a name for the new security group"
		return
	}
	if strings.HasPrefix(strings.ToLower(name), "sg-") {
		state.SGErr = "Security group names can't start with \"sg-\""
		return
	}
	if state.SGVpcID == "" || state.SGVpcID == "None" {
		state.SGErr = "This instance has no VPC; can't create a security group for it"
		return
	}

	out, err := executeAwsCtrlSGCreate(profile, state.SGVpcID, name, desc)
	if err != nil {
		state.SGErr = fmt.Sprintf("create error: %v", err)
		return
	}
	newID := parseSGCreated(out)
	if newID == "" {
		state.SGErr = "Security group was created, but its ID was not found in the awsctrl output:\n" + out
		handleSGOpen(state)
		return
	}
	state.SGNewNameEd.SetText("")
	state.SGNewGroupDesc.SetText("")
	state.SGErr = ""

	if !state.SGCreateAttach.Value {
		handleSGOpen(state)
		state.ErrMsg = ""
		state.InfoMsg = fmt.Sprintf("Created %s (%s); not attached", newID, name)
		return
	}

	ids := append(append([]string(nil), state.SGAttached...), newID)
	if _, err := executeAwsCtrlSetSG(profile, state.SGInstanceID, ids); err != nil {
		handleSGOpen(state)
		state.SGErr = fmt.Sprintf("Created %s, but attaching it to %s failed: %v", newID, state.SGInstanceID, err)
		return
	}
	handleDetail(state, state.SGInstanceID)
	if state.DetailErr != "" {
		state.SGErr = fmt.Sprintf("Created and attached %s, but refreshing details failed: %s", newID, state.DetailErr)
		return
	}
	state.SGSelected = newID // show the new group's rules
	handleSGOpen(state)
	state.ErrMsg = ""
	state.InfoMsg = fmt.Sprintf("Created %s (%s) and attached it to %s - add inbound rules below", newID, name, state.SGInstanceID)
}

// handleSGToggleDir switches the new-rule direction between ingress and egress.
func handleSGToggleDir(state *ui.AppState) {
	if state.SGNewDir == "egress" {
		state.SGNewDir = "ingress"
	} else {
		state.SGNewDir = "egress"
	}
}

var sgProtocols = []string{"tcp", "udp", "icmp", "all"}

// handleSGCycleProto cycles the new-rule protocol tcp -> udp -> icmp -> all.
func handleSGCycleProto(state *ui.AppState) {
	for i, p := range sgProtocols {
		if p == state.SGNewProto {
			state.SGNewProto = sgProtocols[(i+1)%len(sgProtocols)]
			return
		}
	}
	state.SGNewProto = sgProtocols[0]
}

// handleSGAdd validates the add-rule form and authorizes the rule.
func handleSGAdd(state *ui.AppState) {
	profile := strings.TrimSpace(state.SelectedProfile)
	if profile == "" {
		return
	}
	if state.SGSelected == "" {
		state.SGErr = "Select a security group first"
		return
	}
	rule, err := buildSGRule(state.SGNewDir, state.SGNewProto,
		state.SGFromEd.Text(), state.SGToEd.Text(), state.SGCidrEd.Text(), state.SGDescEd.Text())
	if err != nil {
		state.SGErr = err.Error()
		return
	}
	if _, err := executeAwsCtrlSGRule(profile, "sgadd", state.SGSelected, rule); err != nil {
		state.SGErr = fmt.Sprintf("add rule error: %v", err)
		return
	}
	state.SGFromEd.SetText("")
	state.SGToEd.SetText("")
	state.SGCidrEd.SetText("")
	state.SGDescEd.SetText("")
	state.SGErr = ""
	loadSGRules(state, profile, state.SGSelected)
	state.ErrMsg = ""
	state.InfoMsg = fmt.Sprintf("Added %s rule %s %s to %s", rule.Direction, rule.Protocol, rule.Cidr, state.SGSelected)
}

// handleSGDelete deletes rule idx of the selected group. The first click only
// arms the button ("Confirm?"); the second click on the same rule deletes it.
func handleSGDelete(state *ui.AppState, idx int) {
	profile := strings.TrimSpace(state.SelectedProfile)
	if profile == "" || idx < 0 || idx >= len(state.SGRules) || !state.SGRules[idx].Editable {
		return
	}
	if state.SGDelPending != idx {
		state.SGDelPending = idx
		return
	}
	rule := state.SGRules[idx]
	if _, err := executeAwsCtrlSGRule(profile, "sgdel", state.SGSelected, rule); err != nil {
		state.SGDelPending = -1
		state.SGErr = fmt.Sprintf("delete rule error: %v", err)
		return
	}
	state.SGErr = ""
	loadSGRules(state, profile, state.SGSelected)
	state.ErrMsg = ""
	state.InfoMsg = fmt.Sprintf("Deleted %s rule %s %s from %s", rule.Direction, rule.Protocol, rule.Cidr, state.SGSelected)
}

// buildSGRule validates the add-rule form input and returns the rule to send.
func buildSGRule(dir, proto, fromStr, toStr, cidr, desc string) (model.SGRule, error) {
	fromStr, toStr = strings.TrimSpace(fromStr), strings.TrimSpace(toStr)
	cidr, desc = strings.TrimSpace(cidr), strings.TrimSpace(desc)
	r := model.SGRule{Direction: dir, Protocol: proto, Cidr: cidr, Description: desc, Editable: true}

	switch proto {
	case "all":
		r.Protocol, r.FromPort, r.ToPort = "-1", "-1", "-1"
	case "icmp":
		// From = ICMP type, To = ICMP code; -1 (or empty) means all
		from, err := parseIntRange(fromStr, "ICMP type (From)", -1, 255, -1)
		if err != nil {
			return r, err
		}
		to, err := parseIntRange(toStr, "ICMP code (To)", -1, 255, -1)
		if err != nil {
			return r, err
		}
		r.FromPort, r.ToPort = strconv.Itoa(from), strconv.Itoa(to)
	case "tcp", "udp":
		if fromStr == "" {
			return r, fmt.Errorf("From port is required for %s", proto)
		}
		from, err := parseIntRange(fromStr, "From port", 0, 65535, 0)
		if err != nil {
			return r, err
		}
		to, err := parseIntRange(toStr, "To port", 0, 65535, from)
		if err != nil {
			return r, err
		}
		if from > to {
			return r, fmt.Errorf("From port (%d) must not be greater than To port (%d)", from, to)
		}
		r.FromPort, r.ToPort = strconv.Itoa(from), strconv.Itoa(to)
	default:
		return r, fmt.Errorf("unsupported protocol %q", proto)
	}

	if cidr == "" {
		return r, fmt.Errorf("CIDR is required (e.g. 203.0.113.5/32)")
	}
	ip, ipnet, err := net.ParseCIDR(cidr)
	if err != nil || ip.To4() == nil {
		return r, fmt.Errorf("invalid IPv4 CIDR %q (e.g. 203.0.113.5/32)", cidr)
	}
	if ipnet.String() != cidr {
		return r, fmt.Errorf("CIDR %q has host bits set; did you mean %s?", cidr, ipnet.String())
	}
	return r, nil
}

// parseIntRange parses s as an int in [min, max]; empty s yields def.
func parseIntRange(s, name string, min, max, def int) (int, error) {
	if s == "" {
		return def, nil
	}
	v, err := strconv.Atoi(s)
	if err != nil || v < min || v > max {
		return 0, fmt.Errorf("%s must be a number from %d to %d", name, min, max)
	}
	return v, nil
}

func contains(list []string, s string) bool {
	for _, v := range list {
		if v == s {
			return true
		}
	}
	return false
}

func sameSet(a, b []string) bool {
	if len(a) != len(b) {
		return false
	}
	a2 := append([]string(nil), a...)
	b2 := append([]string(nil), b...)
	sort.Strings(a2)
	sort.Strings(b2)
	for i := range a2 {
		if a2[i] != b2[i] {
			return false
		}
	}
	return true
}
