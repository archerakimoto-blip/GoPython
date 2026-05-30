package concurrency

import (
	"sync"

	"github.com/go-py/go-python/pkg/objects"
)

// GoroutineState 表示 GoPy 协程的状态
type GoroutineState int

const (
	GoroutineIdle GoroutineState = iota
	GoroutineRunning
	GoroutineWaiting
	GoroutineSuspended
	GoroutineFinished
)

// Goroutine 表示 GoPy 中的一个协程
type Goroutine struct {
	ID        uint64
	State     GoroutineState
	Stack     []objects.Object
	StackPtr  int
	Closure   *objects.Closure
	IP        int // instruction pointer
	Result    objects.Object
	Error     error
	WaitChan  chan struct{} // 用于等待协程完成
	Ctx       *GoroutineContext
	Parent    *Goroutine // 父协程
	Children  []*Goroutine
	mu        sync.Mutex
}

// GoroutineContext 协程的上下文数据
type GoroutineContext struct {
	Globals     map[string]objects.Object
	Locals      map[string]objects.Object
	Modules     map[string]*objects.Module
}

// Channel 协程间通信通道
type Channel struct {
	ID        uint64
	Buffer    []objects.Object
	Capacity  int
	Closed    bool
	mu        sync.Mutex
	sendCond  *sync.Cond
	recvCond  *sync.Cond
}

// WaitGroup 等待一组协程完成
type WaitGroup struct {
	count int64
	done  chan struct{}
	mu    sync.Mutex
}

// Mutex 互斥锁
type Mutex struct {
	mu sync.Mutex
}

// Once 确保只执行一次
type Once struct {
	once sync.Once
}

// AtomicInt32 原子整数
type AtomicInt32 struct {
	val int32
}

// AtomicInt64 原子整数
type AtomicInt64 struct {
	val int64
}

// Pool 对象池
type Pool struct {
	pool sync.Pool
}
