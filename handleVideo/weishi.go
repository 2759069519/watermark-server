package handleVideo

import (
	"encoding/json"
	"errors"
	"io/ioutil"
	"net/http"
	"net/url"
)

const weishiUrl = "https://h5.weishi.qq.com/webapp/json/weishi/WSH5GetPlayPage"

func requestWeiShi(rUrl, ua string) ([]byte, error) {
	req, err := http.NewRequest("GET", rUrl, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("User-Agent", ua)
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	return body, nil
}

// 微视解析 - 修复：添加nil指针检查，避免panic
func WeiShi(rUrl, ua string) (string, error) {
	parse, err := url.Parse(rUrl)
	if err != nil {
		return "", err
	}
	ids := parse.Query()["id"]
	if len(ids) == 0 {
		return "", errors.New("无效地址：缺少id参数")
	}
	id := ids[0]
	type Video struct {
		Data *struct {
			Feeds []*struct {
				Video_url string
			}
		} `json:"data"`
	}
	body, err := requestWeiShi(weishiUrl+"?feedid="+id, ua)
	if err != nil {
		return "", err
	}
	var res Video
	err = json.Unmarshal([]byte(body), &res)
	if err != nil {
		return "", errors.New("解析响应失败")
	}
	// 修复：添加nil检查
	if res.Data == nil {
		return "", errors.New("未获取到视频数据")
	}
	if len(res.Data.Feeds) == 0 {
		return "", errors.New("未找到有效视频")
	}
	if res.Data.Feeds[0] == nil {
		return "", errors.New("视频数据为空")
	}
	return res.Data.Feeds[0].Video_url, nil
}
