# BT-Music 🎵

音乐下载工具。优先从 B站 用 yt-dlp 下载高质量 mp3，搜不到时通过 BT 多源搜索获取磁力链接。

从 [BT-Spider](https://github.com/huangke19/BT-Spider) 拆分而来的独立轻量版本，无需 BT 下载引擎。

## 功能

- `music` 命令：搜索 B站，选序号下载 mp3（通过 yt-dlp）
- `bt` 命令：多源 BT 搜索（ThePirateBay / BTDigg / Nyaa），选序号获取磁力链接
- 自动过滤不相关 BT 结果
- 支持中英文关键词
- 代理支持（HTTP_PROXY / HTTPS_PROXY）

## 依赖

- [yt-dlp](https://github.com/yt-dlp/yt-dlp)：`brew install yt-dlp`

## 快速开始

### 编译

```bash
go build -o bt-music .
```

### 运行

```bash
# 默认下载目录：~/Downloads/BT-Music/
./bt-music

# 自定义下载目录
./bt-music /path/to/music

# 需要代理时（BT 搜索需要）
HTTPS_PROXY=http://127.0.0.1:7890 ./bt-music
```

### 命令

| 命令 | 说明 |
|------|------|
| `music <关键词>` | 搜索 B站，列出结果，输入序号下载 mp3 |
| `bt <关键词>` | BT 多源搜索，输入序号获取磁力链接 |
| `quit` / `q` | 退出 |

### 示例

```
bt-music> music 周杰伦 以父之名
找到 10 个结果:

  [1] 【4K顶级修复】周杰伦 - 以父之名 MV Flac无损音质封装！ | 364s | UP: 唯一8090
  [2] 【Hi-Res无损音质】｜《以父之名》- 周杰伦 | 343s | UP: VV音乐局
  ...

输入序号下载 mp3（回车跳过）:
1
⬇️  下载: 【4K顶级修复】周杰伦 - 以父之名 MV Flac无损音质封装！
✓ 保存到: ~/Downloads/BT-Music/...mp3
```

```
bt-music> bt Pink Floyd The Wall FLAC
找到 68 个结果（按做种数排序）:

  [1] Pink Floyd - The Wall (2007 Remaster) [FLAC] 88
      433.6 MB | Seeders: 106 | ThePirateBay
  ...

输入序号复制磁力链接（回车跳过）:
1
🔗 磁力链接 [Pink Floyd - The Wall (2007 Remaster) [FLAC] 88]:
magnet:?xt=urn:btih:...
```

## 搜索源

| 来源 | 擅长 | 接口 |
|------|------|------|
| B站（yt-dlp） | 中文音乐、MV、Hi-Res | yt-dlp 抓取 |
| ApiBay (TPB) | 英文专辑、FLAC | JSON API |
| BTDigg | 综合 DHT | HTML 爬取 |
| Nyaa | 日本音乐、动漫 OST | RSS |

## 下载目录

默认：`~/Downloads/BT-Music/`

## 项目结构

```
.
├── main.go              # CLI 入口，交互式 REPL
├── search/
│   ├── search.go        # Provider 接口、并发搜索、去重
│   ├── bilibili.go      # B站搜索 + yt-dlp 下载
│   ├── apibay.go        # ThePirateBay（JSON API）
│   ├── btdig.go         # BTDigg（JSON API）
│   └── nyaa.go          # Nyaa（RSS，动漫/日本音乐）
└── pkg/
    ├── httputil/
    │   └── client.go    # 共享 HTTP 客户端（代理、UA、超时）
    └── utils/
        └── format.go    # FormatBytes

```

## 相关项目

- [BT-Spider](https://github.com/huangke19/BT-Spider) — 完整 BT 下载工具，含引擎 + Telegram Bot
- [BT-Books](https://github.com/huangke19/BT-Books) — 同系列电子书下载工具（Z-Library）

## 许可证

MIT
