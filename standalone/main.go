package main

import (
	"fmt"
	"image/color"
	"os"
	"strings"

	"gioui.org/app"
	"gioui.org/layout"
	"gioui.org/op"
	"gioui.org/unit"
	"gioui.org/widget"
	"gioui.org/widget/material"
)

var (
	awsctrlPath string
	instances   []Instance
	headers     = []string{"ID", "Status", "Type", "PrivateIP", "PublicIP", "Name", "on/off"}
	colWidths   = []unit.Dp{180, 80, 100, 130, 130, 160, 60}

	originalStatus []string
	desiredStatus  []string
	toggleBtns     []widget.Clickable

	headerStatusBtn    widget.Clickable
	headerStatusFilter string
	headerStatusMenuOpen bool
	headerMenuAll        widget.Clickable
	headerMenuRunning    widget.Clickable
	headerMenuStopped    widget.Clickable
	headerMenuOther      widget.Clickable

	profileEditor widget.Editor
	fetchBtn      widget.Clickable
	executeBtn    widget.Clickable
	tableList     widget.List
	errMsg        string
	infoMsg       string

	// visibleIndices is a cached list of instance indices that match the current filter
	visibleIndices []int
	visibleDirty   bool
)

func main() {
	// .env 読み込み
	if err := loadEnv(".env"); err != nil {
		fmt.Fprintf(os.Stderr, ".envファイル読み込みエラー: %v\n", err)
		os.Exit(1)
	}
	if awsctrlPath == "" {
		fmt.Fprintln(os.Stderr, "AWSCTRL_PATH が .env に設定されていません")
		os.Exit(1)
	}

	profileEditor.SetText("default")
	profileEditor.SingleLine = true
	tableList.Axis = layout.Vertical

	go func() {
		w := new(app.Window)
		w.Option(app.Title("EC2 Instances Viewer"))
		w.Option(app.Size(unit.Dp(1020), unit.Dp(500)))
		if err := run(w); err != nil {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(1)
		}
		os.Exit(0)
	}()
	app.Main()
}

// run はウィンドウのイベントループ
func run(w *app.Window) error {
	th := material.NewTheme()
	var ops op.Ops

	for {
		switch e := w.Event().(type) {
		case app.DestroyEvent:
			return e.Err
		case app.FrameEvent:
			gtx := app.NewContext(&ops, e)

			// 「取込」ボタン押下処理
			if fetchBtn.Clicked(gtx) {
				profile := strings.TrimSpace(profileEditor.Text())
				if profile == "" {
					errMsg = "プロファイルを入力してください"
					infoMsg = ""
				} else {
					errMsg = ""
					infoMsg = ""
					output, err := executeAwsCtrl(profile)
					if err != nil {
						errMsg = fmt.Sprintf("awsctrl 実行エラー: %v", err)
					} else {
						instances = parseOutput(output)
						initStatusSlices()
						if len(instances) == 0 {
							infoMsg = "インスタンスが見つかりませんでした"
						}
					}
				}
			}

			// 「実行」ボタン押下処理
			if executeBtn.Clicked(gtx) && hasStatusChanges() {
				profile := strings.TrimSpace(profileEditor.Text())
				var errs []string
				for i := range instances {
					if desiredStatus[i] == originalStatus[i] || originalStatus[i] == "-" {
						continue
					}
					action := "up"
					if desiredStatus[i] == "off" {
						action = "down"
					}
					if err := executeAwsCtrlAction(profile, action, instances[i].ID); err != nil {
						errs = append(errs, fmt.Sprintf("%s: %v", instances[i].ID, err))
					}
				}
				if len(errs) > 0 {
					errMsg = "実行エラー: " + strings.Join(errs, "; ")
					infoMsg = ""
				} else {
					// 成功時: 再取得してステータス更新
					output, err := executeAwsCtrl(profile)
					if err != nil {
						errMsg = fmt.Sprintf("再取得エラー: %v", err)
					} else {
						instances = parseOutput(output)
						initStatusSlices()
						infoMsg = "実行完了"
						errMsg = ""
					}
				}
			}

			// メインレイアウト: 上部バー + メッセージ + テーブル
			layout.Flex{Axis: layout.Vertical}.Layout(gtx,
				layout.Rigid(func(gtx layout.Context) layout.Dimensions {
					return layoutTopBar(gtx, th)
				}),
				layout.Rigid(func(gtx layout.Context) layout.Dimensions {
					if errMsg != "" {
						return layoutMessage(gtx, th, errMsg, color.NRGBA{R: 220, G: 50, B: 50, A: 255})
					}
					if infoMsg != "" {
						return layoutMessage(gtx, th, infoMsg, color.NRGBA{R: 50, G: 50, B: 200, A: 255})
					}
					return layout.Dimensions{}
				}),
				layout.Flexed(1, func(gtx layout.Context) layout.Dimensions {
					return layoutTable(gtx, th)
				}),
			)

			e.Frame(gtx.Ops)
		}
	}
}

