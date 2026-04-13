package search

import (
	"encoding/xml"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"strings"

	"github.com/huangke/bt-music/pkg/httputil"
)

// Nyaa 动漫/音乐资源丰富
type Nyaa struct {
	client *http.Client
}

func NewNyaa() *Nyaa {
	return &Nyaa{client: httputil.NewClient(httputil.DefaultTimeout)}
}

func (n *Nyaa) Name() string { return "Nyaa" }

type nyaaRSS struct {
	Channel struct {
		Items []struct {
			Title    string `xml:"title"`
			Link     string `xml:"link"`
			Seeders  string `xml:"seeders"`
			Leechers string `xml:"leechers"`
			InfoHash string `xml:"infoHash"`
			Size     string `xml:"size"`
		} `xml:"item"`
	} `xml:"channel"`
}

func (n *Nyaa) Search(keyword string, page int) ([]Result, error) {
	apiURL := fmt.Sprintf("https://nyaa.si/?page=rss&q=%s&c=2_0&f=0",
		url.QueryEscape(keyword))
	req, err := http.NewRequest(http.MethodGet, apiURL, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("User-Agent", httputil.DefaultUA)
	resp, err := n.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	body, err := io.ReadAll(io.LimitReader(resp.Body, 2<<20))
	if err != nil {
		return nil, err
	}
	var rss nyaaRSS
	if err := xml.Unmarshal(body, &rss); err != nil {
		return nil, err
	}
	var results []Result
	for _, item := range rss.Channel.Items {
		seeders, _ := strconv.Atoi(strings.TrimSpace(item.Seeders))
		leechers, _ := strconv.Atoi(strings.TrimSpace(item.Leechers))
		result := Result{
			Name:     item.Title,
			Size:     item.Size,
			Seeders:  seeders,
			Leechers: leechers,
			InfoHash: item.InfoHash,
			Source:   n.Name(),
			Magnet:   item.Link,
		}
		if result.Magnet == "" && item.InfoHash != "" {
			result.Magnet = BuildMagnet(item.InfoHash, url.QueryEscape(item.Title))
		}
		results = append(results, result)
	}
	return results, nil
}
