package main

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/huangke/bt-music/search"
)

const version = "0.1.0"

func defaultDownloadDir() string {
	home, _ := os.UserHomeDir()
	return filepath.Join(home, "Downloads", "BT-Music")
}

func main() {
	fmt.Printf("🎵 BT-Music v%s\n", version)
	fmt.Println("命令: music <关键词>  搜B站并下载 | bt <关键词>  BT搜索获取磁力 | quit 退出")
	fmt.Println()

	destDir := defaultDownloadDir()
	// 支持命令行参数指定下载目录
	if len(os.Args) > 1 {
		destDir = os.Args[1]
	}
	fmt.Printf("💾 下载目录: %s\n\n", destDir)

	scanner := bufio.NewScanner(os.Stdin)
	var lastBT []search.Result
	var lastBili []search.BilibiliResult

	for {
		fmt.Print("bt-music> ")
		if !scanner.Scan() {
			break
		}
		input := strings.TrimSpace(scanner.Text())
		if input == "" {
			continue
		}

		switch {
		case input == "quit" || input == "exit" || input == "q":
			fmt.Println("👋 再见!")
			return

		case strings.HasPrefix(strings.ToLower(input), "music "):
			keyword := strings.TrimSpace(input[6:])
			if keyword == "" {
				fmt.Println("⚠️  请输入歌手或歌名")
				continue
			}
			fmt.Printf("🔍 搜索B站: %s\n", keyword)
			results, err := search.BilibiliSearch(keyword, 10)
			if err != nil {
				fmt.Printf("❌ 搜索失败: %v\n", err)
				continue
			}
			if len(results) == 0 {
				fmt.Println("未找到相关视频，可以用 bt <关键词> 尝试BT搜索")
				continue
			}
			lastBili = results
			lastBT = nil
			fmt.Printf("\n找到 %d 个结果:\n\n", len(results))
			for i, r := range results {
				dur := r.Duration
				if dur != "" {
					dur = " | " + dur + "s"
				}
				up := r.Uploader
				if up != "" {
					up = " | UP: " + up
				}
				fmt.Printf("  [%d] %s%s%s\n", i+1, r.Title, dur, up)
			}
			fmt.Println("\n输入序号下载 mp3（回车跳过）: ")

		case strings.HasPrefix(strings.ToLower(input), "bt "):
			keyword := strings.TrimSpace(input[3:])
			if keyword == "" {
				fmt.Println("⚠️  请输入搜索关键词（建议英文）")
				continue
			}
			fmt.Printf("🔍 BT搜索: %s\n", keyword)
			results, err := search.Search(keyword, search.DefaultProviders())
			if err != nil {
				fmt.Printf("❌ 搜索失败: %v\n", err)
				continue
			}
			if len(results) == 0 {
				fmt.Println("未找到有做种的结果")
				continue
			}
			lastBT = results
			lastBili = nil
			limit := 20
			if len(results) < limit {
				limit = len(results)
			}
			fmt.Printf("\n找到 %d 个结果（按做种数排序）:\n\n", len(results))
			for i, r := range results[:limit] {
				fmt.Printf("  [%d] %s\n      %s | Seeders: %d | %s\n",
					i+1, r.Name, r.Size, r.Seeders, r.Source)
			}
			fmt.Println("\n输入序号复制磁力链接（回车跳过）: ")

		default:
			// 序号选择
			num, err := strconv.Atoi(input)
			if err != nil {
				fmt.Println("⚠️  未知命令。输入 music <关键词> 搜B站，bt <关键词> 搜BT，或输入序号选择")
				continue
			}

			if len(lastBili) > 0 {
				if num < 1 || num > len(lastBili) {
					fmt.Println("⚠️  序号超出范围")
					continue
				}
				r := lastBili[num-1]
				// 用标题做文件名（清理特殊字符）
				filename := sanitizeFilename(r.Title)
				fmt.Printf("⬇️  下载: %s\n    %s\n", r.Title, r.URL)
				outFile, err := search.BilibiliDownload(r.URL, destDir, filename)
				if err != nil {
					fmt.Fprintf(os.Stderr, "❌ 下载失败: %v\n", err)
				} else {
					if outFile != "" {
						fmt.Printf("✓ 保存到: %s\n", outFile)
					} else {
						fmt.Printf("✓ 已下载到: %s/\n", destDir)
					}
				}
				fmt.Println()

			} else if len(lastBT) > 0 {
				if num < 1 || num > len(lastBT) {
					fmt.Println("⚠️  序号超出范围")
					continue
				}
				r := lastBT[num-1]
				fmt.Printf("🔗 磁力链接 [%s]:\n%s\n\n", r.Name, r.Magnet)

			} else {
				fmt.Println("⚠️  请先用 music 或 bt 命令搜索")
			}
		}
	}
}

// sanitizeFilename 清理文件名中的非法字符
func sanitizeFilename(s string) string {
	replacer := strings.NewReplacer(
		"/", "-", "\\", "-", ":", "-", "*", "-",
		"?", "", "\"", "", "<", "", ">", "", "|", "-",
	)
	s = replacer.Replace(s)
	s = strings.TrimSpace(s)
	if len([]rune(s)) > 80 {
		runes := []rune(s)
		s = string(runes[:80])
	}
	return s
}
