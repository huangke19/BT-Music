package search

import (
	"fmt"
	"sort"
	"strings"
)

// Result BT 搜索结果
type Result struct {
	Name     string
	Size     string
	Seeders  int
	Leechers int
	Magnet   string
	Source   string
	InfoHash string
}

// Provider BT 搜索源接口
type Provider interface {
	Name() string
	Search(keyword string, page int) ([]Result, error)
}

// DefaultProviders 返回默认 BT 搜索源（音乐相关效果较好的源）
func DefaultProviders() []Provider {
	return []Provider{
		NewApiBay(),
		NewBtDig(),
		NewNyaa(),
	}
}

// Search 并发搜索所有源，去重并按做种数排序
func Search(keyword string, providers []Provider) ([]Result, error) {
	type pr struct {
		results []Result
		err     error
	}
	ch := make(chan pr, len(providers))
	for _, p := range providers {
		go func(p Provider) {
			results, err := p.Search(keyword, 0)
			ch <- pr{results, err}
		}(p)
	}
	var all []Result
	var lastErr error
	for range providers {
		r := <-ch
		if r.err != nil {
			lastErr = r.err
			continue
		}
		all = append(all, r.results...)
	}
	if len(all) == 0 && lastErr != nil {
		return nil, fmt.Errorf("所有搜索源失败: %w", lastErr)
	}

	all = dedup(all)

	var seeded []Result
	for _, r := range all {
		if r.Seeders > 0 {
			seeded = append(seeded, r)
		}
	}
	sort.Slice(seeded, func(i, j int) bool {
		return seeded[i].Seeders > seeded[j].Seeders
	})
	return seeded, nil
}

func dedup(results []Result) []Result {
	seen := make(map[string]int)
	var out []Result
	for _, r := range results {
		hash := strings.ToLower(r.InfoHash)
		if hash == "" {
			out = append(out, r)
			continue
		}
		if idx, ok := seen[hash]; ok {
			if r.Seeders > out[idx].Seeders {
				out[idx] = r
			}
		} else {
			seen[hash] = len(out)
			out = append(out, r)
		}
	}
	return out
}

// BuildMagnet 从 info_hash 构建磁力链接
func BuildMagnet(infoHash, name string) string {
	magnet := fmt.Sprintf("magnet:?xt=urn:btih:%s", infoHash)
	if name != "" {
		magnet += "&dn=" + name
	}
	trackers := []string{
		"udp://tracker.opentrackr.org:1337/announce",
		"udp://open.stealth.si:80/announce",
		"udp://tracker.torrent.eu.org:451/announce",
		"udp://tracker.openbittorrent.com:6969/announce",
	}
	for _, tr := range trackers {
		magnet += "&tr=" + tr
	}
	return magnet
}
