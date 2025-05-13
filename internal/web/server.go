package web

import (
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	"web_scanner_go/internal/storage"
)

type Server struct {
	DB *storage.DBWrapper
}

func NewServer(db *storage.DBWrapper) *Server {
	return &Server{DB: db}
}

func (s *Server) Start(port int) {
	r := gin.Default()

	// 首页：展示已扫描 URL 列表
	r.GET("/", func(c *gin.Context) {
		rows, err := s.DB.DB.Query(`SELECT url, title, code, timestamp FROM scans ORDER BY id DESC`)
		if err != nil {
			c.String(500, "数据库查询失败")
			return
		}
		defer rows.Close()

		var results []gin.H
		for rows.Next() {
			var url, title, timestamp string
			var code int
			rows.Scan(&url, &title, &code, &timestamp)
			results = append(results, gin.H{
				"url":       url,
				"title":     title,
				"code":      code,
				"timestamp": timestamp,
			})
		}

		c.HTML(http.StatusOK, "index.tmpl", gin.H{
			"results": results,
		})
	})

	// 详情页
	r.GET("/detail", func(c *gin.Context) {
		url := c.Query("url")
		if url == "" {
			c.String(400, "缺少 url 参数")
			return
		}

		apiRows, _ := s.DB.DB.Query(`SELECT api_url FROM api_requests WHERE scan_id = (SELECT id FROM scans WHERE url = ?)`, url)
		jsRows, _ := s.DB.DB.Query(`SELECT js_url FROM js_files WHERE scan_id = (SELECT id FROM scans WHERE url = ?)`, url)

		var apis, jss []string
		for apiRows.Next() {
			var api string
			apiRows.Scan(&api)
			apis = append(apis, api)
		}
		for jsRows.Next() {
			var js string
			jsRows.Scan(&js)
			jss = append(jss, js)
		}

		c.HTML(http.StatusOK, "detail.tmpl", gin.H{
			"url":  url,
			"apis": apis,
			"jss":  jss,
		})
	})

	// 加载模板
	r.LoadHTMLGlob("templates/*.tmpl")

	// 启动服务
	r.Run(":" + itoa(port))
}

func itoa(i int) string {
	return fmt.Sprintf("%d", i)
}
