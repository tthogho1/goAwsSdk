package main

import (
	"fmt"
	"strings"

	"ec2viewer/ui"
)

// refreshInstances は awsctrl を実行してインスタンス一覧を取得・更新する
func refreshInstances(state *ui.AppState, profile string) error {
	output, err := executeAwsCtrl(profile)
	if err != nil {
		return err
	}
	state.Instances = parseOutput(output)
	state.InitStatusSlices()
	return nil
}

// handleFetch は「取込」ボタン押下時の処理を担う
func handleFetch(state *ui.AppState) {
	profile := strings.TrimSpace(state.ProfileEditor.Text())
	if profile == "" {
		state.ErrMsg = "プロファイルを入力してください"
		state.InfoMsg = ""
		return
	}
	state.ErrMsg = ""
	state.InfoMsg = ""
	if err := refreshInstances(state, profile); err != nil {
		state.ErrMsg = fmt.Sprintf("awsctrl 実行エラー: %v", err)
		return
	}
	if len(state.Instances) == 0 {
		state.InfoMsg = "インスタンスが見つかりませんでした"
	}
}

// handleExecute は「実行」ボタン押下時の処理を担う
func handleExecute(state *ui.AppState) {
	profile := strings.TrimSpace(state.ProfileEditor.Text())
	var errs []string
	for i := range state.Instances {
		if state.DesiredStatus[i] == state.OriginalStatus[i] || state.OriginalStatus[i] == "-" {
			continue
		}
		action := "up"
		if state.DesiredStatus[i] == "off" {
			action = "down"
		}
		if err := executeAwsCtrlAction(profile, action, state.Instances[i].ID); err != nil {
			errs = append(errs, fmt.Sprintf("%s: %v", state.Instances[i].ID, err))
		}
	}
	if len(errs) > 0 {
		state.ErrMsg = "実行エラー: " + strings.Join(errs, "; ")
		state.InfoMsg = ""
		return
	}
	// 成功時: 再取得してステータス更新
	if err := refreshInstances(state, profile); err != nil {
		state.ErrMsg = fmt.Sprintf("再取得エラー: %v", err)
		return
	}
	state.InfoMsg = "実行完了"
	state.ErrMsg = ""
}
