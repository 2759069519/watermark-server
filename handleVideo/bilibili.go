package handleVideo

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"regexp"
	"strings"
)

type biliViewResp struct {
	Code int `json:"code"`
	Data *struct {
		Cid int `json:"cid"`
		Aid int `json:"aid"`
	} `json:"data"`
}

type biliPlayResp struct {
	Code int `json:"code"`
	Data *struct {
		Durl []struct {
			Url string `json:"url"`
		} `json:"durl"`
	} `json:"data"`
}

func getBiLi(rUrl, ua string) ([]byte, error) {
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
	body, _ := ioutil.ReadAll(resp.Body)
	return body, nil
}

// 提取B站 opus/dynamic 图文中的图片
func extractBiliOpusImages(html string) []string {
	// 将 \u002F 转为 /
	html = strings.ReplaceAll(html, `\u002F`, "/")

	var images []string
	seen := make(map[string]bool)

	// 匹配 pics 数组中的 URL
	// 格式: "pics":[{"url":"https://i0.hdslb.com/bfs/article/xxx.jpg","width":779,"height":1100}]
	picsRe := regexp.MustCompile(`"pics":\[\{[^]]*?"url":"(https?://[^"]+)"`)
	matches := picsRe.FindAllStringSubmatch(html, -1)
	for _, m := range matches {
		if len(m) >= 2 {
			url := m[1]
			if !seen[url] {
				seen[url] = true
				images = append(images, url)
			}
		}
	}

	// 如果上面没找到，用更宽松的方式匹配 hdslb 图片 URL
	if len(images) == 0 {
		urlRe := regexp.MustCompile(`https?://[^\s"<>]+hdslb\.com[^\s"<>]+\.(?:jpg|png|webp)`)
		allURLs := urlRe.FindAllString(html, -1)
		for _, url := range allURLs {
			// 过滤掉头像、emoji、图标等
			if strings.Contains(url, "face") || strings.Contains(url, "emoji") ||
				strings.Contains(url, "icon") || strings.Contains(url, "logo") {
				continue
			}
			if !seen[url] {
				seen[url] = true
				images = append(images, url)
			}
		}
	}

	return images
}

// BiliBili解析 - 支持视频和图文(opus/dynamic)
func BiliBili(rUrl, ua string) (string, error) {
	// 提取 BV号 或 av号
	var bvid string
	var avid string
	var opusId string

	bvRe := regexp.MustCompile(`(BV[a-zA-Z0-9]+)`)
	avRe := regexp.MustCompile(`/video/av(\d+)`)
	opusRe := regexp.MustCompile(`/opus/(\d+)`)

	if m := bvRe.FindStringSubmatch(rUrl); len(m) >= 2 {
		bvid = m[1]
	} else if m := avRe.FindStringSubmatch(rUrl); len(m) >= 2 {
		avid = m[1]
	} else if m := opusRe.FindStringSubmatch(rUrl); len(m) >= 2 {
		opusId = m[1]
	} else {
		// 尝试跟随短链
		client := &http.Client{
			CheckRedirect: func(req *http.Request, via []*http.Request) error {
				return nil
			},
		}
		req, _ := http.NewRequest("GET", rUrl, nil)
		req.Header.Set("User-Agent", ua)
		resp, err := client.Do(req)
		if err == nil {
			finalUrl := resp.Request.URL.String()
			if m := bvRe.FindStringSubmatch(finalUrl); len(m) >= 2 {
				bvid = m[1]
			} else if m := avRe.FindStringSubmatch(finalUrl); len(m) >= 2 {
				avid = m[1]
			} else if m := opusRe.FindStringSubmatch(finalUrl); len(m) >= 2 {
				opusId = m[1]
			}
		}
		if bvid == "" && avid == "" && opusId == "" {
			return "", errors.New("无效地址：无法提取视频或内容ID")
		}
	}

	// 处理 opus 图文
	if opusId != "" {
		return biliOpus(opusId, ua)
	}

	// 以下是视频解析逻辑
	// Step 1: 获取视频信息（cid）
	var viewUrl string
	if bvid != "" {
		viewUrl = fmt.Sprintf("https://api.bilibili.com/x/web-interface/view?bvid=%s", bvid)
	} else {
		viewUrl = fmt.Sprintf("https://api.bilibili.com/x/web-interface/view?aid=%s", avid)
	}

	body, err := getBiLi(viewUrl, ua)
	if err != nil {
		return "", fmt.Errorf("获取视频信息失败: %v", err)
	}

	var viewResp biliViewResp
	if err := json.Unmarshal(body, &viewResp); err != nil {
		return "", fmt.Errorf("解析视频信息失败: %v", err)
	}
	if viewResp.Code != 0 || viewResp.Data == nil {
		return "", errors.New("视频不存在或已下架")
	}
	cid := viewResp.Data.Cid

	// Step 2: 获取播放地址
	var playUrl string
	if bvid != "" {
		playUrl = fmt.Sprintf("https://api.bilibili.com/x/player/playurl?bvid=%s&cid=%d&qn=64&fnval=0", bvid, cid)
	} else {
		playUrl = fmt.Sprintf("https://api.bilibili.com/x/player/playurl?avid=%s&cid=%d&qn=64&fnval=0", avid, cid)
	}

	body, err = getBiLi(playUrl, ua)
	if err != nil {
		return "", fmt.Errorf("获取播放地址失败: %v", err)
	}

	var playResp biliPlayResp
	if err := json.Unmarshal(body, &playResp); err != nil {
		return "", fmt.Errorf("解析播放地址失败: %v", err)
	}
	if playResp.Code != 0 || playResp.Data == nil || len(playResp.Data.Durl) == 0 {
		return "", errors.New("无法获取播放地址")
	}

	return playResp.Data.Durl[0].Url, nil
}

// B站 opus 图文解析
func biliOpus(opusId, ua string) (string, error) {
	log.Println("B站 opus 图文:", opusId)

	// 获取移动端 opus 页面
	pageURL := fmt.Sprintf("https://m.bilibili.com/opus/%s", opusId)
	body, err := getBiLi(pageURL, ua)
	if err != nil {
		return "", fmt.Errorf("请求失败: %v", err)
	}
	html := string(body)

	images := extractBiliOpusImages(html)
	if len(images) == 0 {
		return "", errors.New("未找到图片")
	}

	log.Printf("B站 opus: 找到 %d 张图片\n", len(images))
	return "IMAGE:" + strings.Join(images, ","), nil
}
