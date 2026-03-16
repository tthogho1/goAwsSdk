package main

import (
	"fmt"
	"os"
	"os/exec"

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

// executeAwsCtrl は awsctrl コマンドを実行し標準出力を返す
func executeAwsCtrl(profile string) (string, error) {
	cmd := exec.Command(awsctrlPath, "-profile", profile)
	out, err := cmd.CombinedOutput()
	if err != nil {
		return "", fmt.Errorf("%v\n%s", err, string(out))
	}
	return string(out), nil
}

// executeAwsCtrlAction は指定インスタンスに対してup/downコマンドを実行する
func executeAwsCtrlAction(profile, action, instanceID string) error {
	cmd := exec.Command(awsctrlPath, "-profile", profile, "-c", action, "-t", "EC2", "-i", instanceID)
	out, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("%v\n%s", err, string(out))
	}
	return nil
}
