package scheduler

import (
	"sync"
)

type ScanTask struct {
	URL   string
	Depth int
}

type TaskQueue struct {
	queue []ScanTask
	seen  map[string]bool
	lock  sync.Mutex
}

// 创建新任务队列
func NewTaskQueue() *TaskQueue {
	return &TaskQueue{
		queue: []ScanTask{},
		seen:  make(map[string]bool),
	}
}

// 入队任务，自动去重
func (q *TaskQueue) Enqueue(task ScanTask) {
	q.lock.Lock()
	defer q.lock.Unlock()
	if !q.seen[task.URL] {
		q.queue = append(q.queue, task)
		q.seen[task.URL] = true
	}
}

// 出队任务
func (q *TaskQueue) Dequeue() (ScanTask, bool) {
	q.lock.Lock()
	defer q.lock.Unlock()
	if len(q.queue) == 0 {
		return ScanTask{}, false
	}
	task := q.queue[0]
	q.queue = q.queue[1:]
	return task, true
}

// 队列是否为空
func (q *TaskQueue) IsEmpty() bool {
	q.lock.Lock()
	defer q.lock.Unlock()
	return len(q.queue) == 0
}

// 判断是否已处理/入队过
func (q *TaskQueue) Seen(url string) bool {
	q.lock.Lock()
	defer q.lock.Unlock()
	return q.seen[url]
}

// 批量入队（带深度）
func (q *TaskQueue) EnqueueBatch(urls []string, depth int) {
	for _, url := range urls {
		q.Enqueue(ScanTask{URL: url, Depth: depth})
	}
}
