package main

import (
	"flag"
	"fmt"
	"os"
	"strings"

	"web_scanner_go/configs"
	"web_scanner_go/internal/controller"
	"web_scanner_go/internal/storage"
	"web_scanner_go/internal/web"
)

func main() {
	// 命令行参数
	configPath := flag.String("config", "config.json", "配置文件路径")
	skipScan := flag.Bool("no-scan", false, "跳过扫描，仅启动 Web 展示")
	flag.Parse()

	// 加载配置文件
	cfg, err := configs.LoadConfig(*configPath)
	if err != nil {
		fmt.Printf("❌ 加载配置失败: %v\n", err)
		os.Exit(1)
	}
	fmt.Println("✅ 配置加载成功")

	// 加载 URL 列表
	urls := loadURLList(cfg.URLListPath)
	if len(urls) == 0 {
		fmt.Println("⚠️ 未发现有效 URL，终止扫描。")
		os.Exit(0)
	}

	// 初始化并执行扫描流程（可选）
	if !*skipScan {
		scanController := controller.NewController(cfg)
		scanController.Start(urls)
	}

	// 启动 Web 服务
	db, err := storage.InitDB(cfg.ResultDBPath)
	if err != nil {
		fmt.Printf("❌ 无法打开数据库: %v\n", err)
		os.Exit(1)
	}
	webServer := web.NewServer(db)
	webServer.Start(cfg.ServerPort)
}

// 读取并去重 URL 列表
func loadURLList(path string) []string {
	data, err := os.ReadFile(path)
	if err != nil {
		fmt.Printf("❌ 无法读取 URL 列表文件: %v\n", err)
		return nil
	}
	lines := strings.Split(string(data), "\n")
	var urls []string
	seen := map[string]bool{}
	for _, line := range lines {
		url := strings.TrimSpace(line)
		if url != "" && !seen[url] {
			urls = append(urls, url)
			seen[url] = true
		}
	}
	return urls
}
