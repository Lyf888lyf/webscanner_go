package browser

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"web_scanner_go/internal/logging"
	"web_scanner_go/internal/storage"

	"github.com/playwright-community/playwright-go"
)

type BrowserManager struct {
	pw      *playwright.Playwright
	browser playwright.Browser
}

// 初始化 Playwright 与浏览器
func NewBrowserManager() (*BrowserManager, error) {
	pw, err := playwright.Run()
	if err != nil {
		return nil, fmt.Errorf("could not launch Playwright: %w", err)
	}
	browser, err := pw.Chromium.Launch(playwright.BrowserTypeLaunchOptions{
		Headless: playwright.Bool(true),
	})
	if err != nil {
		return nil, fmt.Errorf("could not launch browser: %w", err)
	}
	return &BrowserManager{pw: pw, browser: browser}, nil
}

// 打开页面、监听请求、分析资源、保存结果
func (bm *BrowserManager) OpenPageWithLogger(url string, logger *logging.Logger) error {
	context, err := bm.browser.NewContext(playwright.BrowserNewContextOptions{
		BypassCSP: playwright.Bool(true),
	})
	if err != nil {
		return err
	}
	page, err := context.NewPage()
	if err != nil {
		return err
	}

	var apiRequests, jsRequests []string
	getCount, postCount := 0, 0

	page.On("request", func(request playwright.Request) {
		reqURL := request.URL()
		method := request.Method()
		resourceType := request.ResourceType()

		if method == "POST" {
			postCount++
			if postData, err := request.PostData(); err == nil && postData != "" {
				os.MkdirAll("results/post_data", os.ModePerm)
				filePath := filepath.Join("results/post_data", sanitizeFullURL(reqURL)+".txt")
				_ = os.WriteFile(filePath, []byte(postData), 0644)
			}
		} else if method == "GET" {
			getCount++
		}

		// 输出带缩进的资源信息
		if resourceType == "xhr" || resourceType == "fetch" || strings.Contains(reqURL, "/api/") {
			logger.Info("  🔄 [API] %s [%s]", reqURL, method)
			apiRequests = append(apiRequests, reqURL)
		}
		if resourceType == "script" || strings.HasSuffix(reqURL, ".js") ||
			strings.Contains(reqURL, ".js?") || strings.Contains(reqURL, "javascript") {
			logger.Info("  📦 [JS] %s", reqURL)
			jsRequests = append(jsRequests, reqURL)
		}
	})

	logger.Info("Opening page: %s", url)
	response, err := page.Goto(url)
	if err != nil {
		return err
	}

	state := playwright.LoadState("networkidle")
	if err = page.WaitForLoadState(playwright.PageWaitForLoadStateOptions{State: &state}); err != nil {
		return err
	}
	page.WaitForTimeout(3000)

	title, _ := page.Title()
	statusCode := 0
	if response != nil {
		statusCode = response.Status()
	}

	timestamp := time.Now().Format(time.RFC3339)
	logger.Info("  📊 Collected %d API, %d JS", len(apiRequests), len(jsRequests))

	db, err := storage.InitDB("./results/scan.db")
	if err != nil {
		return fmt.Errorf("DB init failed: %w", err)
	}
	defer db.DB.Close()

	result := storage.ScanResult{
		URL:       url,
		Title:     title,
		Timestamp: timestamp,
		Code:      statusCode,
		APIList:   apiRequests,
		JSList:    jsRequests,
		GetCount:  getCount,
		PostCount: postCount,
	}
	if err := storage.SaveScanResult(db, result); err != nil {
		return fmt.Errorf("failed to save result: %w", err)
	}

	logger.Success("📁 Result saved to SQLite.")
	return nil
}

// 关闭浏览器和 Playwright
func (bm *BrowserManager) Close() {
	bm.browser.Close()
	bm.pw.Stop()
}

// 将 URL 转为文件安全名
func sanitizeFullURL(input string) string {
	return strings.NewReplacer("/", "_", ":", "_", "?", "_", "&", "_", "=", "_").Replace(input)
}
