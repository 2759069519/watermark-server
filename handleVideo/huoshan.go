package handleVideo

import (
	"encoding/json"
	"errors"
	"net/http"
	"net/url"
	"watermarkServer/modules"
)

const hs_url = "https://share.huoshan.com/api/item/info?item_id="

func getHSRealityUrl(rUrl, ua string) (url.Values, error) {
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
	return resp.Request.URL.Query(), nil
}

type HSRes struct {
	Data *struct {
		Item_info *struct {
			Cover string
			Url   string
		}
	}
}

// 火山解析 - 修复：添加nil指针检查
func HuoShan(rUrl, ua string) (string, error) {
	query, err := getHSRealityUrl(rUrl, ua)
	if err != nil {
		return "", err
	}
	itemId := query["item_id"]
	if itemId == nil {
		return "", errors.New("无效地址：缺少item_id参数")
	}
	body, err := modules.HttpGet(hs_url+itemId[0], ua)
	if err != nil {
		return "", err
	}
	var res HSRes
	json.Unmarshal(body, &res)
	if res.Data == nil || res.Data.Item_info == nil {
		return "", errors.New("未获取到视频数据")
	}
	if res.Data.Item_info.Url == "" {
		return "", errors.New("视频链接为空（可能已下架）")
	}
	return res.Data.Item_info.Url, nil
}
