package core

import (
	"fmt"
	"strings"
)

// ListNode 双向链表节点
type ListNode struct {
	Value RedisValue // 节点存储的值
	Prev  *ListNode  // 前驱节点
	Next  *ListNode  // 后继节点
}

// NewListNode 创建新的链表节点
func NewListNode(value RedisValue) *ListNode {
	return &ListNode{
		Value: value,
		Prev:  nil,
		Next:  nil,
	}
}

// RedisList Redis列表实现（双向链表）
type RedisList struct {
	Head   *ListNode // 头节点
	Tail   *ListNode // 尾节点
	Length int       // 列表长度
}

// NewRedisList 创建新的Redis列表
func NewRedisList() *RedisList {
	// 为了方便处理 我觉得是不是Head 和 Tail 采用特殊节点不存储数据
	return &RedisList{
		Head:   nil,
		Tail:   nil,
		Length: 0,
	}
}

// TODO: Alan 实现头部插入操作 (对应Redis的LPUSH)
// 要求：
// 1. 创建新节点包装value
// 2. 将新节点插入到头部
// 3. 正确处理空列表的情况
// 4. 更新head指针和length
// 5. 时间复杂度：O(1)
func (list *RedisList) LPush(value RedisValue) {
	// 提示：
	// - newNode := NewListNode(value)
	// - 处理空列表：if list.head == nil
	// - 非空列表：更新链接关系
	// - 更新head和length
	newNode := NewListNode(value)
	newNode.Next = list.Head
	list.Head = newNode
	if newNode.Next != nil {
		newNode.Next.Prev = newNode
	}
	list.Length += 1
	// 这里你说要处理 list.head == nil  我没get到 为什么要处理?
	if list.Tail == nil {
		list.Tail = newNode
	}
}

// TODO: Alan 实现尾部插入操作 (对应Redis的RPUSH)
// 要求：
// 1. 创建新节点包装value
// 2. 将新节点插入到尾部
// 3. 正确处理空列表的情况
// 4. 更新tail指针和length
// 5. 时间复杂度：O(1)
func (list *RedisList) RPush(value RedisValue) {
	// 提示：
	// - 类似LPush，但是操作tail端
	// - 注意空列表时head和tail都要设置
	newNode := NewListNode(value)

	if list.Tail != nil {
		list.Tail.Next = newNode
		newNode.Prev = list.Tail
		list.Tail = newNode
		list.Length += 1
		return
	}
	// empty list
	list.Tail = newNode
	list.Head = newNode
	list.Length = 1
}

// TODO: Alan 实现头部弹出操作 (对应Redis的LPOP)
// 要求：
// 1. 检查列表是否为空
// 2. 保存头节点的值
// 3. 移除头节点，更新head指针
// 4. 处理只有一个节点的情况
// 5. 更新length，返回值
// 6. 时间复杂度：O(1)
func (list *RedisList) LPop() RedisValue {
	// 提示：
	// - 空列表检查：if list.head == nil
	// - 保存值：value := list.head.value
	// - 更新链接：list.head = list.head.next
	// - 特殊情况：只有一个节点时tail也要更新
	if list.Head == nil {
		return nil
	}
	list.Length -= 1
	res := list.Head.Value
	if list.Length == 0 {
		list.Head = nil
		list.Tail = nil
		return res
	}
	list.Head = list.Head.Next
	list.Head.Prev = nil
	return res
}

// TODO: Alan 实现尾部弹出操作 (对应Redis的RPOP)
// 要求：类似LPop，但操作tail端
func (list *RedisList) RPop() RedisValue {
	if list.Tail == nil {
		return nil
	}

	res := list.Tail.Value

	if list.Head == list.Tail {
		list.Head = nil
		list.Tail = nil
		list.Length = 0
		return res
	}

	list.Length -= 1
	list.Tail = list.Tail.Prev
	list.Tail.Next = nil
	return res
}

// TODO: Alan 实现按索引获取元素 (对应Redis的LINDEX)
// 要求：
// 1. 支持负数索引：-1表示最后一个元素
// 2. 索引越界返回nil
// 3. 从头或尾开始遍历（选择更短路径）
// 4. 时间复杂度：O(n)，但有优化
func (list *RedisList) LIndex(index int) RedisValue {
	// 提示：
	// - 处理负数索引：if index < 0 { index = list.length + index }
	// - 边界检查：index >= 0 && index < list.length
	// - 路径优化：从头还是从尾开始遍历？
	// - 遍历到目标位置

	if list.Length == 0 || index >= list.Length || index < (-1*list.Length) {
		return nil
	}
	if index < 0 {
		index += list.Length
	}
	if index > (list.Length / 2) {
		curr := list.Tail
		for i := 0; i < list.Length-index-1; i++ {
			curr = curr.Prev
		}
		return curr.Value
	} else {
		curr := list.Head
		for i := 0; i < index; i++ {
			curr = curr.Next
		}
		return curr.Value
	}
}

// Len 返回列表长度 - 已实现
func (list *RedisList) Len() int {
	return list.Length
}

// 实现RedisValue接口
func (list *RedisList) Type() ValueType {
	return ListType
}

func (list *RedisList) String() string {
	if list.Length == 0 {
		return "[]"
	}

	var parts []string
	current := list.Head
	for current != nil {
		parts = append(parts, current.Value.String())
		current = current.Next
	}

	return "[" + strings.Join(parts, ", ") + "]"
}

func (list *RedisList) Size() int64 {
	// 估算内存占用：结构体 + 节点数 * 节点大小
	nodeSize := int64(32) // 估算每个节点的大小
	return int64(24) + int64(list.Length)*nodeSize
}

// IsEmpty 检查列表是否为空 - 已实现
func (list *RedisList) IsEmpty() bool {
	return list.Length == 0
}

// Clear 清空列表 - 已实现
func (list *RedisList) Clear() {
	list.Head = nil
	list.Tail = nil
	list.Length = 0
}

// TODO: Alan 实现范围获取操作 (对应Redis的LRANGE)
// 要求：
// 1. 支持负数索引
// 2. 返回[start, stop]范围内的元素
// 3. 超出范围自动调整
// 4. start > stop时返回空列表
func (list *RedisList) LRange(start, stop int) []RedisValue {
	if list == nil || list.Length == 0 {
		return []RedisValue{}
	}

	if start < 0 {
		start = list.Length + start
	}
	if stop < 0 {
		stop = list.Length + stop
	}

	if start < 0 {
		start = 0
	}
	if stop < 0 || start >= list.Length {
		return []RedisValue{}
	}
	if stop >= list.Length {
		stop = list.Length - 1
	}

	if start > stop {
		return []RedisValue{}
	}

	curr := list.Head
	for i := 0; i < start; i++ {
		curr = curr.Next
	}

	res := make([]RedisValue, 0, stop-start+1)
	for i := start; i <= stop && curr != nil; i++ {
		res = append(res, curr.Value)
		curr = curr.Next
	}

	return res
}

// Debug 调试信息 - 已实现
func (list *RedisList) Debug() string {
	return fmt.Sprintf("RedisList: length=%d, data=%s",
		list.Length, list.String())
}
