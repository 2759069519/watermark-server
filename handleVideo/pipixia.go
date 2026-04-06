package handleVideo

import (
	"errors"
	"fmt"
	"net/http"
	"regexp"
	"watermarkServer/modules"
)

func getPiPiXia(rUrl string, ua string) (string, error) {
	client := http.Client{}
	req, err := http.NewRequest("GET", rUrl, nil)
	if err != nil {
		return "", err
	}
	req.Header.Set("User-Agent", ua)
	resp, err := client.Do(req)
	if err != nil {
		return "", err
	}
	fmt.Println(resp.Request.URL.RequestURI(), resp.Request.URL.Path)
	path := resp.Request.URL.Path
	return path, nil
}

// 皮皮虾解析 - 修复：添加nil检查
func PiPixia(rUrl, ua string) (string, error) {
	path, err := getPiPiXia(rUrl, ua)
	if err != nil {
		return "", err
	}
	reg := regexp.MustCompile("[0-9]+")
	ids := reg.FindAllString(path, -1)
	if len(ids) == 0 {
		return "", errors.New("无效地址：未找到视频ID")
	}
	infoUrl := "https://is.snssdk.com/bds/cell/detail/?cell_type=1&aid=1319&app_name=super&cell_id=" + ids[0]
	body, err := modules.HttpGet(infoUrl, ua)
	if err != nil {
		return "", err
	}
	reg2 := regexp.MustCompile("origin_video_download.*?url_list.*?url.*?:\"(.*?)\"")
	matches := reg2.FindStringSubmatch(string(body))
	if len(matches) < 2 {
		return "", errors.New("未找到视频链接（可能视频已下架）")
	}
	return matches[1], nil
}
