package core

import (
	"sort"
	"strings"
)

// HashEntry 是 HGETALL 返回用的字段-值对。
type HashEntry struct {
	Field string
	Value string
}

// RedisHash 是基于 Dict 构建的简化版 Redis Hash。
// 这一层的关键价值在于：把“通用字典”收敛成“field -> value”的命令语义。
type RedisHash struct {
	dict *Dict
}

// NewRedisHash 创建一个空 hash。
func NewRedisHash() *RedisHash {
	return &RedisHash{
		dict: NewDict(),
	}
}

// HSet 写入一个 field-value。
// 返回 true 表示新增字段，false 表示更新已有字段。
func (hash *RedisHash) HSet(field, value string) bool {
	if hash == nil || hash.dict == nil {
		return false
	}

	return hash.dict.Set(field, NewSDS(value))
}

// HGet 读取指定字段。
// 返回值约定：
// 1. 第一个返回值是字段值。
// 2. 第二个返回值表示字段是否存在。
func (hash *RedisHash) HGet(field string) (string, bool) {
	if hash == nil || hash.dict == nil {
		return "", false
	}

	value := hash.dict.Get(field)
	if value == nil {
		return "", false
	}

	return value.String(), true
}

// HDel 删除一个或多个字段，返回真正删除的个数。
func (hash *RedisHash) HDel(fields ...string) int {
	if hash == nil || hash.dict == nil || len(fields) == 0 {
		return 0
	}

	deleted := 0
	for _, field := range fields {
		if hash.dict.Delete(field) {
			deleted++
		}
	}

	return deleted
}

// HExists 判断字段是否存在。
func (hash *RedisHash) HExists(field string) bool {
	if hash == nil || hash.dict == nil {
		return false
	}

	return hash.dict.Contains(field)
}

// HLen 返回字段数。
func (hash *RedisHash) HLen() int {
	if hash == nil || hash.dict == nil {
		return 0
	}

	return hash.dict.Size()
}

// HGetAll 返回所有字段和值。
// 这里按 field 排序，主要是为了测试和调试时结果稳定。
func (hash *RedisHash) HGetAll() []HashEntry {
	if hash == nil || hash.dict == nil || hash.dict.Size() == 0 {
		return []HashEntry{}
	}

	fields := hash.dict.Keys()
	sort.Strings(fields)

	entries := make([]HashEntry, 0, len(fields))
	for _, field := range fields {
		value, ok := hash.HGet(field)
		if !ok {
			continue
		}

		entries = append(entries, HashEntry{
			Field: field,
			Value: value,
		})
	}

	return entries
}

// Type 实现 RedisValue 接口。
func (hash *RedisHash) Type() ValueType {
	return HashType
}

// String 实现 RedisValue 接口。
func (hash *RedisHash) String() string {
	entries := hash.HGetAll()
	if len(entries) == 0 {
		return "{}"
	}

	parts := make([]string, 0, len(entries))
	for _, entry := range entries {
		parts = append(parts, entry.Field+": "+entry.Value)
	}

	return "{" + strings.Join(parts, ", ") + "}"
}

// Size 估算 hash 的内存占用。
func (hash *RedisHash) Size() int64 {
	if hash == nil || hash.dict == nil {
		return 0
	}

	size := int64(24)
	for _, entry := range hash.HGetAll() {
		size += int64(len(entry.Field) + len(entry.Value))
	}

	return size
}
