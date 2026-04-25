package search

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"

	"github.com/huangke/bt-music/pkg/httputil"
	"github.com/huangke/bt-music/pkg/utils"
)

type BtDig struct {
	client *http.Client
}

func NewBtDig() *BtDig {
	return &BtDig{client: httputil.NewClient(httputil.DefaultTimeout)}
}

func (b *BtDig) Name() string { return "BtDig" }

func (b *BtDig) Search(keyword string, page int) ([]Result, error) {
	apiURL := fmt.Sprintf("https://btdig.com/search?q=%s&p=%d&order=0",
		url.QueryEscape(keyword), page)
	req, err := http.NewRequest(http.MethodGet, apiURL, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("User-Agent", httputil.DefaultUA)
	req.Header.Set("Accept", "application/json")
	resp, err := b.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("BTDigg 返回 %d", resp.StatusCode)
	}
	body, err := io.ReadAll(io.LimitReader(resp.Body, 2<<20))
	if err != nil {
		return nil, err
	}

	var data struct {
		Results []struct {
			Name     string `json:"name"`
			InfoHash string `json:"info_hash"`
			Size     int64  `json:"size"`
		} `json:"results"`
	}
	if err := json.Unmarshal(body, &data); err != nil {
		return nil, err
	}

	var results []Result
	for _, r := range data.Results {
		result := Result{
			Name:     r.Name,
			Size:     utils.FormatBytes(r.Size),
			InfoHash: r.InfoHash,
			Seeders:  1, // btdig 不提供做种数，给1表示有做种
			Source:   b.Name(),
		}
		result.Magnet = BuildMagnet(r.InfoHash, url.QueryEscape(r.Name))
		results = append(results, result)
	}
	return results, nil
}
