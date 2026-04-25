# BT-Music 🎵

音乐下载工具。优先从 B站 用 yt-dlp 下载高质量 mp3，搜不到时通过 BT 多源搜索获取磁力链接。

本项目同时支持交互式使用和 Hermes agent 调用。给 agent 使用时建议始终调用编译后的二进制，并使用一次性命令和 `--json`，避免进入 REPL。

从 [BT-Spider](https://github.com/huangke19/BT-Spider) 拆分而来的独立轻量版本，无需 BT 下载引擎。

## 功能

- `music` 命令：搜索 B站，返回视频结果，可继续下载 mp3（搜索走 B站 API，下载走 yt-dlp）
- `bt` 命令：多源 BT 搜索（ThePirateBay / BTDigg / Nyaa），返回磁力链接
- `download` 命令：按 B站 URL 或 BV 号下载 mp3
- `get` 命令：按关键词搜索 B站并下载指定序号结果，适合 agent 一步式调用
- `--json`：输出稳定 JSON，适合 Hermes agent 解析
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

### Hermes agent 调用

```bash
# 搜索 B站音乐视频
./bt-music --json --limit 5 music "周杰伦 以父之名"

# 搜索 BT 磁力链接
./bt-music --json --limit 10 bt "Pink Floyd The Wall FLAC"

# 下载 B站视频音频为 mp3
./bt-music --json --output-dir /tmp/music download BV1xx411c7mD "song title"

# 一步式搜索并下载第 2 个 B站结果
./bt-music --json --pick 2 --output-dir /tmp/music get "周杰伦 以父之名"
```

当 stdin 不是终端时，不带命令会直接返回错误，避免 agent 误调用后等待交互输入。

Agent 约定：`--json` 模式下 stdout 只输出单个 JSON 对象，stderr 保持安静，便于程序解析。下载失败时会把 `yt-dlp` 的关键 stderr 合并进 `error.message`。

成功时退出码为 `0`，失败时退出码为 `1`。`--json` 模式下 stdout 始终是单个 JSON 对象：

```json
{
  "ok": true,
  "version": "0.2.0",
  "command": "bt",
  "keyword": "Pink Floyd The Wall FLAC",
  "count": 1,
  "results": [
    {
      "name": "Pink Floyd - The Wall",
      "size": "433.6 MB",
      "seeders": 106,
      "leechers": 4,
      "magnet": "magnet:?xt=urn:btih:...",
      "source": "ThePirateBay",
      "info_hash": "..."
    }
  ]
}
```

错误时：

```json
{
  "ok": false,
  "version": "0.2.0",
  "count": 0,
  "error": {
    "message": "..."
  }
}
```

### 交互式运行

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
| `music <关键词>` | 搜索 B站，交互模式下可输入序号下载 mp3 |
| `bt <关键词>` | BT 多源搜索，交互模式下可输入序号获取磁力链接 |
| `download <B站URL或BV号> [文件名]` | 直接下载 mp3 |
| `get <关键词>` | 搜索 B站并下载 `--pick` 指定序号的结果，默认第 1 个 |
| `quit` / `q` | 退出 |

### Agent flags

| 参数 | 说明 |
|------|------|
| `--json` | 输出机器可读 JSON |
| `--limit N` | 最多返回 N 条搜索结果，默认 10 |
| `--pick N` | `get` 命令下载第 N 条结果，默认 1 |
| `--output-dir DIR` / `--dir DIR` | 下载目录，默认 `~/Downloads/BT-Music/` |

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
| B站 | 中文音乐、MV、Hi-Res | 搜索 API + curl 兜底，下载用 yt-dlp |
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
│   ├── bilibili.go      # B站 API 搜索 + yt-dlp 下载
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
