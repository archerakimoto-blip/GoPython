package concurrency

import (
	"context"
	"runtime"
	"sync"
	"sync/atomic"
	"time"

	"github.com/go-py/go-python/pkg/objects"
)

// Scheduler 协程调度器
type Scheduler struct {
	// 全局协程队列
	readyQueue chan *Goroutine
	// 所有协程池
	goroutines map[uint64]*Goroutine
	// 锁
	mu sync.Mutex
	// 工作线程
	workers []*worker
	// 工作线程数
	workerCount int
	// 关闭标志
	stopped atomic.Bool
	// 等待所有工作线程完成
	wg sync.WaitGroup
	// 上下文
	ctx context.Context
	// 取消函数
	cancel context.CancelFunc
}

type worker struct {
	id        int
	scheduler *Scheduler
}

var (
	globalScheduler *Scheduler
	schedulerOnce   sync.Once
	goroutineIDCounter uint64
)

// GetScheduler 获取全局调度器
func GetScheduler() *Scheduler {
	schedulerOnce.Do(func() {
		globalScheduler = NewScheduler(runtime.GOMAXPROCS(0))
	})
	return globalScheduler
}

// NewScheduler 创建新的调度器
func NewScheduler(workerCount int) *Scheduler {
	if workerCount <= 0 {
		workerCount = runtime.GOMAXPROCS(0)
	}
	ctx, cancel := context.WithCancel(context.Background())
	
	s := &Scheduler{
		readyQueue:  make(chan *Goroutine, 10000),
		goroutines: make(map[uint64]*Goroutine),
		workerCount: workerCount,
		ctx:        ctx,
		cancel:      cancel,
	}
	
	// 启动工作线程
	s.startWorkers()
	
	return s
}

// startWorkers 启动工作线程
func (s *Scheduler) startWorkers() {
	s.workers = make([]*worker, s.workerCount)
	for i := 0; i < s.workerCount; i++ {
		w := &worker{
			id:        i,
			scheduler: s,
		}
		s.workers[i] = w
		s.wg.Add(1)
		go w.run()
	}
}

// run 工作线程主循环
func (w *worker) run() {
	defer w.scheduler.wg.Done()
	
	for {
		select {
		case <-w.scheduler.ctx.Done():
			return
		case goroutine := <-w.scheduler.readyQueue:
			w.executeGoroutine(goroutine)
		}
	}
}

// executeGoroutine 执行协程（这里只是模拟执行
func (w *worker) executeGoroutine(g *Goroutine) {
	g.mu.Lock()
	if g.State != GoroutineIdle && g.State != GoroutineSuspended {
		g.mu.Unlock()
		return
	}
	g.State = GoroutineRunning
	g.mu.Unlock()
	
	// 这里应该执行协程的实际代码
	// 我们将在 VM 中集成实际执行逻辑
	
	// 模拟执行完成
	g.mu.Lock()
	g.State = GoroutineFinished
	g.mu.Unlock()
	
	if g.WaitChan != nil {
		close(g.WaitChan)
	}
	
	// 通知父协程
	if g.Parent != nil {
		// 这里可以添加通知逻辑
	}
}

// Go 创建并启动一个新协程
func (s *Scheduler) Go(fn func() objects.Object, parent *Goroutine) *Goroutine {
	if s.stopped.Load() {
		return nil
	}
	
	g := &Goroutine{
		ID:       atomic.AddUint64(&goroutineIDCounter, 1),
		State:    GoroutineIdle,
		Stack:    make([]objects.Object, 0, 128),
		StackPtr: 0,
		WaitChan: make(chan struct{}),
		Ctx: &GoroutineContext{
			Globals: make(map[string]objects.Object),
			Locals:  make(map[string]objects.Object),
			Modules: make(map[string]*objects.Module),
		},
		Parent:   parent,
		Children: make([]*Goroutine, 0),
	}
	
	// 添加到调度器
	s.mu.Lock()
	s.goroutines[g.ID] = g
	s.mu.Unlock()
	
	// 添加到父协程
	if parent != nil {
		parent.mu.Lock()
		parent.Children = append(parent.Children, g)
		parent.mu.Unlock()
	}
	
	// 入队准备执行
	s.readyQueue <- g
	
	return g
}

