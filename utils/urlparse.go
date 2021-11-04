package utils

import "regexp"

var (
	_shareCodePart1 = regexp.MustCompile(`https://cloud.189.cn/t/(\w+)`)
	_shareCodePart2 = regexp.MustCompile(`https://cloud.189.cn/web/share?code=(\w+)`)
)

func ParseShareCode(url string) string {
	var matched []string
	if matched = _shareCodePart1.FindStringSubmatch(url); len(matched) > 1 {
		return matched[1]
	}
	if matched = _shareCodePart2.FindStringSubmatch(url); len(matched) > 1 {
		return matched[1]
	}
	return ""
}
