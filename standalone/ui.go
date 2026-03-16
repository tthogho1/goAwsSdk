package main

import (
	"fmt"
	"image/color"
	"strings"

	"gioui.org/font"
	"gioui.org/layout"
	"gioui.org/op/clip"
	"gioui.org/op/paint"
	"gioui.org/unit"
	"gioui.org/widget/material"
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

// layoutTable はインスタンス情報をテーブル形式で描画
func layoutTable(gtx layout.Context, th *material.Theme) layout.Dimensions {
	// when header clicked, toggle menu open
	for headerStatusBtn.Clicked(gtx) {
		headerStatusMenuOpen = !headerStatusMenuOpen
		fmt.Println("headerStatusBtn clicked; menu open:", headerStatusMenuOpen)
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
			if match {
				visibleIndices = append(visibleIndices, i)
			}
		}
		visibleDirty = false
		fmt.Println("Recomputed visibleIndices; filter=", headerStatusFilter, "count=", len(visibleIndices))
	}

	// header extra row when menu open
	headerExtraRows := 0
	if headerStatusMenuOpen {
		headerExtraRows = 1
	}
	totalRows := 1 + headerExtraRows + len(visibleIndices)

	return material.List(th, &tableList).Layout(gtx, totalRows, func(gtx layout.Context, index int) layout.Dimensions {
		isHeader := index == 0

		// 行の背景色（デフォルト） — will adjust after we know visible-row index
		bgColor := color.NRGBA{R: 255, G: 255, B: 255, A: 255}

		// データ行のon/offトグル処理
		// index mapping:
		// 0 = header
		// 1 = optional menu row (if open)
		// data rows start at offset = 1 + headerExtraRows
		rowIdx := index - 1 - headerExtraRows
		isMenu := headerStatusMenuOpen && index == 1

		if isMenu {
			// render selector row
			return drawRowBackground(gtx, bgColor, func(gtx layout.Context) layout.Dimensions {
				return layout.UniformInset(unit.Dp(4)).Layout(gtx, func(gtx layout.Context) layout.Dimensions {
					// Buttons: All, Running, Stopped, Other
					dims := layout.Flex{Alignment: layout.Middle}.Layout(gtx,
						layout.Rigid(func(gtx layout.Context) layout.Dimensions {
							btn := material.Button(th, &headerMenuAll, "All")
							if headerStatusFilter == "" {
								btn.Background = color.NRGBA{R: 200, G: 220, B: 255, A: 255}
							}
							return btn.Layout(gtx)
						}),
						layout.Rigid(layout.Spacer{Width: unit.Dp(8)}.Layout),
						layout.Rigid(func(gtx layout.Context) layout.Dimensions {
							btn := material.Button(th, &headerMenuRunning, "Running")
							if headerStatusFilter == "running" {
								btn.Background = color.NRGBA{R: 200, G: 220, B: 255, A: 255}
							}
							return btn.Layout(gtx)
						}),
						layout.Rigid(layout.Spacer{Width: unit.Dp(8)}.Layout),
						layout.Rigid(func(gtx layout.Context) layout.Dimensions {
							btn := material.Button(th, &headerMenuStopped, "Stopped")
							if headerStatusFilter == "stopped" {
								btn.Background = color.NRGBA{R: 200, G: 220, B: 255, A: 255}
							}
							return btn.Layout(gtx)
						}),
						layout.Rigid(layout.Spacer{Width: unit.Dp(8)}.Layout),
						layout.Rigid(func(gtx layout.Context) layout.Dimensions {
							btn := material.Button(th, &headerMenuOther, "Other")
							if headerStatusFilter == "other" {
								btn.Background = color.NRGBA{R: 200, G: 220, B: 255, A: 255}
							}
							return btn.Layout(gtx)
						}),
					)

					// handle clicks
					for headerMenuAll.Clicked(gtx) {
						headerStatusFilter = ""
						headerStatusMenuOpen = false
						visibleDirty = true
						fmt.Println("Header menu: All selected")
					}
					for headerMenuRunning.Clicked(gtx) {
						headerStatusFilter = "running"
						headerStatusMenuOpen = false
						visibleDirty = true
						fmt.Println("Header menu: Running selected")
					}
					for headerMenuStopped.Clicked(gtx) {
						headerStatusFilter = "stopped"
						headerStatusMenuOpen = false
						visibleDirty = true
						fmt.Println("Header menu: Stopped selected")
					}
					for headerMenuOther.Clicked(gtx) {
						headerStatusFilter = "other"
						headerStatusMenuOpen = false
						visibleDirty = true
						fmt.Println("Header menu: Other selected")
					}

					return dims
				})
			})
		}

		// map to actual instance index using cached visibleIndices
		var actualIdx int
		displayIdx := rowIdx
		if !isHeader {
			if rowIdx < 0 || rowIdx >= len(visibleIndices) {
				return layout.Dimensions{}
			}
			actualIdx = visibleIndices[rowIdx]
			if actualIdx >= 0 && actualIdx < len(toggleBtns) {
				for toggleBtns[actualIdx].Clicked(gtx) {
					if desiredStatus[actualIdx] == "on" {
						desiredStatus[actualIdx] = "off"
					} else if desiredStatus[actualIdx] == "off" {
						desiredStatus[actualIdx] = "on"
					}
				}
			}
		}

		// Determine zebra striping based on the visible/display index
		if isHeader {
			bgColor = color.NRGBA{R: 220, G: 220, B: 240, A: 255}
		} else {
			if displayIdx%2 == 0 {
				bgColor = color.NRGBA{R: 245, G: 245, B: 245, A: 255}
			} else {
				bgColor = color.NRGBA{R: 255, G: 255, B: 255, A: 255}
			}
		}

		return drawRowBackground(gtx, bgColor, func(gtx layout.Context) layout.Dimensions {
			return layout.UniformInset(unit.Dp(4)).Layout(gtx, func(gtx layout.Context) layout.Dimensions {
				var cells []string
				if isHeader {
					cells = headers
				} else {
					inst := instances[actualIdx]
					statusDisplay := "-"
					if actualIdx < len(desiredStatus) {
						statusDisplay = desiredStatus[actualIdx]
						if desiredStatus[actualIdx] != originalStatus[actualIdx] {
							statusDisplay += "*"
						}
					}
					cells = []string{inst.ID, inst.Status, inst.InstanceType, inst.PrivateIP, inst.PublicIP, inst.Name, statusDisplay}
				}

				children := make([]layout.FlexChild, len(cells))
				for i, cell := range cells {
					cellText := cell
					colW := colWidths[i]
					bold := isHeader
					cellIdx := i
					children[i] = layout.Rigid(func(gtx layout.Context) layout.Dimensions {
						gtx.Constraints.Min.X = gtx.Dp(colW)
						gtx.Constraints.Max.X = gtx.Dp(colW)

						// header status clickable
						if isHeader && cellIdx == 1 {
							// display current filter
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

						// on/off列（最終列）の特殊レンダリング
						if cellIdx == 6 && !isHeader && actualIdx < len(toggleBtns) {
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
						}

						lbl := material.Body2(th, cellText)
						if bold {
							lbl.Font.Weight = font.Bold
						}
						lbl.MaxLines = 1
						return lbl.Layout(gtx)
					})
				}
				return layout.Flex{}.Layout(gtx, children...)
			})
		})
	})
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
