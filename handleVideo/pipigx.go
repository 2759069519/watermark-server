package handleVideo

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"strconv"
	"strings"
)

const piPiGXUrl = "https://h5.pipigx.com/ppapi/share/fetch_content"

func requestPiPiGX(rUrl string, ids []string, ua string) ([]byte, error) {
	data := make(map[string]interface{})
	data["pid"], _ = strconv.Atoi(ids[0])
	data["mid"], _ = strconv.Atoi(ids[1])
	data["type"] = "post"
	bytesData, _ := json.Marshal(data)
	req, err := http.NewRequest("POST", rUrl, bytes.NewBuffer(bytesData))
	req.Header.Set("Content-Type", "text/plain;charset=UTF-8")
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

// 皮皮搞笑解析 - 修复：添加nil指针检查
func PiPiGX(rUrl, ua string) (string, error) {
	fmt.Println(rUrl)
	parse, err := url.Parse(rUrl)
	if err != nil {
		return "", err
	}
	query := parse.Query()
	if len(query["mid"]) == 0 {
		return "", errors.New("无效地址：缺少mid参数")
	}
	mid := query["mid"][0]
	paths := strings.Split(parse.Path, "/")
	pid := paths[len(paths)-1]
	body, err := requestPiPiGX(piPiGXUrl, []string{pid, mid}, ua)
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
