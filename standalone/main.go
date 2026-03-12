package main

import (
	"bufio"
	"fmt"
	"image/color"
	"os"
	"os/exec"
	"strings"

	"gioui.org/app"
	"gioui.org/font"
	"gioui.org/layout"
	"gioui.org/op"
	"gioui.org/op/clip"
	"gioui.org/op/paint"
	"gioui.org/unit"
	"gioui.org/widget"
	"gioui.org/widget/material"
)

// Instance はEC2インスタンス情報を保持する構造体
type Instance struct {
	ID           string
	Status       string
	InstanceType string
	PrivateIP    string
	PublicIP     string
	Cost         string
	Name         string
}

var (
	awsctrlPath string
	instances   []Instance
	headers     = []string{"ID", "Status", "Type", "PrivateIP", "PublicIP", "Cost", "Name"}
	colWidths   = []unit.Dp{180, 80, 100, 130, 130, 80, 160}

	profileEditor widget.Editor
	fetchBtn      widget.Clickable
	tableList     widget.List
	errMsg        string
	infoMsg       string
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
		w.Option(app.Size(unit.Dp(950), unit.Dp(500)))
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
						if len(instances) == 0 {
							infoMsg = "インスタンスが見つかりませんでした"
						}
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

// layoutTopBar はプロファイル入力欄と「取込」ボタンを描画
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
	totalRows := len(instances) + 1 // +1 ヘッダ行

	return material.List(th, &tableList).Layout(gtx, totalRows, func(gtx layout.Context, index int) layout.Dimensions {
		isHeader := index == 0

		// 行の背景色
		bgColor := color.NRGBA{R: 255, G: 255, B: 255, A: 255}
		if isHeader {
			bgColor = color.NRGBA{R: 220, G: 220, B: 240, A: 255}
		} else if index%2 == 0 {
			bgColor = color.NRGBA{R: 245, G: 245, B: 245, A: 255}
		}

		return drawRowBackground(gtx, bgColor, func(gtx layout.Context) layout.Dimensions {
			return layout.UniformInset(unit.Dp(4)).Layout(gtx, func(gtx layout.Context) layout.Dimensions {
				var cells []string
				if isHeader {
					cells = headers
				} else {
					inst := instances[index-1]
					cells = []string{inst.ID, inst.Status, inst.InstanceType, inst.PrivateIP, inst.PublicIP, inst.Cost, inst.Name}
				}

				children := make([]layout.FlexChild, len(cells))
				for i, cell := range cells {
					cellText := cell
					colW := colWidths[i]
					bold := isHeader
					children[i] = layout.Rigid(func(gtx layout.Context) layout.Dimensions {
						gtx.Constraints.Min.X = gtx.Dp(colW)
						gtx.Constraints.Max.X = gtx.Dp(colW)
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

// loadEnv は .env ファイルから設定を読み込む
func loadEnv(path string) error {
	f, err := os.Open(path)
	if err != nil {
		return err
	}
	defer f.Close()

	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		parts := strings.SplitN(line, "=", 2)
		if len(parts) != 2 {
			continue
		}
		key := strings.TrimSpace(parts[0])
		value := strings.TrimSpace(parts[1])
		if key == "AWSCTRL_PATH" {
			awsctrlPath = value
		}
	}
	return scanner.Err()
}

// executeAwsCtrl は awsctrl コマンドを実行し標準出力を返す
func executeAwsCtrl(profile string) (string, error) {
	cmd := exec.Command(awsctrlPath, "-profile", profile)
	out, err := cmd.CombinedOutput()
	if err != nil {
		return "", fmt.Errorf("%v\n%s", err, string(out))
	}
	return string(out), nil
}

// parseOutput は awsctrl の出力テキストをパースしてインスタンス情報に変換する
func parseOutput(output string) []Instance {
	var result []Instance
	lines := strings.Split(output, "\n")
	for _, line := range lines {
		trimmed := strings.TrimSpace(line)
		if !strings.HasPrefix(trimmed, "ID:") {
			continue
		}
		inst := parseInstanceLine(trimmed)
		if inst.ID != "" {
			result = append(result, inst)
		}
	}
	return result
}

// parseInstanceLine は1行のインスタンス情報をパースする
// 入力形式: "ID: <id>,  <status>, <type>, <privateIP>, <publicIP>  <TagKey>: <TagVal>  ..."
func parseInstanceLine(line string) Instance {
	line = strings.TrimPrefix(line, "ID: ")

	parts := strings.SplitN(line, ", ", 5)
	if len(parts) < 5 {
		return Instance{}
	}

	id := strings.TrimSpace(parts[0])
	status := strings.TrimSpace(parts[1])
	instanceType := strings.TrimSpace(parts[2])
	privateIP := strings.TrimSpace(parts[3])
	lastPart := parts[4]

	publicIP := lastPart
	name := ""

	// publicIP 以降のタグ部分を分離（タグは "  " で区切られている）
	tagIdx := strings.Index(lastPart, "  ")
	if tagIdx >= 0 {
		publicIP = strings.TrimSpace(lastPart[:tagIdx])
		tagStr := lastPart[tagIdx:]
		tagEntries := strings.Split(tagStr, "  ")
		for _, entry := range tagEntries {
			entry = strings.TrimSpace(entry)
			if entry == "" {
				continue
			}
			if kv := strings.SplitN(entry, ": ", 2); len(kv) == 2 {
				if strings.TrimSpace(kv[0]) == "Name" {
					name = strings.TrimSpace(kv[1])
				}
			}
		}
	} else {
		publicIP = strings.TrimSpace(publicIP)
	}

	return Instance{
		ID:           id,
		Status:       status,
		InstanceType: instanceType,
		PrivateIP:    privateIP,
		PublicIP:     publicIP,
		Cost:         "-",
		Name:         name,
	}
}
