package controllers

import (
	"encoding/base64"
	"io"
	"net/http"
	"time"
	"github.com/gin-gonic/gin"
)

func Proxy(ctx *gin.Context) {
	var targetUrl string

	// 支持路径参数 /proxy/<base64>
	encoded := ctx.Param("encoded")
	if encoded != "" {
		// 去掉开头的 /
		if encoded[0] == '/' {
			encoded = encoded[1:]
		}
		decoded, err := base64.StdEncoding.DecodeString(encoded)
		if err == nil {
			targetUrl = string(decoded)
		}
	}

	// 支持 POST JSON body
	if targetUrl == "" && ctx.Request.Method == "POST" {
		var body struct {
			Url string `json:"url"`
		}
		if err := ctx.ShouldBindJSON(&body); err == nil {
			targetUrl = body.Url
		}
	}

	// 支持 GET query
	if targetUrl == "" {
		targetUrl = ctx.Query("url")
	}

	if targetUrl == "" {
		ctx.JSON(http.StatusBadRequest, gin.H{"code": 400, "msg": "缺少url参数"})
		return
	}

	client := &http.Client{Timeout: 60 * time.Second}
	req, err := http.NewRequest("GET", targetUrl, nil)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"code": 500, "msg": "请求失败"})
		return
	}
	req.Header.Set("Referer", "https://www.bilibili.com/")
	req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/120.0.0.0 Safari/537.36")

	resp, err := client.Do(req)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"code": 500, "msg": "请求失败"})
		return
	}
	defer resp.Body.Close()

	ctx.Header("Content-Type", resp.Header.Get("Content-Type"))
	ctx.Header("Content-Length", resp.Header.Get("Content-Length"))
	ctx.Status(resp.StatusCode)
	io.Copy(ctx.Writer, resp.Body)
}
