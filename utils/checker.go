// utils.go
package utils

import (
	"regexp"

	"github.com/aws/aws-sdk-go/service/ec2"
)

// Tag名に指定されたPatternに一致するものがあればTrueを返す
func IsPatternMatch(tags []*ec2.Tag, pattern *string) bool {
	if pattern == nil {
		return true
	}

	for _, tag := range tags {
		matched, _ := regexp.MatchString(*pattern, *tag.Value)
		if matched {
			// パターンにマッチした場合の処理
			return true
		}
	}

	return false
}

// Tag名に指定されたPatternに一致するものがあればTrueを返す
func IsServicePatternMatch(serviceName *string, pattern *string) bool {
	if pattern == nil {
		return true
	}
	matched, _ := regexp.MatchString(*pattern, *serviceName)
	if matched {
		// パターンにマッチした場合の処理
		return true
	}

	return false
}
