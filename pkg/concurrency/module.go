package concurrency

import (
	"fmt"
	"time"
	"sync"

	"github.com/go-py/go-python/pkg/objects"
)

// 全局变量，用于存储创建的通道和等待组
var (
	channels      = make(map[uint64]*Channel)
	channelsMutex sync.Mutex
	nextChannelID uint64 = 1

	waitGroups      = make(map[uint64]*WaitGroup)
	waitGroupsMutex sync.Mutex
	nextWaitGroupID uint64 = 1

	mutexes      = make(map[uint64]*Mutex)
	mutexesMutex sync.Mutex
	nextMutexID uint64 = 1
)

// CreateConcurrencyModule 创建并发模块
func CreateConcurrencyModule() *objects.Module {
	module := &objects.Module{
		Name:   "concurrency",
		Fields: make(map[string]objects.Object),
	}
	
	// go 函数 - 启动协程
	module.Fields["go"] = &objects.Builtin{
		Name: "go",
		Fn: func(args ...objects.Object) objects.Object {
			if len(args) < 1 {
				return objects.NewError("go() expects at least 1 argument")
			}
			
			scheduler := GetScheduler()
			g := scheduler.Go(func() objects.Object { return objects.None_ }, nil)
			
			return &objects.Integer{Value: int64(g.ID)}
		},
	}
	
	// channel 函数 - 创建通道
	module.Fields["channel"] = &objects.Builtin{
		Name: "channel",
		Fn: func(args ...objects.Object) objects.Object {
			capacity := 0
			if len(args) >= 1 {
				if i, ok := args[0].(*objects.Integer); ok {
					capacity = int(i.Value)
				}
			}
			
			ch := NewChannel(capacity)
			channelsMutex.Lock()
			chID := nextChannelID
			nextChannelID++
			channels[chID] = ch
			channelsMutex.Unlock()
			
			return &objects.Integer{Value: int64(chID)}
		},
	}
	
	// send 函数 - 发送数据
	module.Fields["send"] = &objects.Builtin{
		Name: "send",
		Fn: func(args ...objects.Object) objects.Object {
			if len(args) < 2 {
				return objects.NewError("send() expects at least 2 arguments")
			}
			
			chIDArg, ok := args[0].(*objects.Integer)
			if !ok {
				return objects.NewTypeError("send() first argument must be a channel ID (integer)")
			}
			
			channelsMutex.Lock()
			ch, ok := channels[uint64(chIDArg.Value)]
			channelsMutex.Unlock()
			
			if !ok {
				return objects.NewError(fmt.Sprintf("channel %d not found", chIDArg.Value))
			}
			
			err := ch.Send(args[1])
			if err != nil {
				return objects.NewError(err.Error())
			}
			
			return objects.None_
		},
	}
	
	// recv 函数 - 接收数据
	module.Fields["recv"] = &objects.Builtin{
		Name: "recv",
		Fn: func(args ...objects.Object) objects.Object {
			if len(args) < 1 {
				return objects.NewError("recv() expects at least 1 argument")
			}
			
			chIDArg, ok := args[0].(*objects.Integer)
			if !ok {
				return objects.NewTypeError("recv() first argument must be a channel ID (integer)")
			}
			
			channelsMutex.Lock()
			ch, ok := channels[uint64(chIDArg.Value)]
			channelsMutex.Unlock()
			
			if !ok {
				return objects.NewError(fmt.Sprintf("channel %d not found", chIDArg.Value))
			}
			
			val, _, err := ch.Receive()
			if err != nil {
				return objects.NewError(err.Error())
			}
			
			return val
		},
	}
	
	// close 函数 - 关闭通道
	module.Fields["close"] = &objects.Builtin{
		Name: "close",
		Fn: func(args ...objects.Object) objects.Object {
			if len(args) < 1 {
				return objects.NewError("close() expects at least 1 argument")
			}
			
			chIDArg, ok := args[0].(*objects.Integer)
			if !ok {
				return objects.NewTypeError("close() first argument must be a channel ID (integer)")
			}
			
			channelsMutex.Lock()
			ch, ok := channels[uint64(chIDArg.Value)]
			channelsMutex.Unlock()
			
			if !ok {
				return objects.NewError(fmt.Sprintf("channel %d not found", chIDArg.Value))
			}
			
			ch.Close()
			return objects.None_
		},
	}
	
	// sleep 函数 - 休眠
	module.Fields["sleep"] = &objects.Builtin{
		Name: "sleep",
		Fn: func(args ...objects.Object) objects.Object {
			if len(args) < 1 {
				return objects.NewError("sleep() expects at least 1 argument")
			}
			
			var secs float64
			switch v := args[0].(type) {
			case *objects.Integer:
				secs = float64(v.Value)
			case *objects.Float:
				secs = v.Value
			default:
				return objects.NewTypeError("sleep() argument must be a number")
			}
			
			time.Sleep(time.Duration(secs * float64(time.Second)))
			return objects.None_
		},
	}
	
	// waitgroup 相关
	module.Fields["waitgroup"] = &objects.Builtin{
		Name: "waitgroup",
		Fn: func(args ...objects.Object) objects.Object {
			wg := NewWaitGroup()
			waitGroupsMutex.Lock()
			wgID := nextWaitGroupID
			nextWaitGroupID++
			waitGroups[wgID] = wg
			waitGroupsMutex.Unlock()
			
			return &objects.Integer{Value: int64(wgID)}
		},
	}
	
	// waitgroup add
	module.Fields["add"] = &objects.Builtin{
		Name: "add",
		Fn: func(args ...objects.Object) objects.Object {
			if len(args) < 2 {
				return objects.NewError("add() expects waitgroup ID and delta as arguments")
			}
			
			wgIDArg, ok := args[0].(*objects.Integer)
			if !ok {
				return objects.NewTypeError("add() first argument must be a waitgroup ID (integer)")
			}
			
			deltaArg, ok := args[1].(*objects.Integer)
			if !ok {
				return objects.NewTypeError("add() second argument must be a delta (integer)")
			}
			
			waitGroupsMutex.Lock()
			wg, ok := waitGroups[uint64(wgIDArg.Value)]
			waitGroupsMutex.Unlock()
			
			if !ok {
				return objects.NewError(fmt.Sprintf("waitgroup %d not found", wgIDArg.Value))
			}
			
			wg.Add(int(deltaArg.Value))
			return objects.None_
		},
	}
	
	// waitgroup done
	module.Fields["done"] = &objects.Builtin{
		Name: "done",
		Fn: func(args ...objects.Object) objects.Object {
			if len(args) < 1 {
				return objects.NewError("done() expects waitgroup ID as argument")
			}
			
			wgIDArg, ok := args[0].(*objects.Integer)
			if !ok {
				return objects.NewTypeError("done() argument must be a waitgroup ID (integer)")
			}
			
			waitGroupsMutex.Lock()
			wg, ok := waitGroups[uint64(wgIDArg.Value)]
			waitGroupsMutex.Unlock()
			
			if !ok {
				return objects.NewError(fmt.Sprintf("waitgroup %d not found", wgIDArg.Value))
			}
			
			wg.Done()
			return objects.None_
		},
	}
	
	// waitgroup wait
	module.Fields["wait"] = &objects.Builtin{
		Name: "wait",
		Fn: func(args ...objects.Object) objects.Object {
			if len(args) < 1 {
				return objects.NewError("wait() expects waitgroup ID as argument")
			}
			
			wgIDArg, ok := args[0].(*objects.Integer)
			if !ok {
				return objects.NewTypeError("wait() argument must be a waitgroup ID (integer)")
			}
			
			waitGroupsMutex.Lock()
			wg, ok := waitGroups[uint64(wgIDArg.Value)]
			waitGroupsMutex.Unlock()
			
			if !ok {
				return objects.NewError(fmt.Sprintf("waitgroup %d not found", wgIDArg.Value))
			}
			
			wg.Wait()
			return objects.None_
		},
	}
	
	// mutex 相关
	module.Fields["mutex"] = &objects.Builtin{
		Name: "mutex",
		Fn: func(args ...objects.Object) objects.Object {
			mu := &Mutex{}
			mutexesMutex.Lock()
			muID := nextMutexID
			nextMutexID++
			mutexes[muID] = mu
			mutexesMutex.Unlock()
			
			return &objects.Integer{Value: int64(muID)}
		},
	}
	
	// lock
	module.Fields["lock"] = &objects.Builtin{
		Name: "lock",
		Fn: func(args ...objects.Object) objects.Object {
			if len(args) < 1 {
				return objects.NewError("lock() expects mutex ID as argument")
			}
			
			muIDArg, ok := args[0].(*objects.Integer)
			if !ok {
				return objects.NewTypeError("lock() argument must be a mutex ID (integer)")
			}
			
			mutexesMutex.Lock()
			mu, ok := mutexes[uint64(muIDArg.Value)]
			mutexesMutex.Unlock()
			
			if !ok {
				return objects.NewError(fmt.Sprintf("mutex %d not found", muIDArg.Value))
			}
			
			mu.Lock()
			return objects.None_
		},
	}
	
	// unlock
	module.Fields["unlock"] = &objects.Builtin{
		Name: "unlock",
		Fn: func(args ...objects.Object) objects.Object {
			if len(args) < 1 {
				return objects.NewError("unlock() expects mutex ID as argument")
			}
			
			muIDArg, ok := args[0].(*objects.Integer)
			if !ok {
				return objects.NewTypeError("unlock() argument must be a mutex ID (integer)")
			}
			
			mutexesMutex.Lock()
			mu, ok := mutexes[uint64(muIDArg.Value)]
			mutexesMutex.Unlock()
			
			if !ok {
				return objects.NewError(fmt.Sprintf("mutex %d not found", muIDArg.Value))
			}
			
			mu.Unlock()
			return objects.None_
		},
	}
	
	return module
}
