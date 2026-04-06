<div align="center">

# 🎬 短视频去水印服务

<p>
  <img src="https://img.shields.io/badge/Go-1.20+-00ADD8?style=flat-square&logo=go&logoColor=white" alt="Go Version" />
  <img src="https://img.shields.io/badge/Gin-1.9-blue?style=flat-square&logo=gin" alt="Gin" />
  <img src="https://img.shields.io/badge/License-MIT-green?style=flat-square" alt="License" />
  <img src="https://img.shields.io/badge/平台-抖音|快手|BiliBili-ff69b4?style=flat-square" alt="Platforms" />
</p>

**一键获取无水印短视频链接，支持抖音、快手、BiliBili 三大平台**

[功能介绍](#-功能介绍) •
[快速开始](#-快速开始) •
[API 文档](#-api-文档) •
[项目结构](#-项目结构) •
[部署指南](#-部署指南)

</div>

---

## ✨ 功能介绍

| 功能 | 说明 |
|------|------|
| 🎥 **视频去水印** | 抖音、快手、BiliBili 视频无水印直链解析 |
| 🖼️ **图文解析** | 抖音/快手/B站图文帖子，自动提取所有图片 |
| 📋 **智能提取** | 支持粘贴分享口令，自动识别并提取链接 |
| 🔗 **短链跟随** | 自动跟随 `v.douyin.com`、`b23.tv` 等短链 |
| 🖥️ **Web 界面** | 内置前端页面，浏览器直接使用 |
| ⚡ **高性能** | 基于 Gin 框架，轻量高效 |

### 支持的平台

| 平台 | 视频 | 图文 | 短链 |
|------|:----:|:----:|:----:|
| 抖音 `douyin.com` | ✅ | ✅ | ✅ |
| 快手 `kuaishou.com` | ✅ | ✅ | ✅ |
| BiliBili `bilibili.com` | ✅ | ✅ | ✅ |

---

## 🚀 快速开始

### 环境要求

- **Go** >= 1.20

### 编译运行

```bash
# 克隆项目
git clone https://github.com/2759069519/watermark-server.git
cd watermark-server

# 编译
go build -o watermark-server .

# 运行（默认监听 :9000）
./watermark-server
```

服务启动后，访问 http://localhost:9000 即可使用 Web 界面。

---

## 📖 API 文档

### 解析接口

```
GET  /parse?key_words=<链接或分享文字>
POST /parse  Body: {"key_words": "<链接或分享文字>"}
```

#### 请求参数

| 参数 | 类型 | 必填 | 说明 |
|------|------|:----:|------|
| `key_words` | string | ✅ | 视频链接、短链或粘贴的分享口令 |

#### 成功响应

```json
{
  "code": 200,
  "data": {
    "platformInfo": {
      "name": "抖音",
      "platform": "douyin"
    },
    "path": "https://v11-cold.douyinvod.com/..."
  },
  "msg": "解析完成"
}
```

> **`data.path`** 即为无水印视频直链，可直接播放或下载。

#### 图文响应

```json
{
  "code": 200,
  "data": {
    "platformInfo": {
      "name": "抖音",
      "platform": "douyin"
    },
    "path": "",
    "images": [
      "https://p3-sign.douyinpic.com/...",
      "https://p9-sign.douyinpic.com/..."
    ]
  },
  "msg": "解析完成"
}
```

#### 错误响应

```json
{
  "code": 400,
  "msg": "暂不支持该平台！当前支持：抖音、快手、BiliBili",
  "data": null
}
```

### 代理接口（B站专用）

B站视频因防盗链机制，返回代理链接：

```
GET  /proxy/<base64编码的URL>
POST /proxy  Body: {"url": "https://..."}
```

---

## 💡 使用示例

### cURL

```bash
# 抖音视频（支持直接粘贴分享文字）
curl "http://localhost:9000/parse?key_words=https://v.douyin.com/xxxxx/"

# 快手视频
curl "http://localhost:9000/parse?key_words=https://v.kuaishou.com/xxxxx"

# BiliBili 视频
curl "http://localhost:9000/parse?key_words=https://b23.tv/xxxxx"

# POST 请求
curl -X POST http://localhost:9000/parse \
  -H "Content-Type: application/json" \
  -d '{"key_words": "https://v.douyin.com/xxxxx/"}'
```

### JavaScript

```javascript
const res = await fetch('http://localhost:9000/parse?key_words=' + encodeURIComponent(videoUrl));
const { data } = await res.json();
console.log('无水印链接:', data.path);
```

---

## 📁 项目结构

```
watermark-server/
├── main.go                          # 🚀 入口 & 路由注册
├── go.mod                           # 📦 依赖管理
├── router/
│   └── index.go                     # 🔀 路由定义
├── controllers/
│   ├── indexController.go           # 🎯 解析控制器（核心逻辑）
│   └── proxyController.go           # 🔀 B站视频代理
├── handleVideo/
│   ├── douyin.go                    # 🎵 抖音解析（视频 + 图文）
│   ├── kuaishou.go                  # ⚡ 快手解析（视频 + 图文）
│   └── bilibili.go                  # 📺 BiliBili 解析（视频 + 图文）
├── modules/
│   ├── baseController.go            # 📤 响应封装
│   ├── request.go                   # 🌐 HTTP 请求工具
│   ├── utils.go                     # 🔧 URL 提取工具
│   └── utils_test.go                # 🧪 测试
├── template/
│   └── dist/                        # 🖥️ 前端页面
├── .gitignore
└── README.md
```

---

## 🛠️ 部署指南

### Nginx 反向代理

```nginx
server {
    listen 80;
    server_name your-domain.com;

    location / {
        proxy_pass http://127.0.0.1:9000;
        proxy_set_header Host $host;
        proxy_set_header X-Real-IP $remote_addr;
        proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
        proxy_set_header X-Forwarded-Proto $scheme;
    }
}
```

### Systemd 服务

```ini
# /etc/systemd/system/watermark-server.service
[Unit]
Description=Watermark Server - 短视频去水印服务
After=network.target

[Service]
Type=simple
User=root
WorkingDirectory=/opt/watermark-server
ExecStart=/opt/watermark-server/watermark-server
Restart=always
RestartSec=5

[Install]
WantedBy=multi-user.target
```

```bash
sudo systemctl daemon-reload
sudo systemctl enable --now watermark-server
```

---

## 📝 注意事项

- B站视频由于防盗链机制，首次请求可能需要等待 CDN 响应（约 20s）
- 抖音图文帖子返回 `images` 数组，包含所有无水印图片直链
- 快手图文帖子同样支持，返回图片列表
- 请合理使用，遵守各平台的使用条款

---

## 📄 License

[MIT](LICENSE)

---

<div align="center">

**如果觉得有用，给个 ⭐ Star 支持一下！**

</div>
