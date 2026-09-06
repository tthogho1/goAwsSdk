package ui

import (
	"fmt"
	"image"
	"image/color"
	"strings"

	"ec2viewer/model"

	"gioui.org/font"
	"gioui.org/layout"
	"gioui.org/op"
	"gioui.org/op/clip"
	"gioui.org/op/paint"
	"gioui.org/unit"
	"gioui.org/widget/material"
	"github.com/atotto/clipboard"
)

// LayoutTopBar draws the profile input field and the "Fetch"/"Execute" buttons
func LayoutTopBar(gtx layout.Context, th *material.Theme, s *AppState) layout.Dimensions {
	// Toggle the profile menu open/closed when the profile button is clicked
	for s.ProfileMenuBtn.Clicked(gtx) {
		s.ProfileMenuOpen = !s.ProfileMenuOpen
	}
	// Handle profile selection from the menu
	for i := range s.ProfileMenuItems {
		for s.ProfileMenuItems[i].Clicked(gtx) {
			if i < len(s.Profiles) {
				s.SelectedProfile = s.Profiles[i]
			}
			s.ProfileMenuOpen = false
		}
	}

	return layout.Flex{Axis: layout.Vertical}.Layout(gtx,
		layout.Rigid(func(gtx layout.Context) layout.Dimensions {
			return layout.UniformInset(unit.Dp(4)).Layout(gtx, func(gtx layout.Context) layout.Dimensions {
				return layout.Flex{Alignment: layout.Middle}.Layout(gtx,
					layout.Rigid(func(gtx layout.Context) layout.Dimensions {
						lbl := material.Body1(th, "Profile: ")
						return lbl.Layout(gtx)
					}),
					layout.Rigid(func(gtx layout.Context) layout.Dimensions {
						return s.ProfileMenuBtn.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
							text := s.SelectedProfile
							if text == "" {
								text = "(select profile)"
							}
							lbl := material.Body1(th, text+" v")
							return layout.UniformInset(unit.Dp(6)).Layout(gtx, lbl.Layout)
						})
					}),
					layout.Flexed(1, func(gtx layout.Context) layout.Dimensions {
						return layout.Dimensions{Size: image.Point{X: gtx.Constraints.Max.X, Y: gtx.Constraints.Min.Y}}
					}),
					layout.Rigid(layout.Spacer{Width: unit.Dp(8)}.Layout),
					layout.Rigid(func(gtx layout.Context) layout.Dimensions {
						btn := material.Button(th, &s.FetchBtn, "Fetch")
						btn.Inset = layout.UniformInset(unit.Dp(6))
						return btn.Layout(gtx)
					}),
					layout.Rigid(layout.Spacer{Width: unit.Dp(8)}.Layout),
					layout.Rigid(func(gtx layout.Context) layout.Dimensions {
						btn := material.Button(th, &s.ExecuteBtn, "Execute")
						btn.Inset = layout.UniformInset(unit.Dp(6))
						if !s.HasStatusChanges() {
							btn.Background = color.NRGBA{R: 180, G: 180, B: 180, A: 255}
						}
						return btn.Layout(gtx)
					}),
				)
			})
		}),
		layout.Rigid(func(gtx layout.Context) layout.Dimensions {
			if !s.ProfileMenuOpen {
				return layout.Dimensions{}
			}
			return drawRowBackground(gtx, color.NRGBA{R: 245, G: 245, B: 245, A: 255}, func(gtx layout.Context) layout.Dimensions {
				return layout.UniformInset(unit.Dp(4)).Layout(gtx, func(gtx layout.Context) layout.Dimensions {
					var children []layout.FlexChild
					for i, p := range s.Profiles {
						idx := i
						name := p
						children = append(children, layout.Rigid(func(gtx layout.Context) layout.Dimensions {
							return s.ProfileMenuItems[idx].Layout(gtx, func(gtx layout.Context) layout.Dimensions {
								lbl := material.Body2(th, name)
								if name == s.SelectedProfile {
									lbl.Color = color.NRGBA{R: 0, G: 80, B: 160, A: 255}
									lbl.Font.Weight = font.Bold
								}
								lbl.MaxLines = 1
								return layout.UniformInset(unit.Dp(4)).Layout(gtx, lbl.Layout)
							})
						}))
					}
					return layout.Flex{Axis: layout.Vertical}.Layout(gtx, children...)
				})
			})
		}),
	)
}

