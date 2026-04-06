package modules

import (
	"errors"
	"fmt"
	"regexp"
	"strings"

	"mvdan.cc/xurls/v2"
)

// 常见短链域名
var shortDomainRe = regexp.MustCompile(`(?i)(v\.douyin\.com|v\.kuaishou\.com|b23\.tv|douyin\.com|kuaishou\.com|bilibili\.com)/\S+`)

// 提取文字中的url
func GetUrl(url string) (string, error) {
	// 方式1: 严格匹配 https?:// 开头的URL
	rxStrict := xurls.Strict()
	src := rxStrict.FindAllString(url, -1)
	fmt.Println(src)
	if len(src) > 0 {
		return src[0], nil
	}

	// 方式2: 宽松匹配（含无协议URL）
	rxRelaxed := xurls.Relaxed()
	src = rxRelaxed.FindAllString(url, -1)
	for _, s := range src {
		// 过滤：必须包含已知域名
		sLower := strings.ToLower(s)
		if strings.Contains(sLower, "douyin.com") ||
			strings.Contains(sLower, "kuaishou.com") ||
			strings.Contains(sLower, "bilibili.com") ||
			strings.Contains(sLower, "b23.tv") {
			// 补全协议
			if !strings.HasPrefix(sLower, "http") {
				s = "https://" + s
			}
			fmt.Println("relaxed:", s)
			return s, nil
		}
	}

	// 方式3: 正则兜底 - 匹配已知短链域名
	matches := shortDomainRe.FindStringSubmatch(url)
	if len(matches) > 0 {
		result := "https://" + matches[0]
		fmt.Println("fallback:", result)
		return result, nil
	}

	return "", errors.New("无效地址")
}
