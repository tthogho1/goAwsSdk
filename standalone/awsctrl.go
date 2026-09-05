package main

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/joho/godotenv"
)

// envFilePath determines the search path for the .env file.
// It prefers .env in the current directory if present, otherwise it
// looks in the directory where the executable itself resides
// (resolving symlinks in case the app was launched via a symlink).
func envFilePath() string {
	if _, err := os.Stat(".env"); err == nil {
		return ".env"
	}

	exePath, err := os.Executable()
	if err != nil {
		return ".env"
	}
	if resolved, err := filepath.EvalSymlinks(exePath); err == nil {
		exePath = resolved
	}
	return filepath.Join(filepath.Dir(exePath), ".env")
}

// loadEnv loads settings from the .env file
func loadEnv(path string) error {
	if err := godotenv.Load(path); err != nil {
		return err
	}
	awsctrlPath = os.Getenv("AWSCTRL_PATH")
	return nil
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
