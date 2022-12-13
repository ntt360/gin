package pool

import (
	"context"
	"sync"
	"sync/atomic"
	"time"
)

// pool default config params
const (
	DefaultPoolSize    = 1               // 线程池默认线程数
	DefaultIdleTimeout = time.Second * 5 // 如果5秒线程未被重复使用，则立即回收线程
)

// Options default option struct
type Options struct {
	Size        int           // 线程池容量大小
	IdleTimeout time.Duration // 线程池线程idle时间，超过idle时间
}

// Task pool single task
type Task struct {
	cb func() // Task执行完回调函数
}

// Pool pool struct
type Pool struct {
	sync.Mutex
	Size        int32         // 最大worker发数据
	cur         int32         // 当前worker数量
	IdleTimeout time.Duration // 线程闲置时间
	taskChan    chan Task
	killSig     chan bool // kill 进程
	ctx         context.Context
	ctxCancel   context.CancelFunc
	close       bool
}

// Close the pool
func (p *Pool) Close() {
	p.Lock()
	defer p.Unlock()
	if p.close {
		return
	}
	p.close = true
	p.ctxCancel()

	// 关掉通道
	close(p.taskChan)
	close(p.killSig)
}

func (p *Pool) allocWorker(num int) {
	for i := 0; i < num; i++ {
		if atomic.AddInt32(&p.cur, 1) <= p.Size {
			go p.workerExec()
		}
	}
}

// Submit one task to pool
func (p *Pool) Submit(callbackFunc func()) {
	t := Task{
		cb: callbackFunc,
	}
	p.reallocWorker()

	p.taskChan <- t

}

func (p *Pool) workerExec() {
	timer := time.NewTimer(p.IdleTimeout)
	for {
		select {
		case t := <-p.taskChan: // 如果当前线程获取到任务
			if t.cb != nil {
				t.cb()
			}
			timer.Reset(p.IdleTimeout)
		case <-timer.C: // 线程到期未使用直接退出
			p.workerExist()
			timer.Stop()
			return
		case <-p.ctx.Done(): // 线程收到主线程的close信号
			p.workerExist()
			timer.Stop()
			return
		}
	}
}

func (p *Pool) workerExist() {
	atomic.AddInt32(&p.cur, -1)
}

// 重新检测分配 goroutine
func (p *Pool) reallocWorker() {
	if p.cur < p.Size {
		p.allocWorker(1)
	}
}

// NewPool create
func NewPool(options *Options) *Pool {
	pool := &Pool{
		Size:        int32(options.Size),
		IdleTimeout: options.IdleTimeout,
	}

	if pool.Size <= 0 {
		pool.Size = DefaultPoolSize
	}

	if pool.IdleTimeout < time.Second {
		pool.IdleTimeout = DefaultIdleTimeout
	}

	pool.ctx, pool.ctxCancel = context.WithCancel(context.Background())

	pool.taskChan = make(chan Task, pool.Size)
	pool.killSig = make(chan bool)

	pool.allocWorker(int(pool.Size))
	return pool
}
