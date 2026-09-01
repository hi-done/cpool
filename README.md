# cpool

[![Go Reference](https://pkg.go.dev/badge/github.com/hi-done/cpool.svg)](https://pkg.go.dev/github.com/hi-done/cpool)
[![Go Version](https://img.shields.io/badge/Go-%3E%3D1.26.5-blue.svg)](https://go.dev/)
[![License](https://img.shields.io/badge/license-MIT-green.svg)](LICENSE)

一个轻量级的 Go 协程池（goroutine pool）库，用于管理并发任务执行，支持优雅关闭和 panic 恢复。

## 特性

- 🚀 **并发控制** — 限制同时运行的 goroutine 数量，避免协程无限制增长
- 🛡️ **Panic 恢复** — 内置 panic 捕获，单个任务崩溃不影响其他任务和池本身
- ⏱️ **超时支持** — 通过 `SubmitWithContext` 支持任务提交超时和取消
- 🔒 **优雅关闭** — `Release()` 等待所有已提交任务完成后安全退出
- ⚙️ **灵活配置** — 通过 Functional Options 模式配置 worker 数量、队列大小、panic 处理器

## 安装

```bash
go get github.com/hi-done/cpool
```

## 快速上手

```go
package main

import (
    "fmt"
    "github.com/hi-done/cpool"
)

func main() {
    // 创建协程池：5 个 worker，队列容量 20
    pool := cpool.NewPool(
        cpool.WithWorkerNum(5),
        cpool.WithQueueSize(20),
    )
    defer pool.Release()

    // 提交任务
    pool.Submit(func() {
        fmt.Println("hello from goroutine pool")
    })
}
```

## API 参考

### 创建池

```go
pool := cpool.NewPool(opts ...Option)
```

| 选项 | 默认值 | 说明 |
|------|--------|------|
| `WithWorkerNum(n)` | 10 | 设置 worker 协程数量 |
| `WithQueueSize(n)` | 100 | 设置任务队列缓冲大小 |
| `WithPanicHandler(h)` | 内置 handler | 自定义 panic 处理函数 |

### 提交任务

```go
// 阻塞式提交，直到有空闲 worker 接收任务
err := pool.Submit(func() {
    // your task logic
})

// 带超时/取消的提交
err := pool.SubmitWithContext(ctx, func() {
    // your task logic
})
```

### 关闭池

```go
// 优雅关闭：停止接收新任务，等待所有已提交任务执行完毕
pool.Release()
```

## 示例

### 并发任务处理

```go
pool := cpool.NewPool(cpool.WithWorkerNum(10))
defer pool.Release()

var wg sync.WaitGroup
for i := 0; i < 100; i++ {
    wg.Add(1)
    i := i
    pool.Submit(func() {
        defer wg.Done()
        // 处理任务 i
        fmt.Printf("task %d done\n", i)
    })
}
wg.Wait()
```

### 自定义 Panic 处理

```go
pool := cpool.NewPool(
    cpool.WithWorkerNum(4),
    cpool.WithPanicHandler(func(r any) {
        log.Printf("recovered from panic: %v", r)
    }),
)
defer pool.Release()
```

### 带超时的任务提交

```go
ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
defer cancel()

err := pool.SubmitWithContext(ctx, func() {
    // 耗时操作
})
if err != nil {
    fmt.Println("submit failed:", err)
}
```

## 开源许可

MIT © 2026 hi-done
