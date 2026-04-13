# BT-Music 🎵

音乐下载工具。优先 B站 yt-dlp 下载高质量 mp3，搜不到时 BT 搜索获取磁力链接。

## 依赖

- [yt-dlp](https://github.com/yt-dlp/yt-dlp)：`brew install yt-dlp`

## 用法

```bash
./bt-music
```

### 命令

| 命令 | 说明 |
|------|------|
| `music <关键词>` | 搜索B站，选序号下载 mp3 |
| `bt <关键词>` | BT搜索（建议英文），选序号获取磁力链接 |
| `quit` | 退出 |

### 示例

```
bt-music> music 周杰伦 稻香
bt-music> music Taylor Swift Anti-Hero
bt-music> bt Pink Floyd The Wall FLAC
```

## 下载目录

默认：`~/Downloads/BT-Music/`

可通过命令行参数指定：

```bash
./bt-music /path/to/music
```

## 代理

BT 搜索需要代理，设置 `HTTPS_PROXY`：

```bash
HTTPS_PROXY=http://127.0.0.1:7890 ./bt-music
```

## 编译

```bash
go build -o bt-music .
```
