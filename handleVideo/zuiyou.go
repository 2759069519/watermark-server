package handleVideo

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
)

const zuiYouUrl = "https://share.xiaochuankeji.cn/planck/share/post/detail"

func requestZuiYou(rUrl, id, ua string) ([]byte, error) {
	data := fmt.Sprintf("{\"pid\":%s}", id)
	req, err := http.NewRequest("POST", rUrl, bytes.NewBuffer([]byte(data)))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
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

// 最右解析 - 修复：添加nil指针检查
func ZuiYou(rUrl, ua string) (string, error) {
	parse, err := url.Parse(rUrl)
	if err != nil {
		return "", err
	}
	query := parse.Query()
	if len(query["pid"]) == 0 {
		return "", errors.New("无效地址：缺少pid参数")
	}
	pid := query["pid"][0]
	body, err := requestZuiYou(zuiYouUrl, pid, ua)
	if err != nil {
		return "", err
	}
	var res *struct {
		Data *struct {
			Post *struct {
				Videos map[string]interface{} `json:"videos"`
			} `json:"post"`
		} `json:"data"`
	}
	err = json.Unmarshal(body, &res)
	if err != nil {
		return "", errors.New("解析响应失败")
	}
	// 修复：添加nil检查
	if res == nil || res.Data == nil || res.Data.Post == nil {
		return "", errors.New("未获取到视频数据")
	}
	videos := res.Data.Post.Videos
	if videos == nil || len(videos) == 0 {
		return "", errors.New("未找到有效视频")
	}
	var info interface{}
	for _, videoInfo := range videos {
		info = videoInfo
	}
	var videoUrl string
	if info, ok := info.(map[string]interface{}); ok {
		videoUrl = info["url"].(string)
	} else {
		return "", errors.New("解析失败")
	}
	return videoUrl, nil
}
