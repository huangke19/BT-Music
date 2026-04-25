package main

import (
	"bufio"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/huangke/bt-music/search"
)

const version = "0.2.0"

type cliConfig struct {
	jsonOutput bool
	limit      int
	pick       int
	destDir    string
}

type jsonResponse struct {
	OK       bool              `json:"ok"`
	Version  string            `json:"version"`
	Command  string            `json:"command,omitempty"`
	Keyword  string            `json:"keyword,omitempty"`
	Count    int               `json:"count"`
	Results  any               `json:"results,omitempty"`
	Download *downloadResult   `json:"download,omitempty"`
	Error    *jsonError        `json:"error,omitempty"`
	Meta     map[string]string `json:"meta,omitempty"`
}

type jsonError struct {
	Message string `json:"message"`
}

type downloadResult struct {
	Title   string `json:"title,omitempty"`
	URL     string `json:"url"`
	OutFile string `json:"out_file,omitempty"`
	OutDir  string `json:"out_dir"`
}

func defaultDownloadDir() string {
	home, _ := os.UserHomeDir()
	return filepath.Join(home, "Downloads", "BT-Music")
}

func main() {
	cfg, command, args, err := parseArgs(os.Args[1:])
	if err != nil {
		exitWithError(cfg, err)
	}
	if command == "" && !isTerminal(os.Stdin) {
		exitWithError(cfg, errors.New("请提供命令；Hermes agent 调用建议使用 --json music、--json bt 或 --json download"))
	}
	if command == "" || command == "interactive" {
		runInteractive(cfg.destDir)
		return
	}

	if err := runCommand(cfg, command, args); err != nil {
		exitWithError(cfg, err)
	}
}

func parseArgs(args []string) (cliConfig, string, []string, error) {
	cfg := cliConfig{
		limit:   10,
		pick:    1,
		destDir: defaultDownloadDir(),
	}
	fs := flag.NewFlagSet("bt-music", flag.ContinueOnError)
	fs.SetOutput(os.Stderr)
	fs.BoolVar(&cfg.jsonOutput, "json", false, "output machine-readable JSON")
	fs.IntVar(&cfg.limit, "limit", 10, "maximum results to return")
	fs.IntVar(&cfg.pick, "pick", 1, "result number to download for get command, 1-based")
	fs.StringVar(&cfg.destDir, "output-dir", cfg.destDir, "download directory")
	fs.StringVar(&cfg.destDir, "dir", cfg.destDir, "download directory")
	fs.Usage = printUsage
	if err := fs.Parse(args); err != nil {
		return cfg, "", nil, err
	}
	rest := fs.Args()
	if len(rest) == 0 {
		return cfg, "", nil, nil
	}
	command := strings.ToLower(rest[0])
	if !isCommand(command) && len(rest) == 1 {
		cfg.destDir = rest[0]
		return cfg, "", nil, nil
	}
	if !isCommand(command) {
		return cfg, "", nil, fmt.Errorf("未知命令: %s", rest[0])
	}
	if cfg.limit <= 0 {
		cfg.limit = 10
	}
	if cfg.pick <= 0 {
		cfg.pick = 1
	}
	return cfg, command, rest[1:], nil
}

func isCommand(command string) bool {
	switch command {
	case "music", "bili", "bilibili", "bt", "download", "download-bili", "get", "interactive", "help", "version":
		return true
	default:
		return false
	}
}

