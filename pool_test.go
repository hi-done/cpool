package cpool

import (
	"fmt"
	"sync"
	"testing"
	"time"
)

func TestBasicSubmit(t *testing.T) {
	pool := NewPool(WithWorkerNum(3), WithQueueSize(10))
	defer pool.Release()

	var wg sync.WaitGroup
	count := 5

	wg.Add(count)
	for i := 0; i < count; i++ {
		i := i
		_ = pool.Submit(func() {
			defer wg.Done()
			fmt.Printf("task %d done\n", i)
			time.Sleep(10 * time.Millisecond)
		})
	}

	wg.Wait()
}

func TestPanicRecovery(t *testing.T) {
	pool := NewPool(WithWorkerNum(2))
	defer pool.Release()

	_ = pool.Submit(func() {
		panic("something went wrong")
	})

	// 给 worker 一点时间执行
	time.Sleep(50 * time.Millisecond)
	// 如果走到这里没崩，说明 recover 生效
}

func TestSubmitAfterClose(t *testing.T) {
	pool := NewPool(WithWorkerNum(2))
	pool.Release()

	err := pool.Submit(func() {
		fmt.Println("should not run")
	})

	if err == nil {
		t.Fatal("expected error after close, got nil")
	}
}

func TestGracefulShutdown(t *testing.T) {
	pool := NewPool(WithWorkerNum(3), WithQueueSize(5))

	var wg sync.WaitGroup
	for i := 0; i < 5; i++ {
		wg.Add(1)
		_ = pool.Submit(func() {
			defer wg.Done()
			time.Sleep(30 * time.Millisecond)
		})
	}

	// 等待所有已提交任务执行完毕，再优雅关闭
	wg.Wait()
	pool.Release()

	// Release 后确认不再接受新任务
	err := pool.Submit(func() {})
	if err == nil {
		t.Fatal("expected error, pool should be closing")
	}
}
