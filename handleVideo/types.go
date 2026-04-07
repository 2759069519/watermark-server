package handleVideo

// VideoInfo 统一返回结构
type VideoInfo struct {
	Platform string   `json:"platform"`  // 平台名: douyin / kuaishou / bilibili
	Title    string   `json:"title"`     // 标题/文案
	Topics   []string `json:"topics"`    // 话题标签
	Cover    string   `json:"cover"`     // 封面图 URL
	Video    string   `json:"video"`     // 无水印视频 URL（视频类）
	Images   []string `json:"images"`    // 图片列表（图文类）
	ShortURL string   `json:"short_url"` // 原始短链
}
