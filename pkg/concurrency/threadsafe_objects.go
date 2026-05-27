package concurrency

import (
	"sync"
	"sync/atomic"
	"unsafe"

	"github.com/go-py/go-python/pkg/objects"
)

type Lock struct {
	mu    sync.Mutex
	owner uint64
	count int32
}

func NewLock() *Lock {
	return &Lock{
		count: 0,
	}
}

func (l *Lock) Acquire() {
	l.mu.Lock()
	l.owner = getThreadID()
	l.count++
}

func (l *Lock) Release() {
	if l.owner != getThreadID() {
		return
	}
	l.count--
	if l.count == 0 {
		l.owner = 0
		l.mu.Unlock()
	}
}

func (l *Lock) TryAcquire() bool {
	if l.mu.TryLock() {
		l.owner = getThreadID()
		l.count = 1
		return true
	}
	return false
}

type RWMutex struct {
	mu        sync.RWMutex
	readers   int32
	writeOwner uint64
}

func NewRWMutex() *RWMutex {
	return &RWMutex{
		readers: 0,
	}
}

func (rw *RWMutex) Lock() {
	rw.mu.Lock()
	rw.writeOwner = getThreadID()
}

func (rw *RWMutex) Unlock() {
	if rw.writeOwner != getThreadID() {
		return
	}
	rw.writeOwner = 0
	rw.mu.Unlock()
}

func (rw *RWMutex) RLock() {
	rw.mu.RLock()
	atomic.AddInt32(&rw.readers, 1)
}

func (rw *RWMutex) RUnlock() {
	atomic.AddInt32(&rw.readers, -1)
	rw.mu.RUnlock()
}

type ThreadSafeList struct {
	elements []objects.Object
	mu       sync.RWMutex
}

func NewThreadSafeList() *ThreadSafeList {
	return &ThreadSafeList{
		elements: make([]objects.Object, 0),
	}
}

func (l *ThreadSafeList) Append(obj objects.Object) {
	l.mu.Lock()
	l.elements = append(l.elements, obj)
	l.mu.Unlock()
}

func (l *ThreadSafeList) Get(index int) (objects.Object, bool) {
	l.mu.RLock()
	defer l.mu.RUnlock()
	if index < 0 || index >= len(l.elements) {
		return nil, false
	}
	return l.elements[index], true
}

func (l *ThreadSafeList) Set(index int, obj objects.Object) bool {
	l.mu.Lock()
	defer l.mu.Unlock()
	if index < 0 || index >= len(l.elements) {
		return false
	}
	l.elements[index] = obj
	return true
}

func (l *ThreadSafeList) Len() int {
	l.mu.RLock()
	defer l.mu.RUnlock()
	return len(l.elements)
}

func (l *ThreadSafeList) Remove(index int) (objects.Object, bool) {
	l.mu.Lock()
	defer l.mu.Unlock()
	if index < 0 || index >= len(l.elements) {
		return nil, false
	}
	obj := l.elements[index]
	l.elements = append(l.elements[:index], l.elements[index+1:]...)
	return obj, true
}

func (l *ThreadSafeList) Iterate() []objects.Object {
	l.mu.RLock()
	defer l.mu.RUnlock()
	result := make([]objects.Object, len(l.elements))
	copy(result, l.elements)
	return result
}

type ThreadSafeDict struct {
	pairs    map[string]objects.Object
	keys     map[string]objects.Object
	mu       sync.RWMutex
}

func NewThreadSafeDict() *ThreadSafeDict {
	return &ThreadSafeDict{
		pairs: make(map[string]objects.Object),
		keys:  make(map[string]objects.Object),
	}
}

func (d *ThreadSafeDict) Set(keyStr string, key objects.Object, value objects.Object) {
	d.mu.Lock()
	d.pairs[keyStr] = value
	d.keys[keyStr] = key
	d.mu.Unlock()
}

func (d *ThreadSafeDict) Get(keyStr string) (objects.Object, bool) {
	d.mu.RLock()
	defer d.mu.RUnlock()
	value, ok := d.pairs[keyStr]
	return value, ok
}

func (d *ThreadSafeDict) Delete(keyStr string) bool {
	d.mu.Lock()
	defer d.mu.Unlock()
	if _, ok := d.pairs[keyStr]; ok {
		delete(d.pairs, keyStr)
		delete(d.keys, keyStr)
		return true
	}
	return false
}

func (d *ThreadSafeDict) Len() int {
	d.mu.RLock()
	defer d.mu.RUnlock()
	return len(d.pairs)
}

func (d *ThreadSafeDict) Keys() []objects.Object {
	d.mu.RLock()
	defer d.mu.RUnlock()
	result := make([]objects.Object, 0, len(d.keys))
	for _, key := range d.keys {
		result = append(result, key)
	}
	return result
}

