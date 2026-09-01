package cpool

// Option 配置函数类型
type Option func(*Pool)

// WithWorkerNum 设置 worker 数量
func WithWorkerNum(n int) Option {
	return func(p *Pool) {
		if n > 0 {
			p.workerNum = n
		}
	}
}

// WithQueueSize 设置任务队列大小
func WithQueueSize(n int) Option {
	return func(p *Pool) {
		if n > 0 {
			p.queueSize = n
		}
	}
}

// WithPanicHandler 设置自定义 panic 处理
func WithPanicHandler(handler PanicHandler) Option {
	return func(p *Pool) {
		if handler != nil {
			p.panicHandler = handler
		}
	}
}
