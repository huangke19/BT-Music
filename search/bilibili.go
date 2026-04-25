package search

import (
	"bytes"
	"encoding/json"
	"fmt"
	"html"
	"io"
	"net/http"
	"net/url"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"

	"github.com/huangke/bt-music/pkg/httputil"
)

// BilibiliSearch 在B站搜索歌曲，返回视频列表（BV号+标题）
// 直接调用 B站搜索 API，避免 yt-dlp 对 search.bilibili.com URL 支持变动导致 agent 调用失败。
func BilibiliSearch(keyword string, limit int) ([]BilibiliResult, error) {
	if limit <= 0 {
		limit = 10
	}
	// B站这个接口加 page_size 时容易触发 412；保持最小参数，返回后本地按 limit 截断。
	apiURL := fmt.Sprintf("https://api.bilibili.com/x/web-interface/search/type?search_type=video&keyword=%s", url.QueryEscape(keyword))
	body, err := fetchBilibiliSearchBody(apiURL)
	if err != nil {
		return nil, err
	}
	var data struct {
		Code    int    `json:"code"`
		Message string `json:"message"`
		Data    struct {
			Result []struct {
				BVID     string          `json:"bvid"`
				Title    string          `json:"title"`
				Author   string          `json:"author"`
				Duration json.RawMessage `json:"duration"`
				ArcURL   string          `json:"arcurl"`
			} `json:"result"`
		} `json:"data"`
	}
	if err := json.Unmarshal(body, &data); err != nil {
		return nil, fmt.Errorf("B站搜索响应解析失败: %w", err)
	}
	if data.Code != 0 {
		return nil, fmt.Errorf("B站搜索失败: %s", data.Message)
	}

	var results []BilibiliResult
	for _, item := range data.Data.Result {
		duration := strings.Trim(string(item.Duration), `"`)
		if duration == "null" || duration == "0" {
			duration = ""
		}
		if seconds, err := strconv.Atoi(duration); err == nil && seconds > 0 {
			duration = strconv.Itoa(seconds)
		}
		if item.BVID == "" || item.Title == "" {
			continue
		}
		videoURL := "https://www.bilibili.com/video/" + item.BVID
		if item.ArcURL != "" {
			videoURL = strings.Replace(item.ArcURL, "http://", "https://", 1)
		}
		results = append(results, BilibiliResult{
			BVid:     item.BVID,
			Title:    cleanBilibiliTitle(item.Title),
			Uploader: item.Author,
			Duration: duration,
			URL:      videoURL,
		})
		if len(results) >= limit {
			break
		}
	}
	return results, nil
}

var htmlTagRE = regexp.MustCompile(`<[^>]+>`)

func fetchBilibiliSearchBody(apiURL string) ([]byte, error) {
	req, err := http.NewRequest(http.MethodGet, apiURL, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("User-Agent", "Mozilla/5.0")
	req.Header.Set("Referer", "https://search.bilibili.com/")
	client := httputil.NewClient(httputil.DefaultTimeout)
	resp, err := client.Do(req)
	if err == nil {
		defer resp.Body.Close()
		body, readErr := io.ReadAll(io.LimitReader(resp.Body, 4<<20))
		if resp.StatusCode == http.StatusOK && readErr == nil {
			return body, nil
		}
		if resp.StatusCode != http.StatusPreconditionFailed {
			return nil, fmt.Errorf("B站搜索 API 返回 %d", resp.StatusCode)
		}
	}

	// B站偶尔会按 Go HTTP/TLS 指纹返回 412；curl 在同机可正常访问，作为 agent 调用兜底。
	cmd := exec.Command("curl", "-fsSL", "-A", "Mozilla/5.0", "-e", "https://search.bilibili.com/", apiURL)
	cmd.Env = os.Environ()
	out, curlErr := cmd.Output()
	if curlErr != nil {
		if err != nil {
			return nil, fmt.Errorf("B站搜索请求失败: %w", err)
		}
		return nil, fmt.Errorf("B站搜索 API 返回 412，curl 兜底也失败: %w", curlErr)
	}
	return out, nil
}

func cleanBilibiliTitle(s string) string {
	s = htmlTagRE.ReplaceAllString(s, "")
	s = strings.ReplaceAll(s, "\\u003c", "<")
	s = strings.ReplaceAll(s, "\\u003e", ">")
	return strings.TrimSpace(html.UnescapeString(s))
}

// BilibiliResult B站视频结果
type BilibiliResult struct {
	BVid     string `json:"bv_id"`
	Title    string `json:"title"`
	Uploader string `json:"uploader"`
	Duration string `json:"duration"`
	URL      string `json:"url"`
}

// BilibiliDownload 用 yt-dlp 从B站下载音频为 mp3
func BilibiliDownload(bvURL, destDir, filename string) (string, error) {
	return bilibiliDownload(bvURL, destDir, filename, os.Stdout, os.Stderr)
}

// BilibiliDownloadQuiet 用 yt-dlp 下载音频，避免向 stdout 输出进度，便于 agent 解析 JSON。
func BilibiliDownloadQuiet(bvURL, destDir, filename string) (string, error) {
	var stderr bytes.Buffer
	outFile, err := bilibiliDownload(bvURL, destDir, filename, io.Discard, &stderr)
	if err != nil && stderr.Len() > 0 {
		return "", fmt.Errorf("%w: %s", err, strings.TrimSpace(stderr.String()))
	}
	return outFile, err
}

func bilibiliDownload(bvURL, destDir, filename string, stdout, stderr io.Writer) (string, error) {
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
	cmd.Stdout = stdout
	cmd.Stderr = stderr
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
