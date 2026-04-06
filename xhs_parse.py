#!/usr/bin/env python3
"""
小红书视频解析脚本 - 被 Go 服务通过 subprocess 调用
用法: python3 xhs_parse.py <分享链接>
输出: JSON {"code": 200, "url": "无水印视频链接"} 或 {"code": 400, "msg": "错误信息"}
"""
import sys
import json
import re
import httpx
from xhshow import Xhshow

API_URL = "https://edith.xiaohongshu.com/api/sns/web/v1/feed"
UA = "Mozilla/5.0 (iPhone; CPU iPhone OS 16_0 like Mac OS X) AppleWebKit/605.1.15 (KHTML, like Gecko) Version/16.0 Mobile/15E148 Safari/604.1"


def extract_note_id(url: str) -> str:
    patterns = [
        r'xiaohongshu\.com/(?:explore|discovery/item|note)/([a-f0-9]+)',
        r'xiaohongshu\.com/(?:explore|discovery/item|note)/([a-zA-Z0-9]+)',
    ]
    for p in patterns:
        m = re.search(p, url)
        if m:
            return m.group(1)
    return ""


def follow_redirect(url: str) -> str:
    try:
        with httpx.Client(follow_redirects=True, timeout=10) as client:
            resp = client.get(url, headers={"User-Agent": UA})
            return str(resp.url)
    except Exception:
        return url



def parse_xhs(share_url: str) -> dict:
    try:
        if "xhslink.com" in share_url or "xhs.com" in share_url:
            share_url = follow_redirect(share_url)

        note_id = extract_note_id(share_url)
        if not note_id:
            return {"code": 400, "msg": "无法从链接中提取笔记ID"}

        encipher = Xhshow()
        params = {"source_note_id": note_id}
        cookies = "a1=1"
        headers = encipher.sign_headers_get(
            uri=API_URL,
            cookies=cookies,
            params=params,
        )
        headers["User-Agent"] = UA
        headers["Referer"] = "https://www.xiaohongshu.com/"
        headers["Cookie"] = cookies

        with httpx.Client(follow_redirects=True, timeout=15) as client:
            resp = client.get(API_URL, params=params, headers=headers)
            data = resp.json()

        if data.get("code") != 0:
            return {"code": 400, "msg": f"API返回错误: {data.get('msg', data.get('code'))}"}

        items = data.get("data", {}).get("items", [])
        if not items:
            return {"code": 400, "msg": "未找到笔记数据"}

        note = items[0].get("note_card", items[0])

        # 方式1: originVideoKey
        video_key = note.get("video", {}).get("consumer", {}).get("originVideoKey", "")
        if video_key:
            return {"code": 200, "url": f"https://sns-video-bd.xhscdn.com/{video_key}"}

        # 方式2: stream.h264
        h264_list = note.get("video", {}).get("media", {}).get("stream", {}).get("h264", [])
        if h264_list:
            best = h264_list[0]
            url = best.get("masterUrl") or (best.get("backupUrls") or [None])[0]
            if url:
                return {"code": 200, "url": url}

        return {"code": 400, "msg": "未找到视频链接（可能是图文笔记）"}

    except Exception as e:
        return {"code": 500, "msg": f"解析异常: {str(e)}"}


if __name__ == "__main__":
    if len(sys.argv) < 2:
        print(json.dumps({"code": 400, "msg": "缺少链接参数"}))
        sys.exit(1)

    result = parse_xhs(sys.argv[1])
    print(json.dumps(result, ensure_ascii=False))
