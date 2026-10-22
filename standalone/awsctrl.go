package main

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"

	"github.com/joho/godotenv"
)

// exeDir returns the directory containing the running executable,
// resolving symlinks in case the app was launched via a symlink
// (e.g. a Desktop shortcut). Falls back to "." on error.
func exeDir() string {
	exePath, err := os.Executable()
	if err != nil {
		return "."
	}
	if resolved, err := filepath.EvalSymlinks(exePath); err == nil {
		exePath = resolved
	}
	return filepath.Dir(exePath)
}

// envFilePath determines the search path for the .env file.
// It prefers .env in the current directory if present, otherwise it
// looks in the directory where the executable itself resides
// (resolving symlinks in case the app was launched via a symlink).
func envFilePath() string {
	if _, err := os.Stat(".env"); err == nil {
		return ".env"
	}
	return filepath.Join(exeDir(), ".env")
}

// loadEnv loads settings from the .env file
func loadEnv(path string) error {
	if err := godotenv.Load(path); err != nil {
		return err
	}
	awsctrlPath = os.Getenv("AWSCTRL_PATH")
	return nil
}

// checkConfig loads .env and verifies AWSCTRL_PATH points to an executable.
// Errors include setup instructions since they are shown directly to the user.
func checkConfig() error {
	path := envFilePath()
	if _, err := os.Stat(path); err != nil {
		return fmt.Errorf("Configuration file .env was not found.\n\n%s", envHelp())
	}
	if err := loadEnv(path); err != nil {
		return fmt.Errorf("Failed to load %s: %v\n\n%s", path, err, envHelp())
	}
	if awsctrlPath == "" {
		return fmt.Errorf("AWSCTRL_PATH is not set in %s.\n\n%s", path, envHelp())
	}
	if _, err := exec.LookPath(awsctrlPath); err != nil {
		return fmt.Errorf("awsctrl executable not found:\n  AWSCTRL_PATH=%s\n\n%s", awsctrlPath, envHelp())
	}
	return nil
}

// envHelp describes where .env goes and what it must contain.
func envHelp() string {
	example := "/path/to/awsctrl"
	if runtime.GOOS == "windows" {
		example = `C:\path\to\awsctrl.exe`
	}
	return fmt.Sprintf("Create a file named .env in:\n  %s\n\n"+
		"with the following content:\n  AWSCTRL_PATH=%s\n\n"+
		"Parameters:\n  AWSCTRL_PATH (required): full path to the awsctrl executable.",
		exeDir(), example)
}

func executeAwsCtrl(profile string) (string, error) {
	cmd := exec.Command(awsctrlPath, "-profile", profile)
	setHideWindow(cmd)

	out, err := cmd.CombinedOutput()
	if err != nil {
		return "", fmt.Errorf("%v\n%s", err, string(out))
	}
	return string(out), nil
}

func executeAwsCtrlAction(profile, action, instanceID string) error {
	cmd := exec.Command(awsctrlPath, "-profile", profile, "-c", action, "-t", "EC2", "-i", instanceID)
	setHideWindow(cmd)

	out, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("%v\n%s", err, string(out))
	}
	return nil
}

// executeAwsCtrlCost runs the "cost" subcommand of awsctrl for the given profile and date range
func executeAwsCtrlCost(profile, start, end string) (string, error) {
	cmd := exec.Command(awsctrlPath, "-profile", profile, "-c", "cost", "-start", start, "-end", end)
	setHideWindow(cmd)

	out, err := cmd.CombinedOutput()
	if err != nil {
		return "", fmt.Errorf("%v\n%s", err, string(out))
	}
	return string(out), nil
}

// executeAwsCtrlDetail runs the "detail" subcommand of awsctrl to fetch
// VPC/Subnet/SecurityGroup/Tag information for a single EC2 instance.
func executeAwsCtrlDetail(profile, instanceID string) (string, error) {
	cmd := exec.Command(awsctrlPath, "-profile", profile, "-c", "detail", "-t", "EC2", "-i", instanceID)
	setHideWindow(cmd)

	out, err := cmd.CombinedOutput()
	if err != nil {
		return "", fmt.Errorf("%v\n%s", err, string(out))
	}
	return string(out), nil
}
