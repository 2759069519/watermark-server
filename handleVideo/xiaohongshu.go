package handleVideo

import (
	"encoding/json"
	"errors"
	"fmt"
	"os/exec"
	"strings"
)

type xhsResult struct {
	Code int    `json:"code"`
	URL  string `json:"url"`
	Msg  string `json:"msg"`
}

// 小红书解析 - 通过调用 Python 脚本实现（需要 xhshow 签名库）
func XiaoHongShu(rUrl, ua string) (string, error) {
	// 调用同目录下的 Python 脚本
	cmd := exec.Command("python3", "/opt/watermark-server/xhs_parse.py", rUrl)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return "", fmt.Errorf("脚本执行失败: %v", err)
	}

	var result xhsResult
	if err := json.Unmarshal(output, &result); err != nil {
		// 尝试清理输出
		clean := strings.TrimSpace(string(output))
		if err := json.Unmarshal([]byte(clean), &result); err != nil {
			return "", fmt.Errorf("解析脚本输出失败: %s", clean)
		}
	}

	if result.Code != 200 {
		return "", errors.New(result.Msg)
	}

	if result.URL == "" {
		return "", errors.New("未获取到视频链接")
	}

	return result.URL, nil
}