func runCommand(cfg cliConfig, command string, args []string) error {
	switch command {
	case "help":
		if cfg.jsonOutput {
			return writeJSON(jsonResponse{
				OK:      true,
				Version: version,
				Command: command,
				Meta:    map[string]string{"usage": usageText()},
			})
		}
		printUsage()
		return nil
	case "version":
		if cfg.jsonOutput {
			return writeJSON(jsonResponse{OK: true, Version: version, Command: command})
		}
		fmt.Println(version)
		return nil
	case "music", "bili", "bilibili":
		keyword := strings.TrimSpace(strings.Join(args, " "))
		if keyword == "" {
			return errors.New("请输入 B 站搜索关键词")
		}
		results, err := search.BilibiliSearch(keyword, cfg.limit)
		if err != nil {
			return err
		}
		if cfg.jsonOutput {
			return writeJSON(jsonResponse{
				OK:      true,
				Version: version,
				Command: "music",
				Keyword: keyword,
				Count:   len(results),
				Results: results,
			})
		}
		printBilibiliResults(results)
		return nil
	case "bt":
		keyword := strings.TrimSpace(strings.Join(args, " "))
		if keyword == "" {
			return errors.New("请输入 BT 搜索关键词")
		}
		results, err := search.Search(keyword, search.DefaultProviders())
		if err != nil {
			return err
		}
		if len(results) > cfg.limit {
			results = results[:cfg.limit]
		}
		if cfg.jsonOutput {
			return writeJSON(jsonResponse{
				OK:      true,
				Version: version,
				Command: "bt",
				Keyword: keyword,
				Count:   len(results),
				Results: results,
			})
		}
		printBTResults(results)
		return nil
	case "get":
		keyword := strings.TrimSpace(strings.Join(args, " "))
		if keyword == "" {
			return errors.New("请输入 B 站搜索关键词")
		}
		results, err := search.BilibiliSearch(keyword, cfg.limit)
		if err != nil {
			return err
		}
		if len(results) == 0 {
			return errors.New("未找到相关 B站视频")
		}
		if cfg.pick < 1 || cfg.pick > len(results) {
			return fmt.Errorf("--pick 超出范围: %d，可选 1-%d", cfg.pick, len(results))
		}
		r := results[cfg.pick-1]
		return downloadBilibili(cfg, "get", r.URL, r.Title)
	case "download", "download-bili":
		if len(args) == 0 || strings.TrimSpace(args[0]) == "" {
			return errors.New("请输入 B站视频 URL 或 BV 号")
		}
		rawURL := strings.TrimSpace(args[0])
		title := strings.TrimSpace(strings.Join(args[1:], " "))
		return downloadBilibili(cfg, "download", normalizeBilibiliURL(rawURL), title)
	default:
		return fmt.Errorf("未知命令: %s", command)
	}
}

func downloadBilibili(cfg cliConfig, command, url, title string) error {
	filename := sanitizeFilename(title)
	var outFile string
	var err error
	if cfg.jsonOutput {
		outFile, err = search.BilibiliDownloadQuiet(url, cfg.destDir, filename)
	} else {
		outFile, err = search.BilibiliDownload(url, cfg.destDir, filename)
	}
	if err != nil {
		return err
	}
	if cfg.jsonOutput {
		return writeJSON(jsonResponse{
			OK:      true,
			Version: version,
			Command: command,
			Download: &downloadResult{
				Title:   title,
				URL:     url,
				OutFile: outFile,
				OutDir:  cfg.destDir,
			},
		})
	}
	if outFile != "" {
		fmt.Printf("保存到: %s\n", outFile)
	} else {
		fmt.Printf("已下载到: %s/\n", cfg.destDir)
	}
	return nil
}

