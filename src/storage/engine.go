package storage

import (
	"sync"

	"redis-from-scratch/src/core"
)

// Engine 是最小可用的内存数据库骨架。
// 这里先把“存取 RedisObject”的通路接好，TTL/淘汰策略放到后续阶段。
type Engine struct {
	data  *core.Dict
	mutex sync.Mutex
}

// NewEngine 创建一个空的内存引擎。
func NewEngine() *Engine {
	return &Engine{
		data: core.NewDict(),
	}
}

// Set 把一个 RedisValue 包装成 RedisObject 后写入引擎。
func (engine *Engine) Set(key string, value core.RedisValue) {
	if engine == nil || engine.data == nil {
		return
	}

	engine.mutex.Lock()
	defer engine.mutex.Unlock()

	engine.data.Set(key, core.NewRedisObject(value))
}

// Get 读取 key 对应的 RedisObject。
// 这里顺手更新访问时间，给后面的 LRU/过期策略留接口。
func (engine *Engine) Get(key string) *core.RedisObject {
	if engine == nil || engine.data == nil {
		return nil
	}

	engine.mutex.Lock()
	defer engine.mutex.Unlock()

	value := engine.data.Get(key)
	if value == nil {
		return nil
	}

	obj, ok := value.(*core.RedisObject)
	if !ok {
		obj = core.NewRedisObject(value)
	}

	if obj.IsExpired() {
		engine.data.Delete(key)
		return nil
	}

	obj.Touch()
	return obj
}

// Delete 删除一个或多个 key，返回真正删除的个数。
func (engine *Engine) Delete(keys ...string) int {
	if engine == nil || engine.data == nil || len(keys) == 0 {
		return 0
	}

	engine.mutex.Lock()
	defer engine.mutex.Unlock()

	deleted := 0
	for _, key := range keys {
		if engine.data.Delete(key) {
			deleted++
		}
	}

	return deleted
}

// Keys 返回当前所有 key。
func (engine *Engine) Keys() []string {
	if engine == nil || engine.data == nil {
		return []string{}
	}

	engine.mutex.Lock()
	defer engine.mutex.Unlock()

	return engine.data.Keys()
}

// Size 返回当前 key 的数量。
func (engine *Engine) Size() int {
	if engine == nil || engine.data == nil {
		return 0
	}

	engine.mutex.Lock()
	defer engine.mutex.Unlock()

	return engine.data.Size()
}
