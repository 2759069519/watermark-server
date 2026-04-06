package controllers

import (
	"encoding/base64"
	"fmt"
	"net/http"
	"strings"
	"watermarkServer/handleVideo"
	"watermarkServer/modules"

	"github.com/gin-gonic/gin"
)

const (
	phone_ua = "Mozilla/5.0 (iPhone; CPU iPhone OS 13_2_3 like Mac OS X) AppleWebKit/605.1.15 (KHTML, like Gecko) Version/13.0.3 Mobile/15E148 Safari/604.1"
)

type IndexController struct{}

var base = &modules.BaseController{}

func (ctr *IndexController) Index(ctx *gin.Context) {
	var path string
	var err error
	type Platform struct {
		Name     string `json:"name"`
		Platform string `json:"platform"`
	}
	var platformInfo = &Platform{}

	// 支持 GET query 和 POST JSON body
	var keyWords string = ctx.Query("key_words")
	if keyWords == "" && ctx.Request.Method == "POST" {
		var body struct {
			KeyWords string `json:"key_words"`
		}
		if bindErr := ctx.ShouldBindJSON(&body); bindErr == nil {
			keyWords = body.KeyWords
		}
	}
	if keyWords == "" {
		ctx.JSON(http.StatusOK, gin.H{
			"code": http.StatusBadRequest,
			"msg":  "缺少关键词",
			"data": nil,
		})
		return
	}
	rUrl, err := modules.GetUrl(keyWords)
	if err != nil {
		base.Err(ctx, err.Error())
		return
	}

	if strings.Contains(rUrl, "douyin.com") { // 抖音（含抖音火山版）
		platformInfo.Name = "抖音"
		platformInfo.Platform = "douyin"
		path, err = handleVideo.DouYin(rUrl, phone_ua)
		fmt.Println("Index", path, err)
	} else if strings.Contains(rUrl, "kuaishou.com") { // 快手
		platformInfo.Name = "快手"
		platformInfo.Platform = "kuaishou"
		path, err = handleVideo.KuaiShou(rUrl, phone_ua)
	} else if strings.Contains(rUrl, "b23.tv") || strings.Contains(rUrl, "bilibili.com") { // BiliBili
		platformInfo.Name = "BiliBili"
		platformInfo.Platform = "bilibili"
		path, err = handleVideo.BiliBili(rUrl, phone_ua)

	} else {
		platformInfo.Name = ""
		platformInfo.Platform = ""
		base.Err(ctx, "暂不支持该平台！当前支持：抖音、快手、BiliBili")
		return
	}

	if err != nil {
		base.Err(ctx, err.Error())
		return
	}
	var res = &struct {
		PlatformInfo *Platform  `json:"platformInfo"`
		Path         string     `json:"path"`
		Images       []string   `json:"images,omitempty"`
	}{
		PlatformInfo: platformInfo,
		Path:         path,
	}

	// B站代理
	if platformInfo.Platform == "bilibili" && path != "" {
		encoded := base64.StdEncoding.EncodeToString([]byte(path))
		scheme := "http"
		if ctx.Request.TLS != nil {
			scheme = "https"
		}
		res.Path = scheme + "://" + ctx.Request.Host + "/proxy/" + encoded
	}

	// 快手图文 - 解析 IMAGE: 前缀
	if strings.HasPrefix(path, "IMAGE:") {
		res.Path = ""
		urlStr := strings.TrimPrefix(path, "IMAGE:")
		res.Images = strings.Split(urlStr, ",")
	}

	base.Success(ctx, res, "解析完成")
}
