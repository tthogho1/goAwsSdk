package main

import (
	"flag"
	"fmt"
	"image/color"
	"os"
	"path/filepath"
	"strings"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"

	"ec2viewer/ui"

	"gioui.org/app"
	"gioui.org/layout"
	"gioui.org/op"
	"gioui.org/unit"
	"gioui.org/widget"
	"gioui.org/widget/material"
)

var (
	awsctrlPath string
	logPath     string
)

func main() {
	// CLI: allow setting logfile path (default: <exe dir>/app.log)
	flag.StringVar(&logPath, "logfile", "", "log file path (default: app.log next to the executable)")
	flag.Parse()
	if logPath == "" {
		logPath = filepath.Join(exeDir(), "app.log")
	}
	// initialize logger; if that fails (e.g. read-only location), fall back to a temp file
	f, lg, s, err := initLogger(logPath)
	if err != nil {
		if tmpPath := filepath.Join(os.TempDir(), "ec2viewer_app.log"); tmpPath != logPath {
			if f2, lg2, s2, err2 := initLogger(tmpPath); err2 == nil {
				f, lg, s, err = f2, lg2, s2, nil
			}
		}
	}
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

	// Load .env (prefer the directory next to the executable)
	if err := loadEnv(envFilePath()); err != nil {
		fmt.Fprintf(os.Stderr, "Failed to load .env file: %v\n", err)
		os.Exit(1)
	}
	if awsctrlPath == "" {
		fmt.Fprintln(os.Stderr, "AWSCTRL_PATH is not set in .env")
		os.Exit(1)
	}

	state := &ui.AppState{Logger: sugar}
	state.Profiles = loadAwsProfiles()
	state.ProfileMenuItems = make([]widget.Clickable, len(state.Profiles))
	if len(state.Profiles) > 0 {
		state.SelectedProfile = state.Profiles[0]
	}
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

// run is the window event loop
func run(w *app.Window, state *ui.AppState) error {
	th := material.NewTheme()
	var ops op.Ops

	for {
		switch e := w.Event().(type) {
		case app.DestroyEvent:
			return e.Err
		case app.FrameEvent:
			gtx := app.NewContext(&ops, e)

			// Handle "Fetch" button press
			if state.FetchBtn.Clicked(gtx) {
				handleFetch(state)
			}

			// Handle "Execute" button press
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

			// Main layout: top bar + message + table
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
				layout.Rigid(func(gtx layout.Context) layout.Dimensions {
					return ui.LayoutFooter(gtx, th, state)
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
