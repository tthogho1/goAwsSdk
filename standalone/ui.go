package main

import (
	"image/color"
	"strings"

	"gioui.org/font"
	"gioui.org/layout"
	"gioui.org/op/clip"
	"gioui.org/op/paint"
	"gioui.org/unit"
	"gioui.org/widget"
	"gioui.org/widget/material"
	"github.com/atotto/clipboard"
)

var (
	headerStatusBtn    widget.Clickable
	headerStatusFilter string
	headerStatusMenuOpen bool
	headerMenuAll        widget.Clickable
	headerMenuRunning    widget.Clickable
	headerMenuStopped    widget.Clickable
	headerMenuOther      widget.Clickable
)

// layoutTopBar はプロファイル入力欄と「取込」「実行」ボタンを描画
func layoutTopBar(gtx layout.Context, th *material.Theme) layout.Dimensions {
	return layout.UniformInset(unit.Dp(8)).Layout(gtx, func(gtx layout.Context) layout.Dimensions {
		return layout.Flex{Alignment: layout.Middle}.Layout(gtx,
			layout.Rigid(func(gtx layout.Context) layout.Dimensions {
				lbl := material.Body1(th, "プロファイル: ")
				return lbl.Layout(gtx)
			}),
			layout.Flexed(1, func(gtx layout.Context) layout.Dimensions {
				ed := material.Editor(th, &profileEditor, "AWSプロファイル名")
				return layout.UniformInset(unit.Dp(4)).Layout(gtx, func(gtx layout.Context) layout.Dimensions {
					return ed.Layout(gtx)
				})
			}),
			layout.Rigid(layout.Spacer{Width: unit.Dp(8)}.Layout),
			layout.Rigid(func(gtx layout.Context) layout.Dimensions {
				btn := material.Button(th, &fetchBtn, "取込")
				return btn.Layout(gtx)
			}),
			layout.Rigid(layout.Spacer{Width: unit.Dp(8)}.Layout),
			layout.Rigid(func(gtx layout.Context) layout.Dimensions {
				btn := material.Button(th, &executeBtn, "実行")
				if !hasStatusChanges() {
					btn.Background = color.NRGBA{R: 180, G: 180, B: 180, A: 255}
				}
				return btn.Layout(gtx)
			}),
		)
	})
}

// layoutMessage はエラーや情報メッセージを描画
func layoutMessage(gtx layout.Context, th *material.Theme, msg string, col color.NRGBA) layout.Dimensions {
	return layout.UniformInset(unit.Dp(8)).Layout(gtx, func(gtx layout.Context) layout.Dimensions {
		lbl := material.Body2(th, msg)
		lbl.Color = col
		return lbl.Layout(gtx)
	})
}

// layoutSearchBar はインスタンス名でフィルタする検索欄を描画
func layoutSearchBar(gtx layout.Context, th *material.Theme) layout.Dimensions {
	return layout.UniformInset(unit.Dp(8)).Layout(gtx, func(gtx layout.Context) layout.Dimensions {
		return layout.Flex{Alignment: layout.Middle}.Layout(gtx,
			layout.Flexed(1, func(gtx layout.Context) layout.Dimensions {
				searchEditor.SingleLine = true
				ed := material.Editor(th, &searchEditor, "Filter by name")
				return ed.Layout(gtx)
			}),
		)
	})
}

