package ui

import (
	"image/color"
	"strings"

	"ec2viewer/model"

	"gioui.org/font"
	"gioui.org/layout"
	"gioui.org/op/clip"
	"gioui.org/op/paint"
	"gioui.org/unit"
	"gioui.org/widget/material"
	"github.com/atotto/clipboard"
)

// LayoutTopBar はプロファイル入力欄と「取込」「実行」ボタンを描画
func LayoutTopBar(gtx layout.Context, th *material.Theme, s *AppState) layout.Dimensions {
	return layout.UniformInset(unit.Dp(8)).Layout(gtx, func(gtx layout.Context) layout.Dimensions {
		return layout.Flex{Alignment: layout.Middle}.Layout(gtx,
			layout.Rigid(func(gtx layout.Context) layout.Dimensions {
				lbl := material.Body1(th, "プロファイル: ")
				return lbl.Layout(gtx)
			}),
			layout.Flexed(1, func(gtx layout.Context) layout.Dimensions {
				ed := material.Editor(th, &s.ProfileEditor, "AWSプロファイル名")
				return layout.UniformInset(unit.Dp(4)).Layout(gtx, func(gtx layout.Context) layout.Dimensions {
					return ed.Layout(gtx)
				})
			}),
			layout.Rigid(layout.Spacer{Width: unit.Dp(8)}.Layout),
			layout.Rigid(func(gtx layout.Context) layout.Dimensions {
				btn := material.Button(th, &s.FetchBtn, "取込")
				return btn.Layout(gtx)
			}),
			layout.Rigid(layout.Spacer{Width: unit.Dp(8)}.Layout),
			layout.Rigid(func(gtx layout.Context) layout.Dimensions {
				btn := material.Button(th, &s.ExecuteBtn, "実行")
				if !s.HasStatusChanges() {
					btn.Background = color.NRGBA{R: 180, G: 180, B: 180, A: 255}
				}
				return btn.Layout(gtx)
			}),
		)
	})
}

// LayoutMessage はエラーや情報メッセージを描画
func LayoutMessage(gtx layout.Context, th *material.Theme, msg string, col color.NRGBA) layout.Dimensions {
	return layout.UniformInset(unit.Dp(8)).Layout(gtx, func(gtx layout.Context) layout.Dimensions {
		lbl := material.Body2(th, msg)
		lbl.Color = col
		return lbl.Layout(gtx)
	})
}

// LayoutSearchBar はインスタンス名でフィルタする検索欄を描画
func LayoutSearchBar(gtx layout.Context, th *material.Theme, s *AppState) layout.Dimensions {
	return layout.UniformInset(unit.Dp(8)).Layout(gtx, func(gtx layout.Context) layout.Dimensions {
		return layout.Flex{Alignment: layout.Middle}.Layout(gtx,
			layout.Flexed(1, func(gtx layout.Context) layout.Dimensions {
				s.SearchEditor.SingleLine = true
				ed := material.Editor(th, &s.SearchEditor, "Filter by name")
				return ed.Layout(gtx)
			}),
		)
	})
}

