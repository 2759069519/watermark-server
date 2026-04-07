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

const biliUA = "Mozilla/5.0 (iPhone; CPU iPhone OS 16_0 like Mac OS X) AppleWebKit/605.1.15 (KHTML, like Gecko) Version/16.0 Mobile/15E148 Safari/604.1"

type biliViewResp struct {
	Code int `json:"code"`
	Data *struct {
		Cid   int    `json:"cid"`
		Aid   int    `json:"aid"`
		Title string `json:"title"`
		Desc  string `json:"desc"`
		Pic   string `json:"pic"`
		Dynamic string `json:"dynamic"`
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

// BiliBili 解析B站视频和图文
func BiliBili(rUrl string) (*VideoInfo, error) {
	var bvid, avid, opusId string

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
		// 跟随短链
		client := &http.Client{CheckRedirect: func(req *http.Request, via []*http.Request) error { return nil }}
		req, _ := http.NewRequest("GET", rUrl, nil)
		req.Header.Set("User-Agent", biliUA)
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
			return nil, errors.New("无效地址：无法提取视频或内容ID")
		}
	}

	info := &VideoInfo{
		Platform: "bilibili",
		ShortURL: rUrl,
	}

	if opusId != "" {
		return biliOpus(opusId, info)
	}

	return biliVideo(bvid, avid, info)
}

// biliVideo 解析B站视频
func biliVideo(bvid, avid string, info *VideoInfo) (*VideoInfo, error) {
	var viewUrl string
	if bvid != "" {
		viewUrl = fmt.Sprintf("https://api.bilibili.com/x/web-interface/view?bvid=%s", bvid)
	} else {
		viewUrl = fmt.Sprintf("https://api.bilibili.com/x/web-interface/view?aid=%s", avid)
	}

	body, err := biliGet(viewUrl)
	if err != nil {
		return nil, fmt.Errorf("获取视频信息失败: %v", err)
	}

	var viewResp biliViewResp
	if err := json.Unmarshal(body, &viewResp); err != nil {
		return nil, fmt.Errorf("解析视频信息失败: %v", err)
	}
	if viewResp.Code != 0 || viewResp.Data == nil {
		return nil, errors.New("视频不存在或已下架")
	}

	// 提取元信息
	info.Title = viewResp.Data.Title
	if viewResp.Data.Desc != "" {
		info.Title = viewResp.Data.Desc
	}
	info.Cover = viewResp.Data.Pic

	// 从 dynamic 字段提取话题
	if viewResp.Data.Dynamic != "" {
		info.Topics = extractBiliTopics(viewResp.Data.Dynamic)
	}

	cid := viewResp.Data.Cid

	// 获取播放地址
	var playUrl string
	if bvid != "" {
		playUrl = fmt.Sprintf("https://api.bilibili.com/x/player/playurl?bvid=%s&cid=%d&qn=64&fnval=0", bvid, cid)
	} else {
		playUrl = fmt.Sprintf("https://api.bilibili.com/x/player/playurl?avid=%s&cid=%d&qn=64&fnval=0", avid, cid)
	}

	body, err = biliGet(playUrl)
	if err != nil {
		return nil, fmt.Errorf("获取播放地址失败: %v", err)
	}

	var playResp biliPlayResp
	if err := json.Unmarshal(body, &playResp); err != nil {
		return nil, fmt.Errorf("解析播放地址失败: %v", err)
	}
	if playResp.Code != 0 || playResp.Data == nil || len(playResp.Data.Durl) == 0 {
		return nil, errors.New("无法获取播放地址")
	}

	info.Video = playResp.Data.Durl[0].Url
	return info, nil
}

// biliOpus 解析B站 opus 图文
func biliOpus(opusId string, info *VideoInfo) (*VideoInfo, error) {
	log.Println("B站 opus 图文:", opusId)

	pageURL := fmt.Sprintf("https://m.bilibili.com/opus/%s", opusId)
	body, err := biliGet(pageURL)
	if err != nil {
		return nil, fmt.Errorf("请求失败: %v", err)
	}
	html := string(body)

	images := extractBiliOpusImages(html)
	if len(images) == 0 {
		return nil, errors.New("未找到图片")
	}
	info.Images = images

	// 提取标题
	info.Title = extractBiliField(html, "title")

	// 提取话题
	info.Topics = extractBiliTopics(html)

	// 封面 = 第一张图
	if len(images) > 0 {
		info.Cover = images[0]
	}

	log.Printf("B站 opus: %d张图片 title=%s", len(images), info.Title)
	return info, nil
}

func biliGet(url string) ([]byte, error) {
	client := &http.Client{}
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("User-Agent", biliUA)
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	body, _ := ioutil.ReadAll(resp.Body)
	return body, nil
}

func extractBiliField(html, field string) string {
	re := regexp.MustCompile(`"` + field + `"\s*:\s*"([^"]*)"`)
	m := re.FindStringSubmatch(html)
	if len(m) >= 2 {
		return strings.TrimSpace(m[1])
	}
	return ""
}

// extractBiliTopics 从HTML中提取B站话题
func extractBiliTopics(html string) []string {
	var topics []string
	seen := make(map[string]bool)

	// B站话题格式: "topic_name":"xxx" 或 #话题#
	re := regexp.MustCompile(`"topic_name"\s*:\s*"([^"]+)"`)
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

	// 备用: 匹配 #xxx# 格式
	if len(topics) == 0 {
		re2 := regexp.MustCompile(`#([^#<>]{1,30})#`)
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

// extractBiliOpusImages 提取B站 opus 图文中的图片
func extractBiliOpusImages(html string) []string {
	html = strings.ReplaceAll(html, `\u002F`, "/")

	var images []string
	seen := make(map[string]bool)

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

	if len(images) == 0 {
		urlRe := regexp.MustCompile(`https?://[^\s"<>]+hdslb\.com[^\s"<>]+\.(?:jpg|png|webp)`)
		allURLs := urlRe.FindAllString(html, -1)
		for _, url := range allURLs {
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
