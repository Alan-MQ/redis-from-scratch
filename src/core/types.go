package core

import (
	"fmt"
	"time"
)

// Redis数据类型枚举
type ValueType int

const (
	StringType ValueType = iota
	ListType
	SetType
	HashType
	ZSetType
)

func (t ValueType) String() string {
	switch t {
	case StringType:
		return "string"
	case ListType:
		return "list"
	case SetType:
		return "set"
	case HashType:
		return "hash"
	case ZSetType:
		return "zset"
	default:
		return "unknown"
	}
}

// Redis值的通用接口
type RedisValue interface {
	Type() ValueType
	String() string
	Size() int64  // 占用内存大小（字节）
}

// Redis对象，包装值和元数据
type RedisObject struct {
	Value      RedisValue    // 实际数据
	ExpireTime *time.Time    // 过期时间，nil表示不过期
	LastAccess time.Time     // 最后访问时间（用于LRU）
}

// 创建新的Redis对象
func NewRedisObject(value RedisValue) *RedisObject {
	return &RedisObject{
		Value:      value,
		LastAccess: time.Now(),
	}
}

// 检查是否过期
func (obj *RedisObject) IsExpired() bool {
	if obj.ExpireTime == nil {
		return false
	}
	return time.Now().After(*obj.ExpireTime)
}

// 设置过期时间
func (obj *RedisObject) SetExpire(ttl time.Duration) {
	expireTime := time.Now().Add(ttl)
	obj.ExpireTime = &expireTime
}

// 更新访问时间
func (obj *RedisObject) Touch() {
	obj.LastAccess = time.Now()
}

// 获取对象信息
func (obj *RedisObject) Info() string {
	expire := "never"
	if obj.ExpireTime != nil {
		expire = obj.ExpireTime.Format("2006-01-02 15:04:05")
	}
	
	return fmt.Sprintf("Type: %s, Size: %d bytes, Expire: %s", 
		obj.Value.Type(), obj.Value.Size(), expire)
}