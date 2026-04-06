package handleVideo

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"regexp"
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

// BiliBili解析 - 修复：改用B站API接口，不再依赖HTML页面结构
func BiliBili(rUrl, ua string) (string, error) {
	// 提取 BV号 或 av号
	var bvid string
	var avid string

	bvRe := regexp.MustCompile(`(BV[a-zA-Z0-9]+)`)
	avRe := regexp.MustCompile(`/video/av(\d+)`)

	if m := bvRe.FindStringSubmatch(rUrl); len(m) >= 2 {
		bvid = m[1]
	} else if m := avRe.FindStringSubmatch(rUrl); len(m) >= 2 {
		avid = m[1]
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
			}
		}
		if bvid == "" && avid == "" {
			return "", errors.New("无效地址：无法提取视频ID")
		}
	}

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
