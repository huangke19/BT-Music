package search

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

// BilibiliSearch 在B站搜索歌曲，返回视频列表（BV号+标题）
// 通过 yt-dlp 的 flat-playlist 模式获取搜索结果
func BilibiliSearch(keyword string, limit int) ([]BilibiliResult, error) {
	if limit <= 0 {
		limit = 10
	}
	searchURL := fmt.Sprintf("https://search.bilibili.com/all?keyword=%s&search_type=video", keyword)
	args := []string{
		"--flat-playlist",
		"--playlist-end", fmt.Sprintf("%d", limit),
		"-j",
		"--no-warnings",
		searchURL,
	}
	cmd := exec.Command("yt-dlp", args...)
	cmd.Env = os.Environ()
	out, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("yt-dlp 搜索失败: %w", err)
	}

	var results []BilibiliResult
	for _, line := range strings.Split(string(out), "\n") {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		// 解析 JSON，提取 id 和 title
		id := jsonExtract(line, "id")
		title := jsonExtract(line, "title")
		uploader := jsonExtract(line, "uploader")
		duration := jsonExtract(line, "duration")
		if id == "" || title == "" {
			continue
		}
		results = append(results, BilibiliResult{
			BVid:      id,
			Title:     title,
			Uploader:  uploader,
			Duration:  duration,
			URL:       "https://www.bilibili.com/video/" + id,
		})
	}
	return results, nil
}

// BilibiliResult B站视频结果
type BilibiliResult struct {
	BVid     string
	Title    string
	Uploader string
	Duration string
	URL      string
}

// BilibiliDownload 用 yt-dlp 从B站下载音频为 mp3
func BilibiliDownload(bvURL, destDir, filename string) (string, error) {
	if destDir == "" {
		home, _ := os.UserHomeDir()
		destDir = filepath.Join(home, "Downloads", "BT-Music")
	}
	if err := os.MkdirAll(destDir, 0755); err != nil {
		return "", err
	}

	outTmpl := filepath.Join(destDir, filename+".%(ext)s")
	if filename == "" {
		outTmpl = filepath.Join(destDir, "%(title)s.%(ext)s")
	}

	args := []string{
		"--extract-audio",
		"--audio-format", "mp3",
		"--audio-quality", "0",
		"-o", outTmpl,
		"--no-playlist",
		bvURL,
	}
	cmd := exec.Command("yt-dlp", args...)
	cmd.Env = os.Environ()
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		return "", fmt.Errorf("yt-dlp 下载失败: %w", err)
	}

	// 找到下载的文件
	if filename != "" {
		outFile := filepath.Join(destDir, filename+".mp3")
		if _, err := os.Stat(outFile); err == nil {
			return outFile, nil
		}
	}
	return "", nil
}

// jsonExtract 简单从 JSON 行提取字符串字段（避免引入 encoding/json 依赖的复杂性）
func jsonExtract(line, key string) string {
	needle := `"` + key + `":`
	idx := strings.Index(line, needle)
	if idx < 0 {
		return ""
	}
	rest := strings.TrimSpace(line[idx+len(needle):])
	if strings.HasPrefix(rest, `"`) {
		// 字符串值
		rest = rest[1:]
		end := strings.Index(rest, `"`)
		if end < 0 {
			return ""
		}
		return rest[:end]
	}
	// 数字或布尔
	end := strings.IndexAny(rest, ",}")
	if end < 0 {
		return rest
	}
	return strings.TrimSpace(rest[:end])
}
