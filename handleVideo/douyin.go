package handleVideo

import (
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"regexp"
	"strings"
	"time"
)

const (
	douyinUA = "Mozilla/5.0 (iPhone; CPU iPhone OS 16_0 like Mac OS X) AppleWebKit/605.1.15 (KHTML, like Gecko) Version/16.0 Mobile/15E148 Safari/604.1"
)

// DouYin 解析抖音视频和图文，返回统一结构
func DouYin(rUrl string) (*VideoInfo, error) {
	// Step 1: 跟随短链
	realUrl, err := getFinalURL(rUrl, douyinUA)
	if err != nil {
		return nil, fmt.Errorf("短链跳转失败: %v", err)
	}
	log.Println("抖音真实URL:", realUrl)

	info := &VideoInfo{
		Platform: "douyin",
		ShortURL: rUrl,
	}

	if strings.Contains(realUrl, "/note/") {
		return douyinNote(realUrl, info)
	}
	return douyinVideo(realUrl, info)
}

// douyinVideo 解析抖音视频
func douyinVideo(realUrl string, info *VideoInfo) (*VideoInfo, error) {
	// 提取视频ID
	re := regexp.MustCompile(`/video/(\d+)`)
	matches := re.FindStringSubmatch(realUrl)
	if len(matches) < 2 {
		re2 := regexp.MustCompile(`(\d{15,25})`)
		matches2 := re2.FindAllString(realUrl, -1)
		if len(matches2) == 0 {
			return nil, errors.New("无效链接：无法提取视频ID")
		}
		matches = []string{"", matches2[0]}
	}
	videoId := matches[1]
	log.Println("抖音视频ID:", videoId)

	// 请求分享页面
	sharePageURL := fmt.Sprintf("https://www.iesdouyin.com/share/video/%s/", videoId)
	client := &http.Client{Timeout: 15 * time.Second}
	req, err := http.NewRequest("GET", sharePageURL, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("User-Agent", douyinUA)
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("分享页请求失败: %v", err)
	}
	defer resp.Body.Close()
	body, _ := ioutil.ReadAll(resp.Body)
	html := string(body)

	// 提取 video_id
	vidRe := regexp.MustCompile(`video_id=([a-zA-Z0-9]+)`)
	vidMatches := vidRe.FindStringSubmatch(html)
	if len(vidMatches) < 2 {
		return nil, errors.New("无法从分享页提取video_id")
	}
	shortVideoId := vidMatches[1]

	// 提取文案（desc 优先，title 兜底）
	if desc := extractDouyinField(html, "desc"); desc != "" {
		info.Title = desc
	} else {
		info.Title = extractDouyinField(html, "title")
	}

	// 提取话题标签
	info.Topics = extractDouyinTopics(html)

	// 提取封面
	info.Cover = extractDouyinField(html, "cover")

	// 获取无水印视频
	playURL := fmt.Sprintf("https://api-play.amemv.com/aweme/v1/play/?video_id=%s&line=0&ratio=720p&watermark=0&media_type=4&vr_type=0&improve_bitrate=0&logo_name=aweme", shortVideoId)
	finalURL, err := getFinalURL(playURL, douyinUA)
	if err != nil {
		return nil, fmt.Errorf("播放API请求失败: %v", err)
	}
	if !strings.Contains(finalURL, "douyinvod.com") && !strings.Contains(finalURL, "douyinstatic.com") {
		return nil, errors.New("播放API未返回有效视频地址")
	}
	info.Video = finalURL

	log.Printf("抖音视频: title=%s topics=%v", info.Title, info.Topics)
	return info, nil
}

