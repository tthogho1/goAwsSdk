package ui

import (
	"ec2viewer/model"

	"gioui.org/layout"
	"gioui.org/unit"
	"gioui.org/widget"
	"go.uber.org/zap"
)

// Headers are the table column names
var Headers = []string{"ID", "Status", "Type", "PrivateIP", "PublicIP", "Name", "on/off", "Select"}

// ColWidths are the widths of each table column
var ColWidths = []unit.Dp{180, 80, 100, 220, 130, 160, 60, 60}

// SelectCol is the index of the "Select" checkbox column used for CFn generation
const SelectCol = 7

// Top-level views, switched with the tab buttons in the top bar
const (
	TabEC2 = iota
	TabBatch
)

// AppState is a struct that centrally manages the UI and data state
type AppState struct {
	// View tabs
	ActiveTab   int // TabEC2 | TabBatch
	TabEC2Btn   widget.Clickable
	TabBatchBtn widget.Clickable

	// Instance data
	Instances      []model.Instance
	OriginalStatus []string
	DesiredStatus  []string
	ToggleBtns     []widget.Clickable
	CellClickables []widget.Clickable
	SelectBoxes    []widget.Bool // "Select" column, used for CFn generation

	// Filter / display
	VisibleIndices []int
	VisibleDirty   bool

	// UI widgets
	SearchEditor widget.Editor
	SearchQuery  string
	FetchBtn     widget.Clickable
	ExecuteBtn   widget.Clickable
	TableList    widget.List

	// Profile selection menu
	Profiles         []string
	SelectedProfile  string
	ProfileMenuOpen  bool
	ProfileMenuBtn   widget.Clickable
	ProfileMenuItems []widget.Clickable

	// Messages
	ErrMsg  string
	InfoMsg string

	// Cost panel
	CostBtn       widget.Clickable
	CostPanelOpen bool
	CostCloseBtn  widget.Clickable
	CostEntries   []model.CostEntry
	CostTotal     float64
	CostPeriod    string

	// Instance detail panel (opened by clicking the ID cell of a row)
	DetailPanelOpen bool
	DetailCloseBtn  widget.Clickable
	Detail          model.InstanceDetail
	DetailErr       string
	// OnRowDetail is invoked with the instance ID when the ID cell of a row
	// is clicked, so the caller (main package) can fetch instance details.
	OnRowDetail func(instanceID string)

	// CloudFormation template generation (from instances checked in the Select column)
	CfnBtn       widget.Clickable
	CfnRelated   widget.Bool // also include VPC / Subnet / Security Groups
	CfnRunning   bool
	CfnPanelOpen bool
	CfnCloseBtn  widget.Clickable
	CfnSaveBtn   widget.Clickable
	CfnCopyBtn   widget.Clickable
	CfnEditor    widget.Editor // read-only template preview
	Cfn          model.CfnResult

	// Security group editing panel (opened from the Detail panel)
	SGEditBtn         widget.Clickable // "Edit SGs" in the Detail panel header
	SGPanelOpen       bool
	SGCloseBtn        widget.Clickable
	SGInstanceID      string
	SGVpcID           string
	SGAttached        []string          // IDs currently attached to the instance
	SGAll             []model.SGSummary // all SGs in the instance's VPC
	SGAttachBoxes     []widget.Bool     // parallel to SGAll
	SGApplyBtn        widget.Clickable
	SGGroupDelBtns    []widget.Clickable // parallel to SGAll; delete group (only when not attached)
	SGGroupDelPending string             // group ID awaiting delete confirmation
	SGSelectBtns      []widget.Clickable // parallel to SGAttached
	SGSelected        string             // group whose rules are shown
	SGRules           []model.SGRule
	SGRuleDelBtns     []widget.Clickable // parallel to SGRules
	SGDelPending      int                // index awaiting delete confirmation, -1 = none
	SGRuleList        widget.List
	SGAttachList      widget.List // horizontal list of SG checkboxes
	SGDirBtn          widget.Clickable
	SGProtoBtn        widget.Clickable
	SGNewDir          string // ingress | egress
	SGNewProto        string // tcp | udp | icmp | all
	SGFromEd          widget.Editor
	SGToEd            widget.Editor
	SGCidrEd          widget.Editor
	SGDescEd          widget.Editor
	SGAddBtn          widget.Clickable
	SGErr             string
	// Create a new SG in the instance's VPC
	SGNewNameEd    widget.Editor
	SGNewGroupDesc widget.Editor
	SGCreateAttach widget.Bool // also attach the new group to the instance
	SGCreateBtn    widget.Clickable

	// AWS Batch view (TabBatch)
	BatchRefreshBtn    widget.Clickable
	BatchLoaded        bool // false until the first load, so the tab loads on first open
	BatchEnvs          []model.BatchComputeEnv
	BatchQueues        []model.BatchJobQueue
	BatchQueueBtns     []widget.Clickable // parallel to BatchQueues; click to list jobs
	BatchSelectedQueue string             // queue whose jobs are shown
	BatchJobs          []model.BatchJob
	BatchList          widget.List

	// Header status filter menu
	HeaderStatusBtn      widget.Clickable
	HeaderStatusFilter   string
	HeaderStatusMenuOpen bool
	HeaderMenuAll        widget.Clickable
	HeaderMenuRunning    widget.Clickable
	HeaderMenuStopped    widget.Clickable
	HeaderMenuOther      widget.Clickable

	// Column visibility customization
	ColVisible   []bool
	ColMenuOpen  bool
	ColMenuBtn   widget.Clickable
	ColMenuItems []widget.Clickable

	// Logger (nilable: no debug logging is emitted if nil)
	Logger *zap.SugaredLogger
}

