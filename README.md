# watermark-server

短视频去水印 API 服务，支持 **抖音、快手、BiliBili**。

## 快速开始

```bash
# 编译
go build -o watermark-server .

# 运行
./watermark-server
```

服务默认监听 `:9000` 端口。

## API

### 解析视频

```
GET /parse?key_words=<视频链接或分享文字>
```

**支持的平台：**
- 抖音（douyin.com）
- 快手（kuaishou.com）
- BiliBili（bilibili.com / b23.tv）

**支持粘贴分享文字**，自动提取链接：
```
GET /parse?key_words=7.99 hOk/ 看看【xxx的作品】https://v.douyin.com/xxxxx/
```

### 返回格式

```json
{
  "code": 200,
  "data": {
    "platformInfo": {
      "name": "抖音",
      "platform": "douyin"
    },
    "path": "https://v11-cold.douyinvod.com/...无水印视频直链..."
  },
  "msg": "解析完成"
}
```

`data.path` 即为无水印视频地址，可直接播放或下载。

> B站视频由于防盗链机制，返回的是代理链接，需要等待 CDN 响应（约 20 秒）。

## 示例

```bash
# 抖音
curl "http://localhost:9000/parse?key_words=https://v.douyin.com/xxxxx/"

# 快手
curl "http://localhost:9000/parse?key_words=https://v.kuaishou.com/xxxxx"

# B站
curl "http://localhost:9000/parse?key_words=https://b23.tv/xxxxx"
```

## 前端页面

访问 `http://localhost:9000/` 即可使用网页版。
