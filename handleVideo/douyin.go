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

// 获取重定向后的最终URL
func getFinalURL(rUrl string, ua string) (string, error) {
	client := &http.Client{
		Timeout: 15 * time.Second,
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			return nil // 跟随重定向
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

// 提取抖音图文图片
func extractDouyinNoteImages(html string) []string {
	// 将 \u002F 转为 /
	html = strings.ReplaceAll(html, `\u002F`, "/")

	// 方法1: 匹配 "images" 数组中的 "url_list" - 提取所有 douyinpic URL
	// 格式: "images":[{"uri":"...","url_list":["https://p3-sign.douyinpic.com/...", ...]}]
	var images []string
	seen := make(map[string]bool)

	// 用正则提取所有 douyinpic URL
	urlRe := regexp.MustCompile(`https?://p[0-9]+-sign\.douyinpic\.com/[^\s"<>]+`)
	allURLs := urlRe.FindAllString(html, -1)

	// 按 URI (去掉参数) 去重, 每个 URI 只保留第一个(最大尺寸)
	uriRe := regexp.MustCompile(`(tos-cn-[^\s?~]+)`)
	for _, url := range allURLs {
		// 跳过带 water 水印的
		if strings.Contains(url, "-water") {
			continue
		}
		// 优先 noop (原图) 或 lqen-new (高清)
		if !strings.Contains(url, "noop") && !strings.Contains(url, "lqen-new") {
			continue
		}
		// 提取 URI 做去重
		uriMatch := uriRe.FindString(url)
		if uriMatch == "" {
			continue
		}
		if seen[uriMatch] {
			continue
		}
		seen[uriMatch] = true
		// 优先 jpeg 格式
		images = append(images, url)
	}

	// 如果上面没找到，用更宽松的匹配
	if len(images) == 0 {
		for _, url := range allURLs {
			if strings.Contains(url, "-water") {
				continue
			}
			uriMatch := uriRe.FindString(url)
			if uriMatch == "" {
				continue
			}
			if seen[uriMatch] {
				continue
			}
			seen[uriMatch] = true
			images = append(images, url)
		}
	}

	return images
}

// 抖音解析 - 支持视频和图文
func DouYin(rUrl string, ua string) (string, error) {
	mobileUA := "Mozilla/5.0 (iPhone; CPU iPhone OS 16_0 like Mac OS X) AppleWebKit/605.1.15 (KHTML, like Gecko) Version/16.0 Mobile/15E148 Safari/604.1"

	// Step 1: 跟随短链获取真实URL
	realUrl, err := getFinalURL(rUrl, mobileUA)
	if err != nil {
		return "", fmt.Errorf("短链跳转失败: %v", err)
	}
	log.Println("真实URL:", realUrl)

	// Step 1.5: 检查是否是图文 note 类型
	if strings.Contains(realUrl, "/note/") {
		return douyinNote(realUrl, mobileUA)
	}

	// 以下是视频解析逻辑
	// Step 2: 提取视频ID
	re := regexp.MustCompile(`/video/(\d+)`)
	matches := re.FindStringSubmatch(realUrl)
	if len(matches) < 2 {
		// 备用：纯数字匹配
		re2 := regexp.MustCompile(`(\d{15,25})`)
		matches2 := re2.FindAllString(realUrl, -1)
		if len(matches2) == 0 {
			return "", errors.New("无效链接：无法提取视频ID")
		}
		matches = []string{"", matches2[0]}
	}
	videoId := matches[1]
	fmt.Println("视频ID:", videoId)

	// Step 3: 请求分享页面获取 video_id（短ID，如 v0200f930000bpdpr5a6tgq5bkt31s50）
	sharePageURL := fmt.Sprintf("https://www.iesdouyin.com/share/video/%s/", videoId)
	client := &http.Client{Timeout: 15 * time.Second}
	req, err := http.NewRequest("GET", sharePageURL, nil)
	if err != nil {
		return "", err
	}
	req.Header.Set("User-Agent", mobileUA)
	resp, err := client.Do(req)
	if err != nil {
		return "", fmt.Errorf("分享页请求失败: %v", err)
	}
	defer resp.Body.Close()
	body, _ := ioutil.ReadAll(resp.Body)
	html := string(body)

	// 提取 video_id
	vidRe := regexp.MustCompile(`video_id=([a-zA-Z0-9]+)`)
	vidMatches := vidRe.FindStringSubmatch(html)
	if len(vidMatches) < 2 {
		return "", errors.New("无法从分享页提取video_id")
	}
	shortVideoId := vidMatches[1]
	fmt.Println("短video_id:", shortVideoId)

	// Step 4: 调用播放API获取无水印视频URL
	playURL := fmt.Sprintf("https://api-play.amemv.com/aweme/v1/play/?video_id=%s&line=0&ratio=720p&watermark=0&media_type=4&vr_type=0&improve_bitrate=0&logo_name=aweme", shortVideoId)

	finalURL, err := getFinalURL(playURL, mobileUA)
	if err != nil {
		return "", fmt.Errorf("播放API请求失败: %v", err)
	}

	// 检查是否是有效URL
	if !strings.Contains(finalURL, "douyinvod.com") && !strings.Contains(finalURL, "douyinstatic.com") {
		return "", errors.New("播放API未返回有效视频地址")
	}

	fmt.Println("无水印视频URL:", finalURL[:100]+"...")
	return finalURL, nil
}

// 抖音图文解析
func douyinNote(rUrl, ua string) (string, error) {
	log.Println("抖音图文:", rUrl)

	client := &http.Client{Timeout: 15 * time.Second}
	req, err := http.NewRequest("GET", rUrl, nil)
	if err != nil {
		return "", err
	}
	req.Header.Set("User-Agent", ua)
	resp, err := client.Do(req)
	if err != nil {
		return "", fmt.Errorf("请求失败: %v", err)
	}
	defer resp.Body.Close()
	body, _ := ioutil.ReadAll(resp.Body)
	html := string(body)

	images := extractDouyinNoteImages(html)
	if len(images) == 0 {
		return "", errors.New("未找到图片")
	}

	log.Printf("抖音图文: 找到 %d 张图片\n", len(images))
	return "IMAGE:" + strings.Join(images, ","), nil
}