// logDebug logs a debug message (no-op if Logger is nil)
func (s *AppState) logDebug(args ...interface{}) {
	if s.Logger != nil {
		s.Logger.Debug(args...)
	}
}

// logDebugf logs a formatted debug message
func (s *AppState) logDebugf(format string, args ...interface{}) {
	if s.Logger != nil {
		s.Logger.Debugf(format, args...)
	}
}

// InitStatusSlices initializes the status-management slices after instances are fetched
func (s *AppState) InitStatusSlices() {
	n := len(s.Instances)
	s.OriginalStatus = make([]string, n)
	s.DesiredStatus = make([]string, n)
	s.ToggleBtns = make([]widget.Clickable, n)
	s.SelectBoxes = make([]widget.Bool, n)
	// Create clickables for instance count x column count
	s.CellClickables = make([]widget.Clickable, n*len(Headers))
	for i, inst := range s.Instances {
		st := model.MapStatus(inst.Status)
		s.OriginalStatus[i] = st
		s.DesiredStatus[i] = st
	}
	// Set the dirty flag so the filter gets recomputed
	s.VisibleDirty = true

	// Reserve capacity for VisibleIndices
	if cap(s.VisibleIndices) < n {
		s.VisibleIndices = make([]int, 0, n)
	} else {
		s.VisibleIndices = s.VisibleIndices[:0]
	}

	// Set layout.Vertical (redundant since ListAxis is already set by the caller, but set it just in case)
	s.TableList.Axis = layout.Vertical
}

// EnsureColVisible initializes column visibility to "show all" if it hasn't been initialized yet
func (s *AppState) EnsureColVisible() {
	if len(s.ColVisible) == len(Headers) {
		return
	}
	s.ColVisible = make([]bool, len(Headers))
	for i := range s.ColVisible {
		s.ColVisible[i] = true
	}
	s.ColMenuItems = make([]widget.Clickable, len(Headers))
}

// HasStatusChanges returns whether any instance has a changed on/off status
func (s *AppState) HasStatusChanges() bool {
	for i := range s.OriginalStatus {
		if s.OriginalStatus[i] != "-" && s.DesiredStatus[i] != s.OriginalStatus[i] {
			return true
		}
	}
	return false
}

// SGIsAttached reports whether a group is currently attached to the SG panel's instance
func (s *AppState) SGIsAttached(id string) bool {
	for _, v := range s.SGAttached {
		if v == id {
			return true
		}
	}
	return false
}

// SGCanDelete reports whether the SG panel offers "Delete" for SGAll[i]:
// not attached to this instance and not the VPC's default group.
func (s *AppState) SGCanDelete(i int) bool {
	if i < 0 || i >= len(s.SGAll) {
		return false
	}
	return !s.SGIsAttached(s.SGAll[i].ID) && s.SGAll[i].Name != "default"
}

// SelectedInstanceIDs returns the IDs of instances checked in the Select column
func (s *AppState) SelectedInstanceIDs() []string {
	var ids []string
	for i := range s.SelectBoxes {
		if s.SelectBoxes[i].Value && i < len(s.Instances) {
			ids = append(ids, s.Instances[i].ID)
		}
	}
	return ids
}
