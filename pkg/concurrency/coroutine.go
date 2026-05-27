package concurrency

import (
	"sync"
	"time"

	"github.com/go-py/go-python/pkg/compiler"
	"github.com/go-py/go-python/pkg/objects"
	"github.com/go-py/go-python/pkg/vm"
)

type CoroutineState int

const (
	CoroutineReady CoroutineState = iota
	CoroutineRunning
	CoroutineBlocked
	CoroutineFinished
)

type Coroutine struct {
	id          uint64
	code        *compiler.Bytecode
	vm          *vm.VM
	state       CoroutineState
	joinChan    chan struct{}
	scheduler   *Scheduler
	stack       []objects.Object
	programCounter int
	mu          sync.Mutex
}

type Scheduler struct {
	runningCoroutines map[uint64]*Coroutine
	readyQueue        []*Coroutine
	mu                sync.Mutex
	nextID            uint64
	running           bool
	workers           int
	quitChan          chan struct{}
}

var schedulerInstance *Scheduler
var schedulerOnce sync.Once

func GetScheduler() *Scheduler {
	schedulerOnce.Do(func() {
		schedulerInstance = &Scheduler{
			runningCoroutines: make(map[uint64]*Coroutine),
			readyQueue:        make([]*Coroutine, 0),
			workers:           4,
			quitChan:          make(chan struct{}),
		}
	})
	return schedulerInstance
}

func (s *Scheduler) Start() {
	s.mu.Lock()
	if s.running {
		s.mu.Unlock()
		return
	}
	s.running = true
	s.mu.Unlock()

	for i := 0; i < s.workers; i++ {
		go s.worker()
	}
}

func (s *Scheduler) Stop() {
	s.mu.Lock()
	if !s.running {
		s.mu.Unlock()
		return
	}
	s.running = false
	s.mu.Unlock()

	close(s.quitChan)
	time.Sleep(100 * time.Millisecond)
}

func (s *Scheduler) worker() {
	for {
		select {
		case <-s.quitChan:
			return
		default:
			s.mu.Lock()
			if len(s.readyQueue) == 0 {
				s.mu.Unlock()
				time.Sleep(1 * time.Millisecond)
				continue
			}

			coroutine := s.readyQueue[0]
			s.readyQueue = s.readyQueue[1:]
			coroutine.state = CoroutineRunning
			s.mu.Unlock()

			coroutine.run()

			s.mu.Lock()
			coroutine.state = CoroutineFinished
			delete(s.runningCoroutines, coroutine.id)
			close(coroutine.joinChan)
			s.mu.Unlock()
		}
	}
}

func (s *Scheduler) Spawn(code *compiler.Bytecode) *Coroutine {
	s.mu.Lock()
	defer s.mu.Unlock()

	coroutine := &Coroutine{
		id:          s.nextID,
		code:        code,
		vm:          vm.New(code),
		state:       CoroutineReady,
		joinChan:    make(chan struct{}),
		scheduler:   s,
		stack:       make([]objects.Object, 0),
		programCounter: 0,
	}

	s.nextID++
	s.runningCoroutines[coroutine.id] = coroutine
	s.readyQueue = append(s.readyQueue, coroutine)

	return coroutine
}

func (s *Scheduler) Yield(coroutine *Coroutine) {
	s.mu.Lock()
	coroutine.state = CoroutineReady
	s.readyQueue = append(s.readyQueue, coroutine)
	s.mu.Unlock()
}

func (c *Coroutine) run() {
	err := c.vm.Run()
	if err != nil {
	}
}

func (c *Coroutine) Join() {
	<-c.joinChan
}

func (c *Coroutine) State() CoroutineState {
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.state
}

func (c *Coroutine) ID() uint64 {
	return c.id
}

func Spawn(code *compiler.Bytecode) *Coroutine {
	return GetScheduler().Spawn(code)
}

func Yield() {
}

func Join(c *Coroutine) {
	c.Join()
}