// layoutTable はインスタンス情報をテーブル形式で描画
func layoutTable(gtx layout.Context, th *material.Theme) layout.Dimensions {
	// handle on/off toggle button clicks
	for i := range toggleBtns {
		for toggleBtns[i].Clicked(gtx) {
			if i < len(desiredStatus) && i < len(originalStatus) && originalStatus[i] != "-" {
				if desiredStatus[i] == "on" {
					desiredStatus[i] = "off"
				} else {
					desiredStatus[i] = "on"
				}
				logDebugf("Toggle clicked for instance %d -> desiredStatus: %s", i, desiredStatus[i])
			}
		}
	}

	// handle per-cell clicks (copy to clipboard)
	cols := len(headers)
	for i := range cellClickables {
		for cellClickables[i].Clicked(gtx) {
			instIdx := i / cols
			colIdx := i % cols
			if instIdx < 0 || instIdx >= len(instances) {
				continue
			}
			var txt string
			inst := instances[instIdx]
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
				if instIdx < len(desiredStatus) {
					txt = desiredStatus[instIdx]
				} else {
					txt = ""
				}
			default:
				txt = ""
			}
			if txt == "" {
				errMsg = "コピー対象が空です"
				infoMsg = ""
				continue
			}
			if err := clipboard.WriteAll(txt); err != nil {
				errMsg = "クリップボードにコピーできませんでした: " + err.Error()
				infoMsg = ""
			} else {
				infoMsg = "Copied: " + txt
				errMsg = ""
			}
		}
	}

	// when header clicked, toggle menu open
	for headerStatusBtn.Clicked(gtx) {
		headerStatusMenuOpen = !headerStatusMenuOpen
		logDebugf("headerStatusBtn clicked; menu open: %t", headerStatusMenuOpen)
	}

	// Recompute visible indices when needed (keeps UI responsive and consistent).
	if visibleDirty {
		visibleIndices = visibleIndices[:0]
		for i := range instances {
			match := false
			if headerStatusFilter == "" {
				match = true
			} else {
				switch headerStatusFilter {
				case "running":
					match = instances[i].Status == "running"
				case "stopped":
					match = instances[i].Status == "stopped"
				case "other":
					match = mapStatus(instances[i].Status) == "-"
				}
			}
			// apply name search filter if present
			if match && searchQuery != "" {
				name := strings.ToLower(instances[i].Name)
				if !strings.Contains(name, strings.ToLower(searchQuery)) {
					match = false
				}
			}
			if match {
				visibleIndices = append(visibleIndices, i)
			}
		}
		visibleDirty = false
		logDebugf("Recomputed visibleIndices; filter=%s count=%d menuOpen=%t", headerStatusFilter, len(visibleIndices), headerStatusMenuOpen)
	}

	// debug print removed to reduce console noise

	// Build flex children: header + optional menu + data rows
	var flexChildren []layout.FlexChild

	// Header row
	flexChildren = append(flexChildren, layout.Rigid(func(gtx layout.Context) layout.Dimensions {
		return drawRowBackground(gtx, color.NRGBA{R: 220, G: 220, B: 240, A: 255}, func(gtx layout.Context) layout.Dimensions {
			return layout.UniformInset(unit.Dp(4)).Layout(gtx, func(gtx layout.Context) layout.Dimensions {
				children := make([]layout.FlexChild, len(headers))
				for i, h := range headers {
					colW := colWidths[i]
					idx := i
					children[i] = layout.Rigid(func(gtx layout.Context) layout.Dimensions {
						gtx.Constraints.Min.X = gtx.Dp(colW)
						gtx.Constraints.Max.X = gtx.Dp(colW)
						if idx == 1 {
							display := "Status"
							if headerStatusFilter != "" {
								display = display + " (" + headerStatusFilter + ")"
							}
							return headerStatusBtn.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
								lbl := material.Body2(th, display)
								lbl.Font.Weight = font.Bold
								if headerStatusFilter != "" {
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
				return layout.Flex{}.Layout(gtx, children...)
			})
		})
	}))

	// Selector menu (outside the list)
	if headerStatusMenuOpen {
		flexChildren = append(flexChildren, layout.Rigid(func(gtx layout.Context) layout.Dimensions {
			return drawRowBackground(gtx, color.NRGBA{R: 245, G: 245, B: 245, A: 255}, func(gtx layout.Context) layout.Dimensions {
				return layout.UniformInset(unit.Dp(4)).Layout(gtx, func(gtx layout.Context) layout.Dimensions {
				logDebug("Rendering selector row (outside list)")
				return layout.Flex{Alignment: layout.Middle}.Layout(gtx,
					layout.Rigid(func(gtx layout.Context) layout.Dimensions {
						logDebug("Render menu button: All")
						return headerMenuAll.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
							lbl := material.Body2(th, "All")
							if headerStatusFilter == "" {
								lbl.Color = color.NRGBA{R: 0, G: 80, B: 160, A: 255}
							}
							return lbl.Layout(gtx)
						})
					}),
					layout.Rigid(layout.Spacer{Width: unit.Dp(8)}.Layout),
					layout.Rigid(func(gtx layout.Context) layout.Dimensions {
						logDebug("Render menu button: Running")
						return headerMenuRunning.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
							lbl := material.Body2(th, "Running")
							if headerStatusFilter == "running" {
								lbl.Color = color.NRGBA{R: 0, G: 80, B: 160, A: 255}
							}
							return lbl.Layout(gtx)
						})
					}),
					layout.Rigid(layout.Spacer{Width: unit.Dp(8)}.Layout),
					layout.Rigid(func(gtx layout.Context) layout.Dimensions {
						logDebug("Render menu button: Stopped")
						return headerMenuStopped.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
							lbl := material.Body2(th, "Stopped")
							if headerStatusFilter == "stopped" {
								lbl.Color = color.NRGBA{R: 0, G: 80, B: 160, A: 255}
							}
							return lbl.Layout(gtx)
						})
					}),
					layout.Rigid(layout.Spacer{Width: unit.Dp(8)}.Layout),
					layout.Rigid(func(gtx layout.Context) layout.Dimensions {
						logDebug("Render menu button: Other")
						return headerMenuOther.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
							lbl := material.Body2(th, "Other")
							if headerStatusFilter == "other" {
								lbl.Color = color.NRGBA{R: 0, G: 80, B: 160, A: 255}
							}
							return lbl.Layout(gtx)
						})
					}),
				)
				})
			})
		}))

		// handle clicks for menu buttons
		for headerMenuAll.Clicked(gtx) {
			headerStatusFilter = ""
			headerStatusMenuOpen = false
			visibleDirty = true
			logDebug("Header menu: All selected")
		}
		for headerMenuRunning.Clicked(gtx) {
			headerStatusFilter = "running"
			headerStatusMenuOpen = false
			visibleDirty = true
			logDebug("Header menu: Running selected")
		}
		for headerMenuStopped.Clicked(gtx) {
			headerStatusFilter = "stopped"
			headerStatusMenuOpen = false
			visibleDirty = true
			logDebug("Header menu: Stopped selected")
		}
		for headerMenuOther.Clicked(gtx) {
			headerStatusFilter = "other"
			headerStatusMenuOpen = false
			visibleDirty = true
			logDebug("Header menu: Other selected")
		}
	}

	// Data rows
	flexChildren = append(flexChildren, layout.Flexed(1, func(gtx layout.Context) layout.Dimensions {
		return material.List(th, &tableList).Layout(gtx, len(visibleIndices), func(gtx layout.Context, idx int) layout.Dimensions {
		// idx maps to visibleIndices
		actualIdx := visibleIndices[idx]
		// zebra background
		bg := color.NRGBA{R: 255, G: 255, B: 255, A: 255}
		if idx%2 == 0 {
			bg = color.NRGBA{R: 245, G: 245, B: 245, A: 255}
		}
		return drawRowBackground(gtx, bg, func(gtx layout.Context) layout.Dimensions {
			return layout.UniformInset(unit.Dp(4)).Layout(gtx, func(gtx layout.Context) layout.Dimensions {
				inst := instances[actualIdx]
				statusDisplay := "-"
				if actualIdx < len(desiredStatus) {
					statusDisplay = desiredStatus[actualIdx]
					if desiredStatus[actualIdx] != originalStatus[actualIdx] {
						statusDisplay += "*"
					}
				}
				cells := []string{inst.ID, inst.Status, inst.InstanceType, inst.PrivateIP, inst.PublicIP, inst.Name, statusDisplay}
				children := make([]layout.FlexChild, len(cells))
				for i, cell := range cells {
					cellText := cell
					colW := colWidths[i]
					cellIdx := i
					children[i] = layout.Rigid(func(gtx layout.Context) layout.Dimensions {
						gtx.Constraints.Min.X = gtx.Dp(colW)
						gtx.Constraints.Max.X = gtx.Dp(colW)
						clickIdx := actualIdx*len(cells) + cellIdx
						// ensure clickIdx within range
						if clickIdx < 0 || clickIdx >= len(cellClickables) {
							lbl := material.Body2(th, cellText)
							lbl.MaxLines = 1
							return lbl.Layout(gtx)
						}
						// last column: keep toggle behavior, but wrap with cell clickable for copy
						if cellIdx == 6 && actualIdx < len(toggleBtns) {
							return cellClickables[clickIdx].Layout(gtx, func(gtx layout.Context) layout.Dimensions {
								return toggleBtns[actualIdx].Layout(gtx, func(gtx layout.Context) layout.Dimensions {
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
						// other columns: wrap label with clickable for copy
						return cellClickables[clickIdx].Layout(gtx, func(gtx layout.Context) layout.Dimensions {
							lbl := material.Body2(th, cellText)
							lbl.MaxLines = 1
							return lbl.Layout(gtx)
						})
					})
				}
				return layout.Flex{}.Layout(gtx, children...)
			})
		})
	})
	}))

	return layout.Flex{Axis: layout.Vertical}.Layout(gtx, flexChildren...)
}

// drawRowBackground は行の背景色を描画
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

