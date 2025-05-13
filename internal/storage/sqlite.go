package storage

import (
	"database/sql"
	"fmt"

	_ "github.com/mattn/go-sqlite3"
)

// 封装数据库连接，便于后续扩展（如缓存、中间件等）
type DBWrapper struct {
	DB *sql.DB
}

type ScanResult struct {
	URL       string
	Title     string
	Timestamp string
	Code      int
	GetCount  int
	PostCount int
	APIList   []string
	JSList    []string
}

// 初始化数据库并建表，返回封装后的 DBWrapper
func InitDB(path string) (*DBWrapper, error) {
	db, err := sql.Open("sqlite3", path)
	if err != nil {
		return nil, err
	}
	db.Exec("PRAGMA foreign_keys = ON;")

	fmt.Printf("📂 当前数据库路径: %s\n", path)

	schema := []string{
		`CREATE TABLE IF NOT EXISTS scans (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			url TEXT NOT NULL UNIQUE,
			timestamp TEXT NOT NULL,
			title TEXT,
			code INTEGER,
			api_count INTEGER DEFAULT 0,
			js_count INTEGER DEFAULT 0,
			get_count INTEGER DEFAULT 0,
			post_count INTEGER DEFAULT 0
		);`,
		`CREATE TABLE IF NOT EXISTS api_requests (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			scan_id INTEGER,
			api_url TEXT UNIQUE,
			FOREIGN KEY (scan_id) REFERENCES scans(id)
		);`,
		`CREATE TABLE IF NOT EXISTS js_files (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			scan_id INTEGER,
			js_url TEXT UNIQUE,
			FOREIGN KEY (scan_id) REFERENCES scans(id)
		);`,
	}

	for _, stmt := range schema {
		if _, err := db.Exec(stmt); err != nil {
			return nil, fmt.Errorf("创建表失败: %w", err)
		}
	}

	return &DBWrapper{DB: db}, nil
}

// 保存扫描结果
func SaveScanResult(db *DBWrapper, result ScanResult) error {
	tx, err := db.DB.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	res, err := tx.Exec(
		`INSERT OR IGNORE INTO scans (url, timestamp, title, code, api_count, js_count, get_count, post_count)
         VALUES (?, ?, ?, ?, ?, ?, ?, ?)`,
		result.URL, result.Timestamp, result.Title, result.Code,
		len(result.APIList), len(result.JSList), result.GetCount, result.PostCount,
	)
	if err != nil {
		return err
	}

	scanID, _ := res.LastInsertId()
	for _, api := range result.APIList {
		_, _ = tx.Exec(`INSERT OR IGNORE INTO api_requests (scan_id, api_url) VALUES (?, ?)`, scanID, api)
	}
	for _, js := range result.JSList {
		_, _ = tx.Exec(`INSERT OR IGNORE INTO js_files (scan_id, js_url) VALUES (?, ?)`, scanID, js)
	}

	return tx.Commit()
}

// 判断某 URL 是否已扫描
func IsURLScanned(db *DBWrapper, url string) (bool, error) {
	var exists bool
	err := db.DB.QueryRow("SELECT EXISTS(SELECT 1 FROM scans WHERE url = ?)", url).Scan(&exists)
	return exists, err
}

// 清空所有历史扫描记录（按外键顺序删除）
func ClearAllScans(db *DBWrapper) error {
	tx, err := db.DB.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	if _, err := tx.Exec("DELETE FROM api_requests"); err != nil {
		return fmt.Errorf("清空 api_requests 失败: %w", err)
	}
	if _, err := tx.Exec("DELETE FROM js_files"); err != nil {
		return fmt.Errorf("清空 js_files 失败: %w", err)
	}
	if _, err := tx.Exec("DELETE FROM scans"); err != nil {
		return fmt.Errorf("清空 scans 失败: %w", err)
	}

	return tx.Commit()
}
