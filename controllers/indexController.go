package controllers

import (
	"encoding/base64"
	"net/http"
	"strings"
	"watermarkServer/handleVideo"
	"watermarkServer/modules"

	"github.com/gin-gonic/gin"
)

type IndexController struct{}

var base = &modules.BaseController{}

// Parse 统一解析入口: GET/POST /parse?url=<链接或分享文字>
func (ctr *IndexController) Parse(ctx *gin.Context) {
	// 支持 GET ?url= 和 POST {"url": ""}
	var rawInput string = ctx.Query("url")
	if rawInput == "" {
		rawInput = ctx.Query("key_words") // 兼容旧参数
	}
	if rawInput == "" && ctx.Request.Method == "POST" {
		var body struct {
			URL      string `json:"url"`
			KeyWords string `json:"key_words"` // 兼容旧参数
		}
		if bindErr := ctx.ShouldBindJSON(&body); bindErr == nil {
			rawInput = body.URL
			if rawInput == "" {
				rawInput = body.KeyWords
			}
		}
	}
	if rawInput == "" {
		base.Err(ctx, "缺少链接参数 url")
		return
	}

	// 从分享文字中提取 URL
	rUrl, err := modules.GetUrl(rawInput)
	if err != nil {
		base.Err(ctx, "无效链接：未找到支持的平台URL")
		return
	}

	// 按平台分发
	var info *handleVideo.VideoInfo
	if strings.Contains(rUrl, "douyin.com") {
		info, err = handleVideo.DouYin(rUrl)
	} else if strings.Contains(rUrl, "kuaishou.com") {
		info, err = handleVideo.KuaiShou(rUrl)
	} else if strings.Contains(rUrl, "b23.tv") || strings.Contains(rUrl, "bilibili.com") {
		info, err = handleVideo.BiliBili(rUrl)
	} else {
		base.Err(ctx, "暂不支持该平台！当前支持：抖音、快手、BiliBili")
		return
	}

	if err != nil {
		base.Err(ctx, err.Error())
		return
	}

	// B站视频有防盗链，转为代理链接
	if info.Platform == "bilibili" && info.Video != "" {
		info.Video = buildProxyURL(ctx, info.Video)
	}

	base.Success(ctx, info, "解析完成")
}

// buildProxyURL 将直链转为本地代理链接
func buildProxyURL(ctx *gin.Context, rawURL string) string {
	encoded := base64.StdEncoding.EncodeToString([]byte(rawURL))
	scheme := "http"
	if ctx.Request.TLS != nil {
		scheme = "https"
	}
	return scheme + "://" + ctx.Request.Host + "/proxy/" + encoded
}
