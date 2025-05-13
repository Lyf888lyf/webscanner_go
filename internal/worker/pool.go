package worker

import (
	"sync"
)

// 扫描任务结构体
type ScanTask struct {
	URL   string
	Depth int
}

// 任务结果回调函数
type TaskCallback func(task ScanTask, err error)

// WorkerPool 并发任务池
type WorkerPool struct {
	workerCount int
	taskChan    chan ScanTask
	wg          sync.WaitGroup
	stopChan    chan struct{}
}

// 创建新的 WorkerPool 实例
func NewWorkerPool(workerCount int) *WorkerPool {
	return &WorkerPool{
		workerCount: workerCount,
		taskChan:    make(chan ScanTask),
		stopChan:    make(chan struct{}),
	}
}

// 启动所有 worker 并指定处理逻辑和回调函数
func (wp *WorkerPool) Start(process func(task ScanTask) error, callback TaskCallback) {
	for i := 0; i < wp.workerCount; i++ {
		wp.wg.Add(1)
		go func(workerID int) {
			defer wp.wg.Done()
			for {
				select {
				case task := <-wp.taskChan:
					err := process(task)
					callback(task, err)
				case <-wp.stopChan:
					return
				}
			}
		}(i)
	}
}

// 提交一个任务到队列中
func (wp *WorkerPool) Submit(task ScanTask) {
	wp.taskChan <- task
}

// 等待所有任务完成
func (wp *WorkerPool) Wait() {
	wp.wg.Wait()
}

// 停止所有 worker
func (wp *WorkerPool) Stop() {
	close(wp.stopChan)
}
