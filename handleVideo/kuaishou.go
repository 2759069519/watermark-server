package handleVideo

import (
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"regexp"
)

func getKSRealityUrl(rUrl, ua string) ([]byte, error) {
	client := &http.Client{}
	req, err := http.NewRequest("GET", rUrl, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("User-Agent", ua)
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	fmt.Println(resp.Request.URL)
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	return body, nil
}

// 快手解析 - 修复：适配2026年HTML结构，从mainMvUrls提取mp4链接
func KuaiShou(rUrl, ua string) (string, error) {
	body, err := getKSRealityUrl(rUrl, ua)
	if err != nil {
		return "", err
	}
	html := string(body)

	// 方式1: 匹配 mainMvUrls 中的 mp4 URL
	regs := regexp.MustCompile(`"url":"(https?://[^"]+\.mp4[^"]*)"`)
	matches := regs.FindAllStringSubmatch(html, -1)
	if len(matches) > 0 && len(matches[0]) >= 2 {
		return matches[0][1], nil
	}

	// 方式2: 兼容旧格式 srcNoMark
	regs2 := regexp.MustCompile(`srcNoMark":"(.*?)"`)
	matches2 := regs2.FindStringSubmatch(html)
	if len(matches2) == 2 {
		return matches2[1], nil
	}

	return "", errors.New("无效地址：未找到视频链接")
}
