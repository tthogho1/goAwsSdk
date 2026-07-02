package main

import (
	"flag"
	"fmt"
	"image/color"
	"os"
	"strings"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"

	"ec2viewer/ui"

	"gioui.org/app"
	"gioui.org/layout"
	"gioui.org/op"
	"gioui.org/unit"
	"gioui.org/widget/material"
)

var (
	awsctrlPath string
	logPath     string
)

func main() {
	// CLI: allow setting logfile path (default: ./app.log)
	flag.StringVar(&logPath, "logfile", "", "log file path (default: ./app.log)")
	flag.Parse()
	if logPath == "" {
		logPath = "./app.log"
	}
	// initialize logger
	f, lg, s, err := initLogger(logPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "cannot initialize logger %s: %v\n", logPath, err)
		os.Exit(1)
	}
	logFile = f
	logger = lg
	sugar = s
	defer func() {
		_ = f.Close()
		_ = logger.Sync()
	}()

	// .env 読み込み
	if err := loadEnv(".env"); err != nil {
		fmt.Fprintf(os.Stderr, ".envファイル読み込みエラー: %v\n", err)
		os.Exit(1)
	}
	if awsctrlPath == "" {
		fmt.Fprintln(os.Stderr, "AWSCTRL_PATH が .env に設定されていません")
		os.Exit(1)
	}

	state := &ui.AppState{Logger: sugar}
	state.ProfileEditor.SetText("default")
	state.ProfileEditor.SingleLine = true
	state.TableList.Axis = layout.Vertical

	go func() {
		w := new(app.Window)
		w.Option(app.Title("EC2 Instances Viewer"))
		w.Option(app.Size(unit.Dp(1020), unit.Dp(500)))
		if err := run(w, state); err != nil {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(1)
		}
		os.Exit(0)
	}()
	app.Main()
}

// run はウィンドウのイベントループ
func run(w *app.Window, state *ui.AppState) error {
	th := material.NewTheme()
	var ops op.Ops

	for {
		switch e := w.Event().(type) {
		case app.DestroyEvent:
			return e.Err
		case app.FrameEvent:
			gtx := app.NewContext(&ops, e)

			// 「取込」ボタン押下処理
			if state.FetchBtn.Clicked(gtx) {
				handleFetch(state)
			}

			// 「実行」ボタン押下処理
			if state.ExecuteBtn.Clicked(gtx) && state.HasStatusChanges() {
				handleExecute(state)
			}

				// update search query each frame; mark dirty when changed
				newQuery := strings.TrimSpace(state.SearchEditor.Text())
				if newQuery != state.SearchQuery {
					state.SearchQuery = newQuery
					state.VisibleDirty = true
				}
				// search is live; editing `SearchEditor` already marks `VisibleDirty`

				// メインレイアウト: 上部バー + メッセージ + テーブル
			layout.Flex{Axis: layout.Vertical}.Layout(gtx,
				layout.Rigid(func(gtx layout.Context) layout.Dimensions {
					return ui.LayoutTopBar(gtx, th, state)
				}),
				layout.Rigid(func(gtx layout.Context) layout.Dimensions {
					if state.ErrMsg != "" {
						return ui.LayoutMessage(gtx, th, state.ErrMsg, color.NRGBA{R: 220, G: 50, B: 50, A: 255})
					}
					if state.InfoMsg != "" {
						return ui.LayoutMessage(gtx, th, state.InfoMsg, color.NRGBA{R: 50, G: 50, B: 200, A: 255})
					}
					return layout.Dimensions{}
				}),
				layout.Rigid(func(gtx layout.Context) layout.Dimensions {
					return ui.LayoutSearchBar(gtx, th, state)
				}),
				layout.Flexed(1, func(gtx layout.Context) layout.Dimensions {
					return ui.LayoutTable(gtx, th, state)
				}),
			)

			e.Frame(gtx.Ops)
		}
	}
}

// initLogger opens the given path for append (creating if missing) and
// returns a zap Logger and SugaredLogger that write to that file.
func initLogger(path string) (*os.File, *zap.Logger, *zap.SugaredLogger, error) {
	f, err := os.OpenFile(path, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0644)
	if err != nil {
		return nil, nil, nil, err
	}
	encCfg := zap.NewProductionEncoderConfig()
	encCfg.EncodeTime = zapcore.ISO8601TimeEncoder
	encoder := zapcore.NewJSONEncoder(encCfg)
	writeSyncer := zapcore.AddSync(f)
	core := zapcore.NewCore(encoder, writeSyncer, zapcore.DebugLevel)
	lg := zap.New(core)
	s := lg.Sugar()
	return f, lg, s, nil
}