// GoWithClosure 使用闭包创建协程
func (s *Scheduler) GoWithClosure(closure *objects.Closure, parent *Goroutine) *Goroutine {
	g := s.Go(func() objects.Object { return nil }, parent)
	g.Closure = closure
	return g
}

// Wait 等待协程完成
func (s *Scheduler) Wait(g *Goroutine) (objects.Object, error) {
	if g.WaitChan != nil {
		<-g.WaitChan
	}
	return g.Result, g.Error
}

// Stop 停止调度器
func (s *Scheduler) Stop() {
	if s.stopped.Swap(true) {
		return // 已经停止
	}
	s.cancel()
	s.wg.Wait()
}

// NumGoroutine 返回当前活跃的协程数
func (s *Scheduler) NumGoroutine() int {
	s.mu.Lock()
	defer s.mu.Unlock()
	return len(s.goroutines)
}

// GetGoroutine 通过 ID 获取协程
func (s *Scheduler) GetGoroutine(id uint64) (*Goroutine, bool) {
	s.mu.Lock()
	defer s.mu.Unlock()
	g, ok := s.goroutines[id]
	return g, ok
}

// Sleep 让当前协程休眠一段时间
func (s *Scheduler) Sleep(duration time.Duration) {
	time.Sleep(duration)
}

// NewWaitGroup 创建新的 WaitGroup
func NewWaitGroup() *WaitGroup {
	return &WaitGroup{
		done: make(chan struct{}),
	}
}

// Add 添加计数
func (wg *WaitGroup) Add(delta int) {
	wg.mu.Lock()
	defer wg.mu.Unlock()
	atomic.AddInt64(&wg.count, int64(delta))
}

// Done 减少计数
func (wg *WaitGroup) Done() {
	wg.mu.Lock()
	defer wg.mu.Unlock()
	
	newCount := atomic.AddInt64(&wg.count, -1)
	if newCount == 0 {
		close(wg.done)
	}
}

// Wait 等待计数为 0
func (wg *WaitGroup) Wait() {
	<-wg.done
}

// Lock 加锁
func (m *Mutex) Lock() {
	m.mu.Lock()
}

// Unlock 解锁
func (m *Mutex) Unlock() {
	m.mu.Unlock()
}

// Do 执行一次
func (o *Once) Do(f func()) {
	o.once.Do(f)
}

// NewAtomicInt32 创建原子整数
func NewAtomicInt32(initial int32) *AtomicInt32 {
	return &AtomicInt32{val: initial}
}

// Load 加载值
func (a *AtomicInt32) Load() int32 {
	return atomic.LoadInt32(&a.val)
}

// Store 存储值
func (a *AtomicInt32) Store(val int32) {
	atomic.StoreInt32(&a.val, val)
}

// Add 增加值
func (a *AtomicInt32) Add(delta int32) int32 {
	return atomic.AddInt32(&a.val, delta)
}

// NewAtomicInt64 创建原子整数
func NewAtomicInt64(initial int64) *AtomicInt64 {
	return &AtomicInt64{val: initial}
}

// Load 加载值
func (a *AtomicInt64) Load() int64 {
	return atomic.LoadInt64(&a.val)
}

// Store 存储值
func (a *AtomicInt64) Store(val int64) {
	atomic.StoreInt64(&a.val, val)
}

// Add 增加值
func (a *AtomicInt64) Add(delta int64) int64 {
	return atomic.AddInt64(&a.val, delta)
}

// NewPool 创建对象池
func NewPool(newFn func() interface{}) *Pool {
	return &Pool{
		pool: sync.Pool{
			New: newFn,
		},
	}
}

// Get 从池中获取对象
func (p *Pool) Get() interface{} {
	return p.pool.Get()
}

// Put 将对象放回池中
func (p *Pool) Put(x interface{}) {
	p.pool.Put(x)
}
