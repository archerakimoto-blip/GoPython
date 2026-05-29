package concurrency

import (
	"errors"
	"sync"
	"sync/atomic"

	"github.com/go-py/go-python/pkg/objects"
)

var (
	channelIDCounter uint64
	ErrChannelClosed = errors.New("channel is closed")
	ErrChannelFull   = errors.New("channel is full")
)

// NewChannel 创建新的通道
func NewChannel(capacity int) *Channel {
	ch := &Channel{
		ID:       atomic.AddUint64(&channelIDCounter, 1),
		Buffer:   make([]objects.Object, 0, capacity),
		Capacity: capacity,
		Closed:   false,
	}
	ch.sendCond = sync.NewCond(&ch.mu)
	ch.recvCond = sync.NewCond(&ch.mu)
	return ch
}

// Send 发送数据到通道
func (ch *Channel) Send(val objects.Object) error {
	ch.mu.Lock()
	defer ch.mu.Unlock()

	if ch.Closed {
		return ErrChannelClosed
	}

	// 等待直到有空间
	for len(ch.Buffer) >= ch.Capacity && !ch.Closed {
		ch.sendCond.Wait()
	}

	if ch.Closed {
		return ErrChannelClosed
	}

	ch.Buffer = append(ch.Buffer, val)
	ch.recvCond.Signal() // 唤醒一个等待接收的协程
	return nil
}

// Receive 从通道接收数据
func (ch *Channel) Receive() (objects.Object, bool, error) {
	ch.mu.Lock()
	defer ch.mu.Unlock()

	// 等待直到有数据或通道关闭
	for len(ch.Buffer) == 0 && !ch.Closed {
		ch.recvCond.Wait()
	}

	if len(ch.Buffer) == 0 && ch.Closed {
		return objects.None_, false, nil
	}

	val := ch.Buffer[0]
	ch.Buffer = ch.Buffer[1:]
	ch.sendCond.Signal() // 唤醒一个等待发送的协程
	return val, true, nil
}

// Close 关闭通道
func (ch *Channel) Close() {
	ch.mu.Lock()
	defer ch.mu.Unlock()

	if ch.Closed {
		return
	}

	ch.Closed = true
	ch.sendCond.Broadcast()
	ch.recvCond.Broadcast()
}

// Len 返回通道中的元素数量
func (ch *Channel) Len() int {
	ch.mu.Lock()
	defer ch.mu.Unlock()
	return len(ch.Buffer)
}

// IsClosed 检查通道是否已关闭
func (ch *Channel) IsClosed() bool {
	ch.mu.Lock()
	defer ch.mu.Unlock()
	return ch.Closed
}

// TrySend 尝试发送，不阻塞
func (ch *Channel) TrySend(val objects.Object) (bool, error) {
	ch.mu.Lock()
	defer ch.mu.Unlock()

	if ch.Closed {
		return false, ErrChannelClosed
	}

	if len(ch.Buffer) >= ch.Capacity {
		return false, nil
	}

	ch.Buffer = append(ch.Buffer, val)
	ch.recvCond.Signal()
	return true, nil
}

// TryReceive 尝试接收，不阻塞
func (ch *Channel) TryReceive() (objects.Object, bool, bool) {
	ch.mu.Lock()
	defer ch.mu.Unlock()

	if len(ch.Buffer) == 0 {
		return objects.None_, false, ch.Closed
	}

	val := ch.Buffer[0]
	ch.Buffer = ch.Buffer[1:]
	ch.sendCond.Signal()
	return val, true, ch.Closed
}
