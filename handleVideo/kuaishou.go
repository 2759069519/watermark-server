package handleVideo

import (
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"regexp"
	"strings"
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

// KuaiShou 解析 - 支持视频(mp4)和图文(图片列表)
// 返回: 视频URL, 图片URL列表(逗号分隔), 错误
func KuaiShou(rUrl, ua string) (string, error) {
	body, err := getKSRealityUrl(rUrl, ua)
	if err != nil {
		return "", err
	}
	html := string(body)

	// 方式1: 匹配视频 mp4 URL
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

	// 方式3: 图文帖子 - 提取 imageUrls
	regs3 := regexp.MustCompile(`"imageUrls":\[(.*?)\]`)
	matches3 := regs3.FindStringSubmatch(html)
	if len(matches3) == 2 {
		urlRegs := regexp.MustCompile(`"url":"(https?://[^"]+)"`)
		urlMatches := urlRegs.FindAllStringSubmatch(matches3[1], -1)
		if len(urlMatches) > 0 {
			var urls []string
			for _, m := range urlMatches {
				if len(m) >= 2 {
					urls = append(urls, m[1])
				}
			}
			if len(urls) > 0 {
				// 用特殊前缀标记图文，让 controller 识别
				return "IMAGE:" + strings.Join(urls, ","), nil
			}
		}
	}

	return "", errors.New("无效地址：未找到视频或图片链接")
}
