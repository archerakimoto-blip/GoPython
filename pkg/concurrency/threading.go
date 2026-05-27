package concurrency

import (
	"sync"
	"sync/atomic"
	"time"

	"github.com/go-py/go-python/pkg/compiler"
	"github.com/go-py/go-python/pkg/vm"
)

type ThreadState int

const (
	ThreadRunning ThreadState = iota
	ThreadBlocked
	ThreadFinished
	ThreadWaiting
)

type Thread struct {
	id        uint64
	code      *compiler.Bytecode
	vm        *vm.VM
	state     ThreadState
	joinChan  chan struct{}
	mu        sync.Mutex
}

type ThreadPool struct {
	threads     map[uint64]*Thread
	mu          sync.Mutex
	nextID      uint64
	activeCount int32
	maxWorkers  int
}

var threadPoolInstance *ThreadPool
var threadPoolOnce sync.Once

func GetThreadPool() *ThreadPool {
	threadPoolOnce.Do(func() {
		threadPoolInstance = &ThreadPool{
			threads:    make(map[uint64]*Thread),
			maxWorkers: 8,
		}
	})
	return threadPoolInstance
}

func (tp *ThreadPool) Submit(code *compiler.Bytecode) *Thread {
	tp.mu.Lock()
	if atomic.LoadInt32(&tp.activeCount) >= int32(tp.maxWorkers) {
		tp.mu.Unlock()
		time.Sleep(1 * time.Millisecond)
		return tp.Submit(code)
	}

	thread := &Thread{
		id:       tp.nextID,
		code:     code,
		vm:       vm.New(code),
		state:    ThreadRunning,
		joinChan: make(chan struct{}),
	}

	tp.nextID++
	tp.threads[thread.id] = thread
	atomic.AddInt32(&tp.activeCount, 1)
	tp.mu.Unlock()

	go tp.runThread(thread)

	return thread
}

func (tp *ThreadPool) runThread(thread *Thread) {
	GIL.Acquire()
	err := thread.vm.Run()
	GIL.Release()

	tp.mu.Lock()
	thread.state = ThreadFinished
	delete(tp.threads, thread.id)
	atomic.AddInt32(&tp.activeCount, -1)
	close(thread.joinChan)
	tp.mu.Unlock()

	if err != nil {
	}
}

func (tp *ThreadPool) SetMaxWorkers(max int) {
	tp.mu.Lock()
	tp.maxWorkers = max
	tp.mu.Unlock()
}

func (tp *ThreadPool) ActiveCount() int {
	return int(atomic.LoadInt32(&tp.activeCount))
}

func (t *Thread) Join() {
	<-t.joinChan
}

func (t *Thread) State() ThreadState {
	t.mu.Lock()
	defer t.mu.Unlock()
	return t.state
}

func (t *Thread) ID() uint64 {
	return t.id
}

type GIL struct {
	mu        sync.Mutex
	owner     uint64
	count     int32
	waiters   int32
	waitChan  chan struct{}
}

var GILInstance *GIL
var GILOnce sync.Once

func initGIL() {
	GILOnce.Do(func() {
		GILInstance = &GIL{
			owner:    0,
			count:    0,
			waiters:  0,
			waitChan: make(chan struct{}, 100),
		}
	})
}

var GIL = &GILWrapper{}

type GILWrapper struct{}

func (g *GILWrapper) Acquire() {
	initGIL()
	GILInstance.acquire()
}

func (g *GILWrapper) Release() {
	initGIL()
	GILInstance.release()
}

func (g *GIL) acquire() {
	g.mu.Lock()

	if g.count == 0 {
		g.owner = getThreadID()
		g.count = 1
		g.mu.Unlock()
		return
	}

	if g.owner == getThreadID() {
		g.count++
		g.mu.Unlock()
		return
	}

	atomic.AddInt32(&g.waiters, 1)
	g.mu.Unlock()

	<-g.waitChan

	g.mu.Lock()
	g.owner = getThreadID()
	g.count = 1
	atomic.AddInt32(&g.waiters, -1)
	g.mu.Unlock()
}

func (g *GIL) release() {
	g.mu.Lock()

	if g.owner != getThreadID() {
		g.mu.Unlock()
		return
	}

	g.count--

	if g.count == 0 {
		g.owner = 0
		if atomic.LoadInt32(&g.waiters) > 0 {
			select {
			case g.waitChan <- struct{}{}:
			default:
			}
		}
	}

	g.mu.Unlock()
}

func getThreadID() uint64 {
	return uint64(time.Now().UnixNano()) % 1000000
}

func SpawnThread(code *compiler.Bytecode) *Thread {
	return GetThreadPool().Submit(code)
}

func ThreadJoin(t *Thread) {
	t.Join()
}

func LockGIL() {
	GIL.Acquire()
}

func UnlockGIL() {
	GIL.Release()
}
