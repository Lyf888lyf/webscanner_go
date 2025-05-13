package utils

import (
	"net/http"
	"net/url"
	"strings"

	"github.com/PuerkitoBio/goquery"
)

// 提取页面中所有的超链接（用于递归扫描）
func ExtractLinks(pageURL string) []string {
	var links []string
	seen := make(map[string]bool)

	// 发起请求
	resp, err := http.Get(pageURL)
	if err != nil {
		return links
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return links
	}

	// 加载 HTML
	doc, err := goquery.NewDocumentFromReader(resp.Body)
	if err != nil {
		return links
	}

	// 解析页面中所有 <a href="...">
	doc.Find("a").Each(func(i int, s *goquery.Selection) {
		href, exists := s.Attr("href")
		if !exists || href == "" {
			return
		}
		href = strings.TrimSpace(href)
		// 过滤 mailto / javascript / 锚点
		if strings.HasPrefix(href, "mailto:") ||
			strings.HasPrefix(href, "javascript:") ||
			strings.HasPrefix(href, "#") {
			return
		}

		// 解析为绝对路径
		absoluteURL := resolveURL(pageURL, href)
		if absoluteURL != "" && !seen[absoluteURL] {
			seen[absoluteURL] = true
			links = append(links, absoluteURL)
		}
	})

	return links
}

// 将 href 转为绝对 URL
func resolveURL(base string, href string) string {
	baseURL, err := url.Parse(base)
	if err != nil {
		return ""
	}
	hrefURL, err := url.Parse(href)
	if err != nil {
		return ""
	}
	return baseURL.ResolveReference(hrefURL).String()
}
