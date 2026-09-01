package cpool

import (
	"context"
	"fmt"
	"sync"

	"github.com/hi-done/clog"
)

// Task 是提交到池中的任务
type Task func()

// PanicHandler 用于处理 worker 中的 panic
type PanicHandler func(any)

// Pool 协程池本体
type Pool struct {
	workerNum    int
	queueSize    int
	tasks        chan Task
	stopChan     chan struct{}
	wg           sync.WaitGroup
	panicHandler PanicHandler
	closed       bool
	mu           sync.Mutex // 保护 closed 状态
}

// NewPool 创建一个新的协程池
func NewPool(opts ...Option) *Pool {
	p := &Pool{
		workerNum:    10,
		queueSize:    100,
		panicHandler: defaultPanicHandler,
	}

	for _, opt := range opts {
		opt(p)
	}

	p.tasks = make(chan Task, p.queueSize)
	p.stopChan = make(chan struct{})

	// 启动 worker
	p.wg.Add(p.workerNum)
	for i := 0; i < p.workerNum; i++ {
		go p.worker(i)
	}

	return p
}

// Submit 提交任务（阻塞式）
func (p *Pool) Submit(task Task) error {
	if task == nil {
		return fmt.Errorf("task cannot be nil")
	}

	p.mu.Lock()
	if p.closed {
		p.mu.Unlock()
		return fmt.Errorf("pool is closed")
	}
	p.mu.Unlock()

	// 阻塞直到有空闲 worker 或 pool 被关闭
	select {
	case p.tasks <- task:
		return nil
	case <-p.stopChan:
		return fmt.Errorf("pool is closing")
	}
}

// SubmitWithContext 带上下文的提交（支持超时/取消）
func (p *Pool) SubmitWithContext(ctx context.Context, task Task) error {
	if task == nil {
		return fmt.Errorf("task cannot be nil")
	}

	p.mu.Lock()
	if p.closed {
		p.mu.Unlock()
		return fmt.Errorf("pool is closed")
	}
	p.mu.Unlock()

	select {
	case p.tasks <- task:
		return nil
	case <-p.stopChan:
		return fmt.Errorf("pool is closing")
	case <-ctx.Done():
		return ctx.Err()
	}
}

// Release 优雅关闭协程池
func (p *Pool) Release() {
	p.mu.Lock()
	if p.closed {
		p.mu.Unlock()
		return
	}
	p.closed = true
	p.mu.Unlock()

	// 关闭 stopChan，通知所有 worker 不再接收新任务
	close(p.stopChan)

	// 关闭 tasks channel，让 worker 处理完剩余任务后退出
	close(p.tasks)

	// 等待所有 worker 退出
	p.wg.Wait()
}

// worker 是真正干活的协程
func (p *Pool) worker(id int) {
	defer p.wg.Done()

	for task := range p.tasks {
		func() {
			defer func() {
				if r := recover(); r != nil {
					p.panicHandler(r)
				}
			}()
			task()
		}()
	}
}

// defaultPanicHandler 默认 panic 处理
func defaultPanicHandler(r any) {
	clog.Errorf("[POOL PANIC] %v\n", r)
}
