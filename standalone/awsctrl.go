package main

import (
	"fmt"
	"os"
	"os/exec"
	"syscall"

	"github.com/joho/godotenv"
)

// loadEnv は .env ファイルから設定を読み込む
func loadEnv(path string) error {
	if err := godotenv.Load(path); err != nil {
		return err
	}
	awsctrlPath = os.Getenv("AWSCTRL_PATH")
	return nil
}

func executeAwsCtrl(profile string) (string, error) {
	cmd := exec.Command(awsctrlPath, "-profile", profile)
	cmd.SysProcAttr = &syscall.SysProcAttr{
		HideWindow: true,
	}

	out, err := cmd.CombinedOutput()
	if err != nil {
		return "", fmt.Errorf("%v\n%s", err, string(out))
	}
	return string(out), nil
}

func executeAwsCtrlAction(profile, action, instanceID string) error {
	cmd := exec.Command(awsctrlPath, "-profile", profile, "-c", action, "-t", "EC2", "-i", instanceID)
	cmd.SysProcAttr = &syscall.SysProcAttr{
		HideWindow: true,
	}

	out, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("%v\n%s", err, string(out))
	}
	return nil
}
