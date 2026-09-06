package ui

import (
	"ec2viewer/model"

	"gioui.org/layout"
	"gioui.org/unit"
	"gioui.org/widget"
	"go.uber.org/zap"
)

// Headers are the table column names
var Headers = []string{"ID", "Status", "Type", "PrivateIP", "PublicIP", "Name", "on/off"}

// ColWidths are the widths of each table column
var ColWidths = []unit.Dp{180, 80, 100, 130, 130, 160, 60}

// AppState is a struct that centrally manages the UI and data state
type AppState struct {
	// Instance data
	Instances      []model.Instance
	OriginalStatus []string
	DesiredStatus  []string
	ToggleBtns     []widget.Clickable
	CellClickables []widget.Clickable

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
