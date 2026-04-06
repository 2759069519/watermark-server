package handleVideo

import (
	"errors"
	"io/ioutil"
	"net/http"
	"regexp"
)

func requestKG3(rUrl, ua string) ([]byte, error) {
	client := &http.Client{}

	request, err := http.NewRequest("GET", rUrl, nil)
	if err != nil {
		return nil, err
	}
	request.Header.Set("Content-Type", "application/json")
	request.Header.Set("user-agent", ua)
	do, err := client.Do(request)
	if err != nil {
		return nil, err
	}
	defer do.Body.Close()
	body, err := ioutil.ReadAll(do.Body)
	if err != nil {
		return nil, err
	}

	return body, nil
}

// 全民K歌解析 - 修复：正则支持末尾无&的URL，兼容 node.kg.qq.com 域名
func QuanMingKGe(sUrl, ua string) (string, error) {
	// 修复：使用 [^&]+ 匹配，支持末尾无&的情况
	s := regexp.MustCompile(`s=([^&\s]+)`).FindStringSubmatch(sUrl)
	if len(s) != 2 {
		return "", errors.New("无效的地址：无法提取s参数")
	}
	body, err := requestKG3("https://kg.qq.com/node/play?s="+s[1], ua)
	if err != nil {
		return "", err
	}
	regs := regexp.MustCompile(`playurl_video":"(.*?)"`).FindStringSubmatch(string(body))
	if len(regs) != 2 || regs[1] == "" {
		return "", errors.New("无效的地址：未找到播放链接（可能视频已失效）")
	}
	return regs[1], nil
}