func (d *ThreadSafeDict) Values() []objects.Object {
	d.mu.RLock()
	defer d.mu.RUnlock()
	result := make([]objects.Object, 0, len(d.pairs))
	for _, value := range d.pairs {
		result = append(result, value)
	}
	return result
}

type AtomicInteger struct {
	value int64
}

func NewAtomicInteger(initial int64) *AtomicInteger {
	return &AtomicInteger{
		value: initial,
	}
}

func (a *AtomicInteger) Get() int64 {
	return atomic.LoadInt64(&a.value)
}

func (a *AtomicInteger) Set(value int64) {
	atomic.StoreInt64(&a.value, value)
}

func (a *AtomicInteger) Add(delta int64) int64 {
	return atomic.AddInt64(&a.value, delta)
}

func (a *AtomicInteger) Increment() int64 {
	return atomic.AddInt64(&a.value, 1)
}

func (a *AtomicInteger) Decrement() int64 {
	return atomic.AddInt64(&a.value, -1)
}

func (a *AtomicInteger) CompareAndSwap(old, new int64) bool {
	return atomic.CompareAndSwapInt64(&a.value, old, new)
}

type AtomicBool struct {
	value int32
}

func NewAtomicBool(initial bool) *AtomicBool {
	v := int32(0)
	if initial {
		v = 1
	}
	return &AtomicBool{
		value: v,
	}
}

func (a *AtomicBool) Get() bool {
	return atomic.LoadInt32(&a.value) == 1
}

func (a *AtomicBool) Set(value bool) {
	v := int32(0)
	if value {
		v = 1
	}
	atomic.StoreInt32(&a.value, v)
}

func (a *AtomicBool) CompareAndSwap(old, new bool) bool {
	oldV := int32(0)
	newV := int32(0)
	if old {
		oldV = 1
	}
	if new {
		newV = 1
	}
	return atomic.CompareAndSwapInt32(&a.value, oldV, newV)
}

type AtomicPointer struct {
	value uintptr
}

func NewAtomicPointer(initial unsafe.Pointer) *AtomicPointer {
	return &AtomicPointer{
		value: uintptr(initial),
	}
}

func (a *AtomicPointer) Get() unsafe.Pointer {
	return unsafe.Pointer(atomic.LoadUintptr(&a.value))
}

func (a *AtomicPointer) Set(value unsafe.Pointer) {
	atomic.StoreUintptr(&a.value, uintptr(value))
}

func (a *AtomicPointer) CompareAndSwap(old, new unsafe.Pointer) bool {
	return atomic.CompareAndSwapUintptr(&a.value, uintptr(old), uintptr(new))
}

type Condition struct {
	mu      sync.Mutex
	cond    *sync.Cond
	waiters int32
}

func NewCondition() *Condition {
	c := &Condition{}
	c.cond = sync.NewCond(&c.mu)
	return c
}

func (c *Condition) Wait() {
	c.mu.Lock()
	atomic.AddInt32(&c.waiters, 1)
	c.cond.Wait()
	atomic.AddInt32(&c.waiters, -1)
	c.mu.Unlock()
}

func (c *Condition) Signal() {
	c.mu.Lock()
	c.cond.Signal()
	c.mu.Unlock()
}

func (c *Condition) Broadcast() {
	c.mu.Lock()
	c.cond.Broadcast()
	c.mu.Unlock()
}

type Barrier struct {
	count      int
	threshold  int
	mu         sync.Mutex
	cond       *sync.Cond
	generation int
}

func NewBarrier(threshold int) *Barrier {
	b := &Barrier{
		threshold: threshold,
	}
	b.cond = sync.NewCond(&b.mu)
	return b
}

func (b *Barrier) Wait() {
	b.mu.Lock()
	gen := b.generation
	b.count++
	if b.count == b.threshold {
		b.generation++
		b.count = 0
		b.cond.Broadcast()
		b.mu.Unlock()
		return
	}
	for gen == b.generation {
		b.cond.Wait()
	}
	b.mu.Unlock()
}

type Semaphore struct {
	count int32
	mu    sync.Mutex
	cond  *sync.Cond
}

func NewSemaphore(initial int) *Semaphore {
	s := &Semaphore{
		count: int32(initial),
	}
	s.cond = sync.NewCond(&s.mu)
	return s
}

func (s *Semaphore) Acquire() {
	s.mu.Lock()
	for s.count <= 0 {
		s.cond.Wait()
	}
	s.count--
	s.mu.Unlock()
}

func (s *Semaphore) Release() {
	s.mu.Lock()
	s.count++
	s.cond.Signal()
	s.mu.Unlock()
}

func (s *Semaphore) TryAcquire() bool {
	s.mu.Lock()
	if s.count > 0 {
		s.count--
		s.mu.Unlock()
		return true
	}
	s.mu.Unlock()
	return false
}