// LayoutTable はインスタンス情報をテーブル形式で描画
func LayoutTable(gtx layout.Context, th *material.Theme, s *AppState) layout.Dimensions {
	// on/off トグルボタンのクリック処理
	for i := range s.ToggleBtns {
		for s.ToggleBtns[i].Clicked(gtx) {
			if i < len(s.DesiredStatus) && i < len(s.OriginalStatus) && s.OriginalStatus[i] != "-" {
				if s.DesiredStatus[i] == "on" {
					s.DesiredStatus[i] = "off"
				} else {
					s.DesiredStatus[i] = "on"
				}
				s.logDebugf("Toggle clicked for instance %d -> desiredStatus: %s", i, s.DesiredStatus[i])
			}
		}
	}

	// セルクリック (クリップボードへコピー) の処理
	cols := len(Headers)
	for i := range s.CellClickables {
		for s.CellClickables[i].Clicked(gtx) {
			instIdx := i / cols
			colIdx := i % cols
			if instIdx < 0 || instIdx >= len(s.Instances) {
				continue
			}
			var txt string
			inst := s.Instances[instIdx]
			switch colIdx {
			case 0:
				txt = inst.ID
			case 1:
				txt = inst.Status
			case 2:
				txt = inst.InstanceType
			case 3:
				txt = inst.PrivateIP
			case 4:
				txt = inst.PublicIP
			case 5:
				txt = inst.Name
			case 6:
				if instIdx < len(s.DesiredStatus) {
					txt = s.DesiredStatus[instIdx]
				} else {
					txt = ""
				}
			default:
				txt = ""
			}
			if txt == "" {
				s.ErrMsg = "コピー対象が空です"
				s.InfoMsg = ""
				continue
			}
			if err := clipboard.WriteAll(txt); err != nil {
				s.ErrMsg = "クリップボードにコピーできませんでした: " + err.Error()
				s.InfoMsg = ""
			} else {
				s.InfoMsg = "Copied: " + txt
				s.ErrMsg = ""
			}
		}
	}

	// ヘッダーステータスボタンのクリックでメニュー開閉
	for s.HeaderStatusBtn.Clicked(gtx) {
		s.HeaderStatusMenuOpen = !s.HeaderStatusMenuOpen
		s.logDebugf("headerStatusBtn clicked; menu open: %t", s.HeaderStatusMenuOpen)
	}

	// 必要に応じて表示インデックスを再計算
	if s.VisibleDirty {
		s.VisibleIndices = s.VisibleIndices[:0]
		for i := range s.Instances {
			match := false
			if s.HeaderStatusFilter == "" {
				match = true
			} else {
				switch s.HeaderStatusFilter {
				case "running":
					match = s.Instances[i].Status == "running"
				case "stopped":
					match = s.Instances[i].Status == "stopped"
				case "other":
					match = model.MapStatus(s.Instances[i].Status) == "-"
				}
			}
			// 名前検索フィルタを適用
			if match && s.SearchQuery != "" {
				name := strings.ToLower(s.Instances[i].Name)
				if !strings.Contains(name, strings.ToLower(s.SearchQuery)) {
					match = false
				}
			}
			if match {
				s.VisibleIndices = append(s.VisibleIndices, i)
			}
		}
		s.VisibleDirty = false
		s.logDebugf("Recomputed visibleIndices; filter=%s count=%d menuOpen=%t", s.HeaderStatusFilter, len(s.VisibleIndices), s.HeaderStatusMenuOpen)
	}

	// フレックス子要素: ヘッダー + オプショナルメニュー + データ行
	var flexChildren []layout.FlexChild

	// ヘッダー行
	flexChildren = append(flexChildren, layout.Rigid(func(gtx layout.Context) layout.Dimensions {
		return drawRowBackground(gtx, color.NRGBA{R: 220, G: 220, B: 240, A: 255}, func(gtx layout.Context) layout.Dimensions {
			return layout.UniformInset(unit.Dp(4)).Layout(gtx, func(gtx layout.Context) layout.Dimensions {
				children := make([]layout.FlexChild, len(Headers))
				for i, h := range Headers {
					colW := ColWidths[i]
					idx := i
					children[i] = layout.Rigid(func(gtx layout.Context) layout.Dimensions {
						gtx.Constraints.Min.X = gtx.Dp(colW)
						gtx.Constraints.Max.X = gtx.Dp(colW)
						if idx == 1 {
							display := "Status"
							if s.HeaderStatusFilter != "" {
								display = display + " (" + s.HeaderStatusFilter + ")"
							}
							return s.HeaderStatusBtn.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
								lbl := material.Body2(th, display)
								lbl.Font.Weight = font.Bold
								if s.HeaderStatusFilter != "" {
									lbl.Color = color.NRGBA{R: 0, G: 80, B: 160, A: 255}
								}
								lbl.MaxLines = 1
								return lbl.Layout(gtx)
							})
						}
						lbl := material.Body2(th, h)
						lbl.Font.Weight = font.Bold
						lbl.MaxLines = 1
						return lbl.Layout(gtx)
					})
				}
				return layout.Center.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
					return layout.Flex{}.Layout(gtx, children...)
				})
			})
		})
	}))

	// セレクタメニュー (リストの外側)
	if s.HeaderStatusMenuOpen {
		flexChildren = append(flexChildren, layout.Rigid(func(gtx layout.Context) layout.Dimensions {
			return drawRowBackground(gtx, color.NRGBA{R: 245, G: 245, B: 245, A: 255}, func(gtx layout.Context) layout.Dimensions {
				return layout.UniformInset(unit.Dp(4)).Layout(gtx, func(gtx layout.Context) layout.Dimensions {
					s.logDebug("Rendering selector row (outside list)")
					return layout.Flex{Alignment: layout.Middle}.Layout(gtx,
						layout.Rigid(func(gtx layout.Context) layout.Dimensions {
							s.logDebug("Render menu button: All")
							return s.HeaderMenuAll.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
								lbl := material.Body2(th, "All")
								if s.HeaderStatusFilter == "" {
									lbl.Color = color.NRGBA{R: 0, G: 80, B: 160, A: 255}
								}
								return lbl.Layout(gtx)
							})
						}),
						layout.Rigid(layout.Spacer{Width: unit.Dp(8)}.Layout),
						layout.Rigid(func(gtx layout.Context) layout.Dimensions {
							s.logDebug("Render menu button: Running")
							return s.HeaderMenuRunning.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
								lbl := material.Body2(th, "Running")
								if s.HeaderStatusFilter == "running" {
									lbl.Color = color.NRGBA{R: 0, G: 80, B: 160, A: 255}
								}
								return lbl.Layout(gtx)
							})
						}),
						layout.Rigid(layout.Spacer{Width: unit.Dp(8)}.Layout),
						layout.Rigid(func(gtx layout.Context) layout.Dimensions {
							s.logDebug("Render menu button: Stopped")
							return s.HeaderMenuStopped.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
								lbl := material.Body2(th, "Stopped")
								if s.HeaderStatusFilter == "stopped" {
									lbl.Color = color.NRGBA{R: 0, G: 80, B: 160, A: 255}
								}
								return lbl.Layout(gtx)
							})
						}),
						layout.Rigid(layout.Spacer{Width: unit.Dp(8)}.Layout),
						layout.Rigid(func(gtx layout.Context) layout.Dimensions {
							s.logDebug("Render menu button: Other")
							return s.HeaderMenuOther.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
								lbl := material.Body2(th, "Other")
								if s.HeaderStatusFilter == "other" {
									lbl.Color = color.NRGBA{R: 0, G: 80, B: 160, A: 255}
								}
								return lbl.Layout(gtx)
							})
						}),
					)
				})
			})
		}))

		// メニューボタンのクリック処理
		for s.HeaderMenuAll.Clicked(gtx) {
			s.HeaderStatusFilter = ""
			s.HeaderStatusMenuOpen = false
			s.VisibleDirty = true
			s.logDebug("Header menu: All selected")
		}
		for s.HeaderMenuRunning.Clicked(gtx) {
			s.HeaderStatusFilter = "running"
			s.HeaderStatusMenuOpen = false
			s.VisibleDirty = true
			s.logDebug("Header menu: Running selected")
		}
		for s.HeaderMenuStopped.Clicked(gtx) {
			s.HeaderStatusFilter = "stopped"
			s.HeaderStatusMenuOpen = false
			s.VisibleDirty = true
			s.logDebug("Header menu: Stopped selected")
		}
		for s.HeaderMenuOther.Clicked(gtx) {
			s.HeaderStatusFilter = "other"
			s.HeaderStatusMenuOpen = false
			s.VisibleDirty = true
			s.logDebug("Header menu: Other selected")
		}
	}

	// データ行
	flexChildren = append(flexChildren, layout.Flexed(1, func(gtx layout.Context) layout.Dimensions {
		return material.List(th, &s.TableList).Layout(gtx, len(s.VisibleIndices), func(gtx layout.Context, idx int) layout.Dimensions {
			actualIdx := s.VisibleIndices[idx]
			// ゼブラ縞の背景
			bg := color.NRGBA{R: 255, G: 255, B: 255, A: 255}
			if idx%2 == 0 {
				bg = color.NRGBA{R: 245, G: 245, B: 245, A: 255}
			}
			return drawRowBackground(gtx, bg, func(gtx layout.Context) layout.Dimensions {
				return layout.UniformInset(unit.Dp(4)).Layout(gtx, func(gtx layout.Context) layout.Dimensions {
					inst := s.Instances[actualIdx]
					statusDisplay := "-"
					if actualIdx < len(s.DesiredStatus) {
						statusDisplay = s.DesiredStatus[actualIdx]
						if s.DesiredStatus[actualIdx] != s.OriginalStatus[actualIdx] {
							statusDisplay += "*"
						}
					}
					cells := []string{inst.ID, inst.Status, inst.InstanceType, inst.PrivateIP, inst.PublicIP, inst.Name, statusDisplay}
					children := make([]layout.FlexChild, len(cells))
					for i, cell := range cells {
						cellText := cell
						colW := ColWidths[i]
						cellIdx := i
						children[i] = layout.Rigid(func(gtx layout.Context) layout.Dimensions {
							gtx.Constraints.Min.X = gtx.Dp(colW)
							gtx.Constraints.Max.X = gtx.Dp(colW)
							clickIdx := actualIdx*len(cells) + cellIdx
							if clickIdx < 0 || clickIdx >= len(s.CellClickables) {
								lbl := material.Body2(th, cellText)
								lbl.MaxLines = 1
								return lbl.Layout(gtx)
							}
							// 最終カラム: トグル + コピー
							if cellIdx == 6 && actualIdx < len(s.ToggleBtns) {
								return s.CellClickables[clickIdx].Layout(gtx, func(gtx layout.Context) layout.Dimensions {
									return s.ToggleBtns[actualIdx].Layout(gtx, func(gtx layout.Context) layout.Dimensions {
										lbl := material.Body2(th, cellText)
										lbl.Font.Weight = font.Bold
										switch {
										case strings.HasPrefix(cellText, "on"):
											lbl.Color = color.NRGBA{R: 0, G: 160, B: 0, A: 255}
										case strings.HasPrefix(cellText, "off"):
											lbl.Color = color.NRGBA{R: 220, G: 50, B: 50, A: 255}
										default:
											lbl.Color = color.NRGBA{R: 150, G: 150, B: 150, A: 255}
										}
										lbl.MaxLines = 1
										return lbl.Layout(gtx)
									})
								})
							}
							// その他カラム: コピー用クリッカブルでラップ
							return s.CellClickables[clickIdx].Layout(gtx, func(gtx layout.Context) layout.Dimensions {
								lbl := material.Body2(th, cellText)
								lbl.MaxLines = 1
								return lbl.Layout(gtx)
							})
						})
					}
					return layout.Center.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
						return layout.Flex{}.Layout(gtx, children...)
					})
				})
			})
		})
	}))

	return layout.Flex{Axis: layout.Vertical}.Layout(gtx, flexChildren...)
}

// drawRowBackground は行の背景色を描画する
func drawRowBackground(gtx layout.Context, col color.NRGBA, w layout.Widget) layout.Dimensions {
	return layout.Stack{}.Layout(gtx,
		layout.Expanded(func(gtx layout.Context) layout.Dimensions {
			sz := gtx.Constraints.Min
			paint.FillShape(gtx.Ops, col, clip.Rect{Max: sz}.Op())
			return layout.Dimensions{Size: sz}
		}),
		layout.Stacked(w),
	)
}
