package concurrency

import (
	"fmt"
	"sync"

	"github.com/go-py/go-python/pkg/objects"
)

// ConcurrentList 并发安全的列表
type ConcurrentList struct {
	Elements []objects.Object
	mu       sync.RWMutex
}

// NewConcurrentList 创建新的并发安全列表
func NewConcurrentList() *ConcurrentList {
	return &ConcurrentList{
		Elements: make([]objects.Object, 0),
	}
}

// Append 添加元素
func (cl *ConcurrentList) Append(obj objects.Object) {
	cl.mu.Lock()
	defer cl.mu.Unlock()
	cl.Elements = append(cl.Elements, obj)
}

// Get 获取元素
func (cl *ConcurrentList) Get(index int) (objects.Object, bool) {
	cl.mu.RLock()
	defer cl.mu.RUnlock()
	if index < 0 || index >= len(cl.Elements) {
		return objects.None_, false
	}
	return cl.Elements[index], true
}

// Set 设置元素
func (cl *ConcurrentList) Set(index int, obj objects.Object) bool {
	cl.mu.Lock()
	defer cl.mu.Unlock()
	if index < 0 || index >= len(cl.Elements) {
		return false
	}
	cl.Elements[index] = obj
	return true
}

// Len 获取长度
func (cl *ConcurrentList) Len() int {
	cl.mu.RLock()
	defer cl.mu.RUnlock()
	return len(cl.Elements)
}

// Pop 弹出元素
func (cl *ConcurrentList) Pop(index ...int) (objects.Object, error) {
	cl.mu.Lock()
	defer cl.mu.Unlock()
	
	idx := len(cl.Elements) - 1
	if len(index) > 0 {
		idx = index[0]
		if idx < 0 {
			idx = len(cl.Elements) + idx
		}
	}
	
	if idx < 0 || idx >= len(cl.Elements) {
		return objects.None_, nil // Temporarily return nil as error
	}
	
	obj := cl.Elements[idx]
	cl.Elements = append(cl.Elements[:idx], cl.Elements[idx+1:]...)
	return obj, nil
}

// ConcurrentDict 并发安全的字典
type ConcurrentDict struct {
	Pairs map[string]objects.Object
	Keys  map[string]objects.Object
	mu    sync.RWMutex
}

// NewConcurrentDict 创建新的并发安全字典
func NewConcurrentDict() *ConcurrentDict {
	return &ConcurrentDict{
		Pairs: make(map[string]objects.Object),
		Keys:  make(map[string]objects.Object),
	}
}

// Get 获取值
func (cd *ConcurrentDict) Get(key objects.Object) (objects.Object, bool) {
	cd.mu.RLock()
	defer cd.mu.RUnlock()
	
	keyStr := cd.hashKey(key)
	val, ok := cd.Pairs[keyStr]
	return val, ok
}

// Set 设置值
func (cd *ConcurrentDict) Set(key, value objects.Object) {
	cd.mu.Lock()
	defer cd.mu.Unlock()
	
	keyStr := cd.hashKey(key)
	cd.Pairs[keyStr] = value
	cd.Keys[keyStr] = key
}

// Has 检查键是否存在
func (cd *ConcurrentDict) Has(key objects.Object) bool {
	cd.mu.RLock()
	defer cd.mu.RUnlock()
	
	keyStr := cd.hashKey(key)
	_, ok := cd.Pairs[keyStr]
	return ok
}

// Delete 删除键值对
func (cd *ConcurrentDict) Delete(key objects.Object) {
	cd.mu.Lock()
	defer cd.mu.Unlock()
	
	keyStr := cd.hashKey(key)
	delete(cd.Pairs, keyStr)
	delete(cd.Keys, keyStr)
}

// Len 获取长度
func (cd *ConcurrentDict) Len() int {
	cd.mu.RLock()
	defer cd.mu.RUnlock()
	return len(cd.Pairs)
}

// KeysSlice 获取所有键
func (cd *ConcurrentDict) KeysSlice() []objects.Object {
	cd.mu.RLock()
	defer cd.mu.RUnlock()
	
	keys := make([]objects.Object, 0, len(cd.Keys))
	for _, key := range cd.Keys {
		keys = append(keys, key)
	}
	return keys
}

// ValuesSlice 获取所有值
func (cd *ConcurrentDict) ValuesSlice() []objects.Object {
	cd.mu.RLock()
	defer cd.mu.RUnlock()
	
	values := make([]objects.Object, 0, len(cd.Pairs))
	for _, val := range cd.Pairs {
		values = append(values, val)
	}
	return values
}

func (cd *ConcurrentDict) hashKey(obj objects.Object) string {
	switch o := obj.(type) {
	case *objects.Integer:
		return fmt.Sprintf("int:%d", o.Value)
	case *objects.Float:
		return fmt.Sprintf("float:%f", o.Value)
	case *objects.Boolean:
		return fmt.Sprintf("bool:%t", o.Value)
	case *objects.String:
		return "str:" + o.Value
	default:
		// 使用指针地址作为哈希
		return fmt.Sprintf("obj:%p", obj)
	}
}
