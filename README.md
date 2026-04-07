# watermark-server

短视频去水印 API 服务，支持 **抖音、快手、BiliBili** 三大平台。

> Go 1.20+ / Gin 1.9 / MIT

## 功能

- 🎥 视频去水印 — 无水印视频直链
- 🖼️ 图文解析 — 提取所有图片
- 📋 元信息提取 — 文案、话题、封面
- 🔗 短链跟随 — 自动跳转短链
- 📋 口令识别 — 直接粘贴分享文字

## 快速开始

```bash
git clone https://github.com/2759069519/watermark-server.git
cd watermark-server
go build -o watermark-server .
./watermark-server
```

默认监听 `:9000`，访问 http://localhost:9000 使用 Web 界面。

## API

### 解析

```
GET  /parse?url=<链接或分享文字>
POST /parse  Body: {"url": "<链接或分享文字>"}
```

也兼容旧参数 `key_words`。

### 返回结构

```json
{
  "code": 200,
  "data": {
    "platform": "douyin",
    "title": "文案内容",
    "topics": ["#话题1", "#话题2"],
    "cover": "https://封面URL",
    "video": "https://无水印视频URL",
    "images": [],
    "short_url": "https://v.douyin.com/xxxxx/"
  },
  "msg": "解析完成"
}
```

| 字段 | 说明 |
|------|------|
| `platform` | 平台: douyin / kuaishou / bilibili |
| `title` | 视频标题或文案 |
| `topics` | 话题标签列表 |
| `cover` | 封面图 URL |
| `video` | 无水印视频 URL（视频类） |
| `images` | 图片列表（图文类） |
| `short_url` | 原始输入链接 |

### 示例

```bash
# 抖音视频
curl "http://localhost:9000/parse?url=https://v.douyin.com/xxxxx/"

# 快手视频
curl "http://localhost:9000/parse?url=https://v.kuaishou.com/xxxxx"

# B站视频
curl "http://localhost:9000/parse?url=https://b23.tv/xxxxx"

# 粘贴分享口令也行
curl -X POST http://localhost:9000/parse \
  -H "Content-Type: application/json" \
  -d '{"url": "复制这段描述... https://v.douyin.com/xxxxx/ ...打开抖音"}'
```

## 项目结构

```
main.go                  → 入口 & 路由
router/index.go          → 路由定义
controllers/
  indexController.go     → 解析控制器
  proxyController.go     → B站视频代理
handleVideo/
  types.go               → VideoInfo 统一结构
  douyin.go              → 抖音解析
  kuaishou.go            → 快手解析
  bilibili.go            → BiliBili 解析
modules/
  baseController.go      → 响应封装
  request.go             → HTTP 工具
  utils.go               → URL 提取
template/dist/           → 前端页面
```

## 部署

### Nginx 反代

```nginx
server {
    listen 80;
    server_name your-domain.com;
    location / {
        proxy_pass http://127.0.0.1:9000;
        proxy_set_header Host $host;
        proxy_set_header X-Real-IP $remote_addr;
    }
}
```

### Systemd

```ini
[Unit]
Description=Watermark Server
After=network.target

[Service]
Type=simple
WorkingDirectory=/opt/watermark-server
ExecStart=/opt/watermark-server/watermark-server
Restart=always

[Install]
WantedBy=multi-user.target
```

```bash
sudo systemctl daemon-reload
sudo systemctl enable --now watermark-server
```

## 注意

- B站视频因防盗链，首次请求可能需要等待 CDN 响应（~20s）
- 请合理使用，遵守各平台使用条款

## License

[MIT](LICENSE)
