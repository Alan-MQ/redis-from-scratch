package core

import (
	"sort"
	"strings"
)

// RedisSet 是基于 Dict 构建的简化版集合。
// 这里我直接把薄封装补齐，让你把精力集中在 dict 的渐进式 rehash。
type RedisSet struct {
	dict *Dict
}

// NewRedisSet 创建一个空集合。
func NewRedisSet() *RedisSet {
	return &RedisSet{
		dict: NewDict(),
	}
}

// SAdd 添加一个或多个成员。
// 返回本次真正新增的成员数。
func (set *RedisSet) SAdd(members ...string) int {
	if set == nil || set.dict == nil || len(members) == 0 {
		return 0
	}

	added := 0
	for _, member := range members {
		if set.dict.Set(member, NewSDS(member)) {
			added++
		}
	}

	return added
}

// SRem 删除一个或多个成员。
// 返回本次真正删除的成员数。
func (set *RedisSet) SRem(members ...string) int {
	if set == nil || set.dict == nil || len(members) == 0 {
		return 0
	}

	removed := 0
	for _, member := range members {
		if set.dict.Delete(member) {
			removed++
		}
	}

	return removed
}

// SIsMember 判断成员是否存在。
func (set *RedisSet) SIsMember(member string) bool {
	if set == nil || set.dict == nil {
		return false
	}

	return set.dict.Contains(member)
}

// SMembers 返回所有成员。
func (set *RedisSet) SMembers() []string {
	if set == nil || set.dict == nil {
		return []string{}
	}

	return set.dict.Keys()
}

// SCard 返回集合基数。
func (set *RedisSet) SCard() int {
	if set == nil || set.dict == nil {
		return 0
	}
	return set.dict.Size()
}

// Type 实现 RedisValue 接口。
func (set *RedisSet) Type() ValueType {
	return SetType
}

// String 实现 RedisValue 接口。
func (set *RedisSet) String() string {
	members := set.SMembers()
	if len(members) == 0 {
		return "{}"
	}

	sort.Strings(members)
	return "{" + strings.Join(members, ", ") + "}"
}

// Size 估算集合的内存占用。
func (set *RedisSet) Size() int64 {
	if set == nil || set.dict == nil {
		return 0
	}

	size := int64(24)
	for _, member := range set.SMembers() {
		size += int64(len(member))
	}
	return size
}
