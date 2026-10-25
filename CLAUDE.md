# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Repository layout: several independent Go modules

This repo holds **four separate Go modules**. Run Go commands from inside the module you are changing; the root module does not include the others.

| Directory | Module | What it is | Go |
|---|---|---|---|
| `.` (root) | `awsctrl` | CLI for controlling AWS resources (aws-sdk-go **v1**) | 1.22 |
| `standalone/` | `ec2viewer` | Desktop GUI (Gio) that drives the `awsctrl` binary | 1.23 (toolchain 1.24) |
| `sample-s3-list/` | `s3-list` | Standalone S3 listing sample | 1.21 |
| `sample-s3-upload/` | — | Standalone S3 folder-upload sample | — |

`scripts/` belongs to the root module: one-off S3 maintenance programs, each run with `go run <file>.go` (see `scripts/README.md`). `delete_s3_object.go_` is renamed on purpose so it stays out of the build.

## Commands

There are no tests in the repo. CI (`.github/workflows/`) only builds each module for linux/amd64 and windows/amd64.

```bash
# Root CLI
go build -o awsctrl .
go run . -profile default -c describe -t EC2        # same args as .vscode/launch.json
go run . -h help                                    # usage

# GUI (Linux needs CGO_ENABLED=1 plus X11/Wayland/GL dev libs; see the CI workflow)
cd standalone && go run .
cd standalone && go build -o ec2viewer .
GOOS=windows go build -ldflags="-H=windowsgui" -o ec2viewer.exe .   # no console window on Windows

# Samples
cd sample-s3-list && go build ./...
```

`go vet ./...` / `gofmt` are the only lint tools available; nothing else is configured.

## Architecture

### `awsctrl` (root)
- `AwsCtrl.go` holds all flag parsing (`OptionParse`) and a `switch` on `-c` (command) and then `-t` (target) that builds the right aws-sdk-go v1 service client and calls into `utils/`.
- `utils/` holds one file per area (`describe.go`, `updown.go`, `create.go`, `exec.go` via SSM, `s3.go`, `cost.go`, `cfn.go`). The functions print results straight to stdout rather than returning them.
- `types/common.go` defines `NetWorkIF`, a `flag.Value` that parses the `-network-interfaces` JSON for `create`.
- The default region is `ap-northeast-1`. Credentials come from shared-config profiles (`-profile`).
- The README's usage section is partly out of date. The `Usage()` function in `AwsCtrl.go` lists the real commands.

### `ec2viewer` (standalone/) — talks to awsctrl through its stdout
The GUI does **not** call AWS itself. It runs the `awsctrl` executable (path from `AWSCTRL_PATH` in a `.env` next to the executable or in the CWD) and parses its text output:

- `awsctrl.go`: `executeAwsCtrl*` functions run `awsctrl -profile <p> [-c up|down|cost|detail ...]`.
- `parser.go`: parses the plain-text formats printed by `utils/describe.go` (`ID: <id>,  <status>, <type>, <privateIP>, <publicIP>  Key: Val ...`, plus `InstanceId:/VpcId:/SubnetId:/SecurityGroup:/Tag:` lines for `detail`) and by `utils/cost.go` (`Service: X, Amount: Y` / `Total: N`), and by `utils/cfn.go` (`Resource:` / `Status:` / `Warning:` / `Template:` lines; the YAML itself goes to the `-o` file, not stdout).
- **If you change a `fmt.Printf` format in `utils/`, update `standalone/parser.go` to match.** Otherwise the GUI breaks without any error.
- CFn generation is the one long-running action: `startCfn` runs awsctrl in a goroutine and sends the result on a channel that the event loop in `main.go` drains each frame (`applyCfnResult`). The goroutine must not touch `AppState`.
- `controller.go` holds the button handlers (Fetch / Execute up-down diff / Cost / Detail), which change `ui.AppState`. `ui/layout.go` and `ui/state.go` hold the Gio immediate-mode layout and the state. `model/` holds the plain data structs.
- `awsprofiles.go` fills the profile picker from `~/.aws/credentials` and `~/.aws/config`.
- Platform-specific code uses `_windows.go` / `_other.go` build-tag pairs: hiding the child-process console window, and showing startup errors in a MessageBox (the Windows build has no console). Startup failures go through `fatalStartup` so the user sees them instead of the app exiting silently.
- Logging uses zap and writes to `app.log` next to the executable (`-logfile` overrides it; falls back to the temp dir).

## Conventions
- Comments and some messages in the older code are in Japanese. Newer code in `standalone/` uses English.
- Built binaries, `.env`, `*.bat`, `*.log`, and `standalone/EC2Viewer.app/` are gitignored and machine-specific. Don't commit them.
