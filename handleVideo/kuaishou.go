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

// extractKSAtlasImages - 从快手 HTML 中提取 atlas 图文列表
// 格式: "cdn":["xxx.com","yyy.com"],"list":["/ufile/atlas/xxx_0.jpg",...]
func extractKSAtlasImages(html string) []string {
	regsAtlas := regexp.MustCompile(`"cdn":\[(.*?)\],"list":\[(.*?)\]`)
	matchesAtlas := regsAtlas.FindStringSubmatch(html)
	if len(matchesAtlas) != 3 {
		return nil
	}

	// 解析 cdn 数组
	cdnRe := regexp.MustCompile(`"([^"]+)"`)
	cdnMatches := cdnRe.FindAllStringSubmatch(matchesAtlas[1], -1)
	var cdnArr []string
	for _, cm := range cdnMatches {
		if len(cm) >= 2 {
			cdnArr = append(cdnArr, cm[1])
		}
	}

	// 解析 list 数组
	listMatches := cdnRe.FindAllStringSubmatch(matchesAtlas[2], -1)
	var listArr []string
	for _, lm := range listMatches {
		if len(lm) >= 2 {
			listArr = append(listArr, lm[1])
		}
	}

	if len(cdnArr) == 0 || len(listArr) == 0 {
		return nil
	}

	// 去重: list 中 jpg 和 webp 各一份，只保留 jpg
	var filtered []string
	seen := make(map[string]bool)
	for _, path := range listArr {
		key := path
		// 去掉扩展名做去重
		dotIdx := strings.LastIndex(key, ".")
		if dotIdx > 0 {
			key = key[:dotIdx]
		}
		if seen[key] {
			continue
		}
		seen[key] = true
		// 优先 jpg
		if strings.HasSuffix(path, ".jpg") || !strings.HasSuffix(path, ".webp") {
			filtered = append(filtered, path)
		} else {
			filtered = append(filtered, path)
		}
	}

	baseURL := "https://" + cdnArr[0]
	var urls []string
	for _, path := range filtered {
		if strings.HasPrefix(path, "http") {
			urls = append(urls, path)
		} else {
			urls = append(urls, baseURL+path)
		}
	}
	return urls
}

// KuaiShou 解析 - 支持视频(mp4)和图文(图片列表)
// 返回: 视频URL, 图片URL列表(IMAGE:前缀分隔), 错误
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

	// 方式3: 图文帖子 - atlas list + cdn (新格式)
	atlasImages := extractKSAtlasImages(html)
	if len(atlasImages) > 0 {
		return "IMAGE:" + strings.Join(atlasImages, ","), nil
	}

	// 方式4: 图文帖子 - imageUrls (旧格式，兜底)
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
				return "IMAGE:" + strings.Join(urls, ","), nil
			}
		}
	}

	return "", errors.New("无效地址：未找到视频或图片链接")
}