// douyinNote 解析抖音图文
func douyinNote(rUrl string, info *VideoInfo) (*VideoInfo, error) {
	log.Println("抖音图文:", rUrl)

	client := &http.Client{Timeout: 15 * time.Second}
	req, err := http.NewRequest("GET", rUrl, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("User-Agent", douyinUA)
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("请求失败: %v", err)
	}
	defer resp.Body.Close()
	body, _ := ioutil.ReadAll(resp.Body)
	html := string(body)

	// 提取图片
	images := extractDouyinNoteImages(html)
	if len(images) == 0 {
		return nil, errors.New("未找到图片")
	}
	info.Images = images

	// 提取文案
	if desc := extractDouyinField(html, "desc"); desc != "" {
		info.Title = desc
	} else {
		info.Title = extractDouyinField(html, "title")
	}

	// 提取话题
	info.Topics = extractDouyinTopics(html)

	// 提取封面（第一张图作为封面）
	if len(images) > 0 {
		info.Cover = images[0]
	}

	log.Printf("抖音图文: %d张图片 title=%s", len(images), info.Title)
	return info, nil
}

// extractDouyinField 从HTML中提取指定字段的值
func extractDouyinField(html, field string) string {
	// 匹配 "field":"value" 格式
	re := regexp.MustCompile(`"` + field + `"\s*:\s*"([^"]*)"`)
	m := re.FindStringSubmatch(html)
	if len(m) >= 2 {
		val := m[1]
		// 解码 unicode 转义
		val = strings.ReplaceAll(val, `\u0026`, "&")
		val = strings.ReplaceAll(val, `\u002F`, "/")
		val = strings.ReplaceAll(val, `\n`, " ")
		return strings.TrimSpace(val)
	}
	return ""
}

// extractDouyinTopics 从HTML中提取话题标签
func extractDouyinTopics(html string) []string {
	var topics []string
	seen := make(map[string]bool)

	// 匹配 hashtags: "cha_name":"xxx" 或 "text_extra":[{"hashtag_name":"xxx"}]
	re := regexp.MustCompile(`"hashtag_name"\s*:\s*"([^"]+)"`)
	matches := re.FindAllStringSubmatch(html, -1)
	for _, m := range matches {
		if len(m) >= 2 {
			t := strings.TrimSpace(m[1])
			if t != "" && !seen[t] {
				seen[t] = true
				topics = append(topics, "#"+t)
			}
		}
	}

	// 备用: cha_name
	if len(topics) == 0 {
		re2 := regexp.MustCompile(`"cha_name"\s*:\s*"([^"]+)"`)
		matches2 := re2.FindAllStringSubmatch(html, -1)
		for _, m := range matches2 {
			if len(m) >= 2 {
				t := strings.TrimSpace(m[1])
				if t != "" && !seen[t] {
					seen[t] = true
					topics = append(topics, "#"+t)
				}
			}
		}
	}

	return topics
}

// getFinalURL 跟随重定向获取最终URL
func getFinalURL(rUrl string, ua string) (string, error) {
	client := &http.Client{
		Timeout: 15 * time.Second,
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			return nil
		},
	}
	req, err := http.NewRequest("GET", rUrl, nil)
	if err != nil {
		return "", err
	}
	req.Header.Set("User-Agent", ua)
	resp, err := client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	return resp.Request.URL.String(), nil
}

// extractDouyinNoteImages 从抖音HTML中提取无水印图片
func extractDouyinNoteImages(html string) []string {
	html = strings.ReplaceAll(html, `\u002F`, "/")

	var images []string
	seen := make(map[string]bool)

	urlRe := regexp.MustCompile(`https?://p[0-9]+-sign\.douyinpic\.com/[^\s"<>]+`)
	allURLs := urlRe.FindAllString(html, -1)

	uriRe := regexp.MustCompile(`(tos-cn-[^\s?~]+)`)

	// 优先 noop/lqen-new
	for _, url := range allURLs {
		if strings.Contains(url, "-water") {
			continue
		}
		if !strings.Contains(url, "noop") && !strings.Contains(url, "lqen-new") {
			continue
		}
		uriMatch := uriRe.FindString(url)
		if uriMatch == "" || seen[uriMatch] {
			continue
		}
		seen[uriMatch] = true
		images = append(images, url)
	}

	// 兜底
	if len(images) == 0 {
		for _, url := range allURLs {
			if strings.Contains(url, "-water") {
				continue
			}
			uriMatch := uriRe.FindString(url)
			if uriMatch == "" || seen[uriMatch] {
				continue
			}
			seen[uriMatch] = true
			images = append(images, url)
		}
	}

	return images
}
