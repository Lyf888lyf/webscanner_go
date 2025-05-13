package controller

import (
	"fmt"
	"time"

	"web_scanner_go/configs"
	"web_scanner_go/internal/browser"
	"web_scanner_go/internal/logging"
	"web_scanner_go/internal/scheduler"
	"web_scanner_go/internal/storage"
	"web_scanner_go/internal/utils"
	"web_scanner_go/internal/worker"
)

type Controller struct {
	config *configs.AppConfig
	queue  *scheduler.TaskQueue
	pool   *worker.WorkerPool
	db     *storage.DBWrapper
	logger *logging.Logger
}

// 初始化控制器
func NewController(cfg *configs.AppConfig) *Controller {
	db, err := storage.InitDB(cfg.ResultDBPath)
	if err != nil {
		panic("❌ 无法初始化数据库: " + err.Error())
	}
	return &Controller{
		config: cfg,
		queue:  scheduler.NewTaskQueue(),
		pool:   worker.NewWorkerPool(cfg.ThreadCount),
		db:     db,
		logger: logging.NewLogger(cfg.Verbose),
	}
}

// 启动主任务流程
func (c *Controller) Start(urls []string) {
	// 清空历史记录
	if c.config.ClearResume {
		_ = storage.ClearAllScans(c.db)
		fmt.Println("✅ 已清空历史扫描记录")
	}

	// 初始化任务队列
	c.queue.EnqueueBatch(urls, c.config.Depth)

	// 启动并发任务池
	c.pool.Start(c.handleTask, c.handleResult)

	// 消费任务队列
	for {
		task, ok := c.queue.Dequeue()
		if !ok {
			break
		}
		// 跳过已扫描
		if c.config.Resume {
			exists, _ := storage.IsURLScanned(c.db, task.URL)
			if exists {
				fmt.Printf("🔁 跳过已扫描：%s\n", task.URL)
				continue
			}
		}
		c.pool.Submit(worker.ScanTask{URL: task.URL, Depth: task.Depth})
	}

	// 等待所有任务完成
	c.pool.Wait()
}

// 扫描任务处理逻辑
func (c *Controller) handleTask(task worker.ScanTask) error {
	logger := logging.NewLogger(c.config.Verbose)

	bm, err := browser.NewBrowserManager()
	if err != nil {
		logger.Error("浏览器初始化失败: %v", err)
		return err
	}
	defer bm.Close()

	// 打开网页并记录日志与资源
	if err := bm.OpenPageWithLogger(task.URL, logger); err != nil {
		return err
	}

	// 递归抓取子链接
	if task.Depth > 1 {
		links := utils.ExtractLinks(task.URL) // ⚠️ 你需要实现这个函数
		for _, link := range links {
			if !c.queue.Seen(link) {
				c.queue.Enqueue(scheduler.ScanTask{
					URL:   link,
					Depth: task.Depth - 1,
				})
			}
		}
	}

	time.Sleep(300 * time.Millisecond) // 限速，防止被封
	return nil
}

// 任务完成后的回调函数
func (c *Controller) handleResult(task worker.ScanTask, err error) {
	if err != nil {
		c.logger.Error("❌ 任务失败 %s：%v", task.URL, err)
	} else {
		c.logger.Success("✅ 扫描成功：%s", task.URL)
	}
}