// LayoutMessage draws an error or informational message
func LayoutMessage(gtx layout.Context, th *material.Theme, msg string, col color.NRGBA) layout.Dimensions {
	return layout.UniformInset(unit.Dp(8)).Layout(gtx, func(gtx layout.Context) layout.Dimensions {
		lbl := material.Body2(th, msg)
		lbl.Color = col
		return lbl.Layout(gtx)
	})
}

// LayoutSearchBar draws the search field for filtering by instance name
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

// LayoutTable draws the instance information in a table format
func LayoutTable(gtx layout.Context, th *material.Theme, s *AppState) layout.Dimensions {
	s.EnsureColVisible()

	// Handle clicks on the on/off toggle buttons
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

	// Handle cell clicks (copy to clipboard)
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
				s.ErrMsg = "Nothing to copy"
				s.InfoMsg = ""
				continue
			}
			if err := clipboard.WriteAll(txt); err != nil {
				s.ErrMsg = "Failed to copy to clipboard: " + err.Error()
				s.InfoMsg = ""
			} else {
				s.InfoMsg = "Copied: " + txt
				s.ErrMsg = ""
			}
		}
	}

	// Toggle the menu open/closed when the header status button is clicked
	for s.HeaderStatusBtn.Clicked(gtx) {
		s.HeaderStatusMenuOpen = !s.HeaderStatusMenuOpen
		s.logDebugf("headerStatusBtn clicked; menu open: %t", s.HeaderStatusMenuOpen)
	}

	// Toggle the column visibility menu open/closed when clicked
	for s.ColMenuBtn.Clicked(gtx) {
		s.ColMenuOpen = !s.ColMenuOpen
	}
	// Toggle column visibility when a column-visibility menu item is clicked (the ID column is excluded since it's required)
	for i := range s.ColMenuItems {
		for s.ColMenuItems[i].Clicked(gtx) {
			if i == 0 {
				continue
			}
			s.ColVisible[i] = !s.ColVisible[i]
		}
	}

	// Recompute visible indices if necessary
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
			// Apply the name search filter
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

	// Flex children: header + optional menu + data rows
	var flexChildren []layout.FlexChild

	// Header row
	flexChildren = append(flexChildren, layout.Rigid(func(gtx layout.Context) layout.Dimensions {
		return drawRowBackground(gtx, color.NRGBA{R: 220, G: 220, B: 240, A: 255}, func(gtx layout.Context) layout.Dimensions {
			return layout.UniformInset(unit.Dp(4)).Layout(gtx, func(gtx layout.Context) layout.Dimensions {
				var children []layout.FlexChild
				for i, h := range Headers {
					if !s.ColVisible[i] {
						continue
					}
					colW := ColWidths[i]
					idx := i
					children = append(children, layout.Rigid(func(gtx layout.Context) layout.Dimensions {
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
					}))
				}
				children = append(children, layout.Rigid(layout.Spacer{Width: unit.Dp(8)}.Layout))
				children = append(children, layout.Rigid(func(gtx layout.Context) layout.Dimensions {
					return s.ColMenuBtn.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
						// Use ASCII only (symbols like ▾ inflate row height via fallback font)
						lbl := material.Body2(th, "Columns v")
						lbl.Font.Weight = font.Bold
						lbl.MaxLines = 1
						return lbl.Layout(gtx)
					})
				}))
				return layout.Flex{}.Layout(gtx, children...)
			})
		})
	}))

	// Column visibility toggle menu
	if s.ColMenuOpen {
		flexChildren = append(flexChildren, layout.Rigid(func(gtx layout.Context) layout.Dimensions {
			return drawRowBackground(gtx, color.NRGBA{R: 245, G: 245, B: 245, A: 255}, func(gtx layout.Context) layout.Dimensions {
				return layout.UniformInset(unit.Dp(4)).Layout(gtx, func(gtx layout.Context) layout.Dimensions {
					var children []layout.FlexChild
					for i, h := range Headers {
						idx := i
						label := h
						children = append(children, layout.Rigid(func(gtx layout.Context) layout.Dimensions {
							return s.ColMenuItems[idx].Layout(gtx, func(gtx layout.Context) layout.Dimensions {
								mark := "[ ] "
								if s.ColVisible[idx] {
									mark = "[x] "
								}
								lbl := material.Body2(th, mark+label)
								lbl.MaxLines = 1
								if idx == 0 {
									lbl.Color = color.NRGBA{R: 150, G: 150, B: 150, A: 255}
								}
								return lbl.Layout(gtx)
							})
						}))
						children = append(children, layout.Rigid(layout.Spacer{Width: unit.Dp(8)}.Layout))
					}
					return layout.Flex{Alignment: layout.Middle}.Layout(gtx, children...)
				})
			})
		}))
	}

	// Selector menu (outside the list)
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

		// Handle menu button clicks
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

	// Data rows
	flexChildren = append(flexChildren, layout.Flexed(1, func(gtx layout.Context) layout.Dimensions {
		return material.List(th, &s.TableList).Layout(gtx, len(s.VisibleIndices), func(gtx layout.Context, idx int) layout.Dimensions {
			actualIdx := s.VisibleIndices[idx]
			// Zebra-stripe background
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
					var children []layout.FlexChild
					for i, cell := range cells {
						if !s.ColVisible[i] {
							continue
						}
						cellText := cell
						colW := ColWidths[i]
						cellIdx := i
						children = append(children, layout.Rigid(func(gtx layout.Context) layout.Dimensions {
							gtx.Constraints.Min.X = gtx.Dp(colW)
							gtx.Constraints.Max.X = gtx.Dp(colW)
							clickIdx := actualIdx*len(cells) + cellIdx
							if clickIdx < 0 || clickIdx >= len(s.CellClickables) {
								lbl := material.Body2(th, cellText)
								lbl.MaxLines = 1
								return lbl.Layout(gtx)
							}
							// Final column: toggle + copy
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
							// Other columns: wrap with a clickable for copying
							return s.CellClickables[clickIdx].Layout(gtx, func(gtx layout.Context) layout.Dimensions {
								lbl := material.Body2(th, cellText)
								lbl.MaxLines = 1
								return lbl.Layout(gtx)
							})
						}))
					}
					return layout.Flex{}.Layout(gtx, children...)
				})
			})
		})
	}))

	// Center the table horizontally (hidden columns are excluded from the width calculation)
	tableWidth := 0
	for i, w := range ColWidths {
		if !s.ColVisible[i] {
			continue
		}
		tableWidth += gtx.Dp(w)
	}
	tableWidth += 2 * gtx.Dp(unit.Dp(4)) // UniformInset(4) left + right

	availW := gtx.Constraints.Max.X
	if availW > tableWidth {
		leftPx := (availW - tableWidth) / 2
		macro := op.Record(gtx.Ops)
		gtx.Constraints.Max.X = tableWidth
		gtx.Constraints.Min.X = tableWidth
		dims := layout.Flex{Axis: layout.Vertical}.Layout(gtx, flexChildren...)
		call := macro.Stop()
		stack := op.Offset(image.Point{X: leftPx}).Push(gtx.Ops)
		call.Add(gtx.Ops)
		stack.Pop()
		dims.Size.X = availW
		return dims
	}
	return layout.Flex{Axis: layout.Vertical}.Layout(gtx, flexChildren...)
}

// LayoutFooter draws the footer showing the instance count and hints
func LayoutFooter(gtx layout.Context, th *material.Theme, s *AppState) layout.Dimensions {
	return drawRowBackground(gtx, color.NRGBA{R: 220, G: 220, B: 240, A: 255}, func(gtx layout.Context) layout.Dimensions {
		return layout.UniformInset(unit.Dp(6)).Layout(gtx, func(gtx layout.Context) layout.Dimensions {
			total := len(s.Instances)
			visible := len(s.VisibleIndices)
			countText := fmt.Sprintf("Showing: %d / %d total", visible, total)
			return layout.Flex{Alignment: layout.Middle}.Layout(gtx,
				layout.Rigid(func(gtx layout.Context) layout.Dimensions {
					lbl := material.Body2(th, countText)
					return lbl.Layout(gtx)
				}),
				layout.Flexed(1, layout.Spacer{}.Layout),
				layout.Rigid(func(gtx layout.Context) layout.Dimensions {
					lbl := material.Body2(th, "Click a cell to copy to clipboard")
					lbl.Color = color.NRGBA{R: 100, G: 100, B: 100, A: 255}
					return lbl.Layout(gtx)
				}),
			)
		})
	})
}

// drawRowBackground draws the background color of a row
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
