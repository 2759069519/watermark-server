package handleVideo

import (
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"regexp"
	"strings"
)

const kuaishouUA = "Mozilla/5.0 (iPhone; CPU iPhone OS 16_0 like Mac OS X) AppleWebKit/605.1.15 (KHTML, like Gecko) Version/16.0 Mobile/15E148 Safari/604.1"

// KuaiShou 解析快手视频和图文
func KuaiShou(rUrl string) (*VideoInfo, error) {
	client := &http.Client{}
	req, err := http.NewRequest("GET", rUrl, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("User-Agent", kuaishouUA)
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	html := string(body)
	realURL := resp.Request.URL.String()
	log.Println("快手真实URL:", realURL)

	info := &VideoInfo{
		Platform: "kuaishou",
		ShortURL: rUrl,
	}

	// 提取文案
	info.Title = extractKSField(html, "caption")

	// 提取话题
	info.Topics = extractKSTopics(html)

	// 提取封面
	info.Cover = extractKSCover(html)

	// 方式1: 视频 mp4
	reMp4 := regexp.MustCompile(`"url":"(https?://[^"]+\.mp4[^"]*)"`)
	matches := reMp4.FindAllStringSubmatch(html, -1)
	if len(matches) > 0 && len(matches[0]) >= 2 {
		info.Video = matches[0][1]
		return info, nil
	}

	// 方式2: srcNoMark
	reSrc := regexp.MustCompile(`srcNoMark":"(.*?)"`)
	matches2 := reSrc.FindStringSubmatch(html)
	if len(matches2) == 2 {
		info.Video = matches2[1]
		return info, nil
	}

	// 方式3: 图文 - atlas
	atlasImages := extractKSAtlasImages(html)
	if len(atlasImages) > 0 {
		info.Images = atlasImages
		return info, nil
	}

	// 方式4: 图文 - imageUrls
	re3 := regexp.MustCompile(`"imageUrls":\[(.*?)\]`)
	matches3 := re3.FindStringSubmatch(html)
	if len(matches3) == 2 {
		urlRe := regexp.MustCompile(`"url":"(https?://[^"]+)"`)
		urlMatches := urlRe.FindAllStringSubmatch(matches3[1], -1)
		var urls []string
		for _, m := range urlMatches {
			if len(m) >= 2 {
				urls = append(urls, m[1])
			}
		}
		if len(urls) > 0 {
			info.Images = urls
			return info, nil
		}
	}

	return nil, errors.New("无效地址：未找到视频或图片链接")
}

// extractKSAtlasImages 提取快手 atlas 图文列表
func extractKSAtlasImages(html string) []string {
	re := regexp.MustCompile(`"cdn":\[(.*?)\],"list":\[(.*?)\]`)
	matches := re.FindStringSubmatch(html)
	if len(matches) != 3 {
		return nil
	}

	cdnRe := regexp.MustCompile(`"([^"]+)"`)
	cdnMatches := cdnRe.FindAllStringSubmatch(matches[1], -1)
	var cdnArr []string
	for _, cm := range cdnMatches {
		if len(cm) >= 2 {
			cdnArr = append(cdnArr, cm[1])
		}
	}

	listMatches := cdnRe.FindAllStringSubmatch(matches[2], -1)
	var listArr []string
	for _, lm := range listMatches {
		if len(lm) >= 2 {
			listArr = append(listArr, lm[1])
		}
	}

	if len(cdnArr) == 0 || len(listArr) == 0 {
		return nil
	}

	seen := make(map[string]bool)
	var urls []string
	baseURL := "https://" + cdnArr[0]
	for _, path := range listArr {
		key := path
		if dotIdx := strings.LastIndex(key, "."); dotIdx > 0 {
			key = key[:dotIdx]
		}
		if seen[key] {
			continue
		}
		seen[key] = true
		if strings.HasPrefix(path, "http") {
			urls = append(urls, path)
		} else {
			urls = append(urls, baseURL+path)
		}
	}
	return urls
}

// extractKSField 从快手HTML中提取字段
func extractKSField(html, field string) string {
	re := regexp.MustCompile(`"` + field + `"\s*:\s*"([^"]*)"`)
	m := re.FindStringSubmatch(html)
	if len(m) >= 2 {
		val := m[1]
		val = strings.ReplaceAll(val, `\u0026`, "&")
		val = strings.ReplaceAll(val, `\u002F`, "/")
		return strings.TrimSpace(val)
	}
	return ""
}

// extractKSTopics 提取快手话题标签
func extractKSTopics(html string) []string {
	var topics []string
	seen := make(map[string]bool)

	// 快手话题格式: "atlasTitle":"#话题#" 或 "topic":"xxx"
	re := regexp.MustCompile(`"topic"\s*:\s*"([^"]+)"`)
	matches := re.FindAllStringSubmatch(html, -1)
	for _, m := range matches {
		if len(m) >= 2 {
			t := strings.TrimSpace(m[1])
			if t != "" && !seen[t] {
				seen[t] = true
				if !strings.HasPrefix(t, "#") {
					t = "#" + t
				}
				topics = append(topics, t)
			}
		}
	}

	// 备用: tagName
	if len(topics) == 0 {
		re2 := regexp.MustCompile(`"tagName"\s*:\s*"([^"]+)"`)
		matches2 := re2.FindAllStringSubmatch(html, -1)
		for _, m := range matches2 {
			if len(m) >= 2 {
				t := strings.TrimSpace(m[1])
				if t != "" && !seen[t] {
					seen[t] = true
					if !strings.HasPrefix(t, "#") {
						t = "#" + t
					}
					topics = append(topics, t)
				}
			}
		}
	}

	return topics
}

// extractKSCover 提取快手封面
func extractKSCover(html string) string {
	// 封面格式: "coverUrl":"xxx" 或 "photoUrl":"xxx" 或 "thumbnailUrl":"xxx"
	fields := []string{"coverUrl", "photoUrl", "thumbnailUrl", "cover", "poster"}
	for _, f := range fields {
		re := regexp.MustCompile(`"` + f + `"\s*:\s*"(https?://[^"]+)"`)
		m := re.FindStringSubmatch(html)
		if len(m) >= 2 {
			url := strings.ReplaceAll(m[1], `\u002F`, "/")
			if url != "" {
				return url
			}
		}
	}
	return ""
}
