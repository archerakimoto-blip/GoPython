package concurrency

import (
	"sync"
	"time"

	"github.com/go-py/go-python/pkg/objects"
)

type Channel struct {
	buffer     []objects.Object
	capacity   int
	mu         sync.Mutex
	readers    []chan objects.Object
	writers    []chan objects.Object
	closed     bool
	closedChan chan struct{}
}

func NewChannel(capacity int) *Channel {
	return &Channel{
		buffer:     make([]objects.Object, 0, capacity),
		capacity:   capacity,
		readers:    make([]chan objects.Object, 0),
		writers:    make([]chan objects.Object, 0),
		closed:     false,
		closedChan: make(chan struct{}),
	}
}

func (ch *Channel) Send(obj objects.Object) bool {
	ch.mu.Lock()

	if ch.closed {
		ch.mu.Unlock()
		return false
	}

	if len(ch.readers) > 0 {
		reader := ch.readers[0]
		ch.readers = ch.readers[1:]
		ch.mu.Unlock()
		reader <- obj
		return true
	}

	if ch.capacity > 0 && len(ch.buffer) < ch.capacity {
		ch.buffer = append(ch.buffer, obj)
		ch.mu.Unlock()
		return true
	}

	if ch.capacity == 0 {
		ch.mu.Unlock()
		select {
		case reader := <-ch.getReader():
			reader <- obj
			return true
		case <-ch.closedChan:
			return false
		}
	}

	writeChan := make(chan objects.Object, 1)
	ch.writers = append(ch.writers, writeChan)
	ch.mu.Unlock()

	select {
	case writeChan <- obj:
		return true
	case <-ch.closedChan:
		return false
	}
}

func (ch *Channel) Receive() (objects.Object, bool) {
	ch.mu.Lock()

	if ch.closed && len(ch.buffer) == 0 {
		ch.mu.Unlock()
		return nil, false
	}

	if len(ch.buffer) > 0 {
		obj := ch.buffer[0]
		ch.buffer = ch.buffer[1:]

		if len(ch.writers) > 0 {
			writer := ch.writers[0]
			ch.writers = ch.writers[1:]
			ch.mu.Unlock()
			go func() {
				obj := <-writer
				ch.mu.Lock()
				ch.buffer = append(ch.buffer, obj)
				ch.mu.Unlock()
			}()
		} else {
			ch.mu.Unlock()
		}
		return obj, true
	}

	if len(ch.writers) > 0 {
		writer := ch.writers[0]
		ch.writers = ch.writers[1:]
		ch.mu.Unlock()
		obj := <-writer
		return obj, true
	}

	readChan := make(chan objects.Object, 1)
	ch.readers = append(ch.readers, readChan)
	ch.mu.Unlock()

	select {
	case obj := <-readChan:
		return obj, true
	case <-ch.closedChan:
		return nil, false
	}
}

func (ch *Channel) getReader() <-chan objects.Object {
	ch.mu.Lock()
	defer ch.mu.Unlock()

	if len(ch.readers) > 0 {
		reader := ch.readers[0]
		ch.readers = ch.readers[1:]
		return reader
	}

	readChan := make(chan objects.Object, 1)
	ch.readers = append(ch.readers, readChan)
	return readChan
}

func (ch *Channel) Close() {
	ch.mu.Lock()
	if ch.closed {
		ch.mu.Unlock()
		return
	}
	ch.closed = true
	close(ch.closedChan)

	for _, reader := range ch.readers {
		close(reader)
	}
	for _, writer := range ch.writers {
		close(writer)
	}
	ch.mu.Unlock()
}

func (ch *Channel) Len() int {
	ch.mu.Lock()
	defer ch.mu.Unlock()
	return len(ch.buffer)
}

func (ch *Channel) Cap() int {
	return ch.capacity
}

func (ch *Channel) Closed() bool {
	ch.mu.Lock()
	defer ch.mu.Unlock()
	return ch.closed
}

type SelectCase struct {
	channel *Channel
	isSend  bool
	value   objects.Object
	result  chan selectResult
}

type selectResult struct {
	success bool
	value   objects.Object
	index   int
}

func Select(cases []SelectCase) (int, objects.Object, bool) {
	if len(cases) == 0 {
		return -1, nil, false
	}

	resultChan := make(chan selectResult, len(cases))
	done := make(chan struct{})

	for i, c := range cases {
		go func(index int, sc SelectCase) {
			select {
			case <-done:
				return
			default:
				if sc.isSend {
					success := sc.channel.Send(sc.value)
					if success {
						select {
						case resultChan <- selectResult{success: true, value: sc.value, index: index}:
						case <-done:
						}
					}
				} else {
					value, success := sc.channel.Receive()
					if success {
						select {
						case resultChan <- selectResult{success: true, value: value, index: index}:
						case <-done:
						}
					}
				}
			}
		}(i, c)
	}

	timeout := time.After(5 * time.Minute)
	select {
	case result := <-resultChan:
		close(done)
		time.Sleep(10 * time.Millisecond)
		return result.index, result.value, result.success
	case <-timeout:
		close(done)
		return -1, nil, false
	}
}

func MakeChannel(capacity int) *Channel {
	return NewChannel(capacity)
}

func ChanSend(ch *Channel, obj objects.Object) bool {
	return ch.Send(obj)
}

func ChanReceive(ch *Channel) (objects.Object, bool) {
	return ch.Receive()
}

func ChanClose(ch *Channel) {
	ch.Close()
}
