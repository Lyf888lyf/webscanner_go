package configs

import (
	"encoding/json"
	"fmt"
	"os"
)

type AppConfig struct {
	ThreadCount   int    `json:"thread_count"`
	MaxScanCount  int    `json:"max_scan_count"`
	Depth         int    `json:"depth"`
	URLListPath   string `json:"url_list"`
	Verbose       bool   `json:"verbose"`
	Resume        bool   `json:"resume"`
	ClearResume   bool   `json:"clear_resume"`
	ResultDBPath  string `json:"result_db"`
	ServerPort    int    `json:"server_port"`
	ScreenshotDir string `json:"screenshot_dir"`
}

// 从配置文件加载配置
func LoadConfig(path string) (*AppConfig, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, fmt.Errorf("打开配置文件失败: %w", err)
	}
	defer file.Close()

	decoder := json.NewDecoder(file)
	var config AppConfig
	if err := decoder.Decode(&config); err != nil {
		return nil, fmt.Errorf("解析配置文件失败: %w", err)
	}

	return &config, nil
}
