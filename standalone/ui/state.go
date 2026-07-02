package ui

import (
	"ec2viewer/model"

	"gioui.org/layout"
	"gioui.org/unit"
	"gioui.org/widget"
	"go.uber.org/zap"
)

// Headers はテーブルのカラム名
var Headers = []string{"ID", "Status", "Type", "PrivateIP", "PublicIP", "Name", "on/off"}

// ColWidths はテーブルの各カラム幅
var ColWidths = []unit.Dp{180, 80, 100, 130, 130, 160, 60}

// AppState はUIおよびデータの状態を一元管理する構造体
type AppState struct {
	// インスタンスデータ
	Instances      []model.Instance
	OriginalStatus []string
	DesiredStatus  []string
	ToggleBtns     []widget.Clickable
	CellClickables []widget.Clickable

	// フィルタ・表示
	VisibleIndices []int
	VisibleDirty   bool

	// UI ウィジェット
	ProfileEditor widget.Editor
	SearchEditor  widget.Editor
	SearchQuery   string
	FetchBtn      widget.Clickable
	ExecuteBtn    widget.Clickable
	TableList     widget.List

	// メッセージ
	ErrMsg  string
	InfoMsg string

	// ヘッダーステータスフィルタメニュー
	HeaderStatusBtn      widget.Clickable
	HeaderStatusFilter   string
	HeaderStatusMenuOpen bool
	HeaderMenuAll        widget.Clickable
	HeaderMenuRunning    widget.Clickable
	HeaderMenuStopped    widget.Clickable
	HeaderMenuOther      widget.Clickable

	// ロガー (nil 可: nil の場合デバッグログを出力しない)
	Logger *zap.SugaredLogger
}

// logDebug はデバッグメッセージを記録する (Logger が nil の場合はno-op)
func (s *AppState) logDebug(args ...interface{}) {
	if s.Logger != nil {
		s.Logger.Debug(args...)
	}
}

// logDebugf はフォーマット付きデバッグメッセージを記録する
func (s *AppState) logDebugf(format string, args ...interface{}) {
	if s.Logger != nil {
		s.Logger.Debugf(format, args...)
	}
}

// InitStatusSlices はインスタンス取得後にステータス管理用スライスを初期化する
func (s *AppState) InitStatusSlices() {
	n := len(s.Instances)
	s.OriginalStatus = make([]string, n)
	s.DesiredStatus = make([]string, n)
	s.ToggleBtns = make([]widget.Clickable, n)
	// インスタンス数 × カラム数 のクリッカブルを生成
	s.CellClickables = make([]widget.Clickable, n*len(Headers))
	for i, inst := range s.Instances {
		st := model.MapStatus(inst.Status)
		s.OriginalStatus[i] = st
		s.DesiredStatus[i] = st
	}
	// フィルタを再計算するためにダーティフラグを立てる
	s.VisibleDirty = true

	// VisibleIndices のスライスを確保しておく
	if cap(s.VisibleIndices) < n {
		s.VisibleIndices = make([]int, 0, n)
	} else {
		s.VisibleIndices = s.VisibleIndices[:0]
	}

	// layout.Vertical を設定 (ListAxis は呼び出し元で設定済みのため不要だが念のため)
	s.TableList.Axis = layout.Vertical
}

// HasStatusChanges はon/off が変更されたインスタンスがあるかを返す
func (s *AppState) HasStatusChanges() bool {
	for i := range s.OriginalStatus {
		if s.OriginalStatus[i] != "-" && s.DesiredStatus[i] != s.OriginalStatus[i] {
			return true
		}
	}
	return false
}