func runInteractive(destDir string) {
	fmt.Printf("BT-Music v%s\n", version)
	fmt.Println("命令: music <关键词>  搜B站并下载 | bt <关键词>  BT搜索获取磁力 | quit 退出")
	fmt.Println()
	fmt.Printf("下载目录: %s\n\n", destDir)

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
			fmt.Println("再见!")
			return

		case strings.HasPrefix(strings.ToLower(input), "music "):
			keyword := strings.TrimSpace(input[6:])
			if keyword == "" {
				fmt.Println("请输入歌手或歌名")
				continue
			}
			fmt.Printf("搜索B站: %s\n", keyword)
			results, err := search.BilibiliSearch(keyword, 10)
			if err != nil {
				fmt.Printf("搜索失败: %v\n", err)
				continue
			}
			if len(results) == 0 {
				fmt.Println("未找到相关视频，可以用 bt <关键词> 尝试BT搜索")
				continue
			}
			lastBili = results
			lastBT = nil
			printBilibiliResults(results)
			fmt.Println("\n输入序号下载 mp3（回车跳过）: ")

		case strings.HasPrefix(strings.ToLower(input), "bt "):
			keyword := strings.TrimSpace(input[3:])
			if keyword == "" {
				fmt.Println("请输入搜索关键词（建议英文）")
				continue
			}
			fmt.Printf("BT搜索: %s\n", keyword)
			results, err := search.Search(keyword, search.DefaultProviders())
			if err != nil {
				fmt.Printf("搜索失败: %v\n", err)
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
			printBTResults(results[:limit])
			fmt.Println("\n输入序号复制磁力链接（回车跳过）: ")

		default:
			num, err := strconv.Atoi(input)
			if err != nil {
				fmt.Println("未知命令。输入 music <关键词> 搜B站，bt <关键词> 搜BT，或输入序号选择")
				continue
			}

			if len(lastBili) > 0 {
				if num < 1 || num > len(lastBili) {
					fmt.Println("序号超出范围")
					continue
				}
				r := lastBili[num-1]
				filename := sanitizeFilename(r.Title)
				fmt.Printf("下载: %s\n    %s\n", r.Title, r.URL)
				outFile, err := search.BilibiliDownload(r.URL, destDir, filename)
				if err != nil {
					fmt.Fprintf(os.Stderr, "下载失败: %v\n", err)
				} else if outFile != "" {
					fmt.Printf("保存到: %s\n", outFile)
				} else {
					fmt.Printf("已下载到: %s/\n", destDir)
				}
				fmt.Println()

			} else if len(lastBT) > 0 {
				if num < 1 || num > len(lastBT) {
					fmt.Println("序号超出范围")
					continue
				}
				r := lastBT[num-1]
				fmt.Printf("磁力链接 [%s]:\n%s\n\n", r.Name, r.Magnet)

			} else {
				fmt.Println("请先用 music 或 bt 命令搜索")
			}
		}
	}
}

func printBilibiliResults(results []search.BilibiliResult) {
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
}

func printBTResults(results []search.Result) {
	fmt.Printf("\n找到 %d 个结果（按做种数排序）:\n\n", len(results))
	for i, r := range results {
		fmt.Printf("  [%d] %s\n      %s | Seeders: %d | %s\n",
			i+1, r.Name, r.Size, r.Seeders, r.Source)
	}
}

func writeJSON(resp jsonResponse) error {
	enc := json.NewEncoder(os.Stdout)
	enc.SetIndent("", "  ")
	return enc.Encode(resp)
}

func exitWithError(cfg cliConfig, err error) {
	if cfg.jsonOutput {
		_ = writeJSON(jsonResponse{
			OK:      false,
			Version: version,
			Error:   &jsonError{Message: err.Error()},
		})
	} else {
		fmt.Fprintf(os.Stderr, "错误: %v\n", err)
	}
	os.Exit(1)
}

func normalizeBilibiliURL(s string) string {
	if strings.HasPrefix(s, "http://") || strings.HasPrefix(s, "https://") {
		return s
	}
	return "https://www.bilibili.com/video/" + s
}

func isTerminal(f *os.File) bool {
	info, err := f.Stat()
	if err != nil {
		return false
	}
	return info.Mode()&os.ModeCharDevice != 0
}

func printUsage() {
	fmt.Fprint(os.Stderr, usageText())
}

func usageText() string {
	return fmt.Sprintf(`BT-Music v%s

Usage:
  bt-music [flags] music <keyword...>
  bt-music [flags] bt <keyword...>
  bt-music [flags] download <bilibili-url-or-bvid> [filename...]
  bt-music [flags] get <keyword...>
  bt-music [output-dir]

Flags:
  --json            output machine-readable JSON
  --limit N         maximum results to return, default 10
  --pick N          result number to download for get command, default 1
  --output-dir DIR  download directory, default ~/Downloads/BT-Music

Agent examples:
  bt-music --json --limit 5 music "周杰伦 以父之名"
  bt-music --json --limit 10 bt "Pink Floyd The Wall FLAC"
  bt-music --json --output-dir /tmp/music download BV1xx411c7mD "song title"
  bt-music --json --pick 2 --output-dir /tmp/music get "周杰伦 以父之名"

`, version)
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
