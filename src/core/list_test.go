package core

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewRedisList(t *testing.T) {
	list := NewRedisList()

	assert.NotNil(t, list)
	assert.Equal(t, 0, list.Len())
	assert.True(t, list.IsEmpty())
	assert.Equal(t, ListType, list.Type())
}

func TestListLPush(t *testing.T) {
	list := NewRedisList()

	// 测试单个元素插入
	str1 := NewSDS("hello")
	if str1 == nil {
		t.Skip("SDS not implemented yet")
		return
	}

	list.LPush(str1)
	if list.Len() == 0 {
		t.Skip("TODO: Alan 需要先实现LPush函数")
		return
	}

	assert.Equal(t, 1, list.Len())
	assert.False(t, list.IsEmpty())

	// 测试多个元素插入
	str2 := NewSDS("world")
	list.LPush(str2)
	assert.Equal(t, 2, list.Len())
}

func TestListRPush(t *testing.T) {
	list := NewRedisList()

	str1 := NewSDS("first")
	if str1 == nil {
		t.Skip("SDS not implemented yet")
		return
	}

	list.RPush(str1)
	if list.Len() == 0 {
		t.Skip("TODO: Alan 需要先实现RPush函数")
		return
	}

	assert.Equal(t, 1, list.Len())

	str2 := NewSDS("second")
	list.RPush(str2)
	assert.Equal(t, 2, list.Len())
}

func TestListLPop(t *testing.T) {
	list := NewRedisList()

	// 空列表弹出
	result := list.LPop()
	assert.Nil(t, result, "空列表LPop应该返回nil")

	// 单个元素
	str1 := NewSDS("only")
	if str1 == nil {
		t.Skip("SDS not implemented yet")
		return
	}

	list.LPush(str1)
	if list.Len() == 0 {
		t.Skip("LPush not implemented yet")
		return
	}

	result = list.LPop()
	if result == nil {
		t.Skip("TODO: Alan 需要先实现LPop函数")
		return
	}

	assert.Equal(t, "only", result.String())
	assert.Equal(t, 0, list.Len())
	assert.True(t, list.IsEmpty())
}

func TestListRPop(t *testing.T) {
	list := NewRedisList()

	// 空列表弹出
	result := list.RPop()
	assert.Nil(t, result, "空列表RPop应该返回nil")

	// 多个元素测试
	str1 := NewSDS("first")
	str2 := NewSDS("second")
	if str1 == nil || str2 == nil {
		t.Skip("SDS not implemented yet")
		return
	}

	list.RPush(str1)
	list.RPush(str2)
	if list.Len() != 2 {
		t.Skip("RPush not implemented yet")
		return
	}

	result = list.RPop()
	if result == nil {
		t.Skip("TODO: Alan 需要先实现RPop函数")
		return
	}

	assert.Equal(t, "second", result.String())
	assert.Equal(t, 1, list.Len())
	assert.NotNil(t, list.Head)
	assert.NotNil(t, list.Tail)
	assert.Equal(t, "first", list.Head.Value.String())
	assert.Equal(t, "first", list.Tail.Value.String())
}

func TestListLIndex(t *testing.T) {
	list := NewRedisList()

	// 空列表
	result := list.LIndex(0)
	assert.Nil(t, result, "空列表LIndex应该返回nil")

	// 添加测试数据
	for i := 0; i < 3; i++ {
		str := NewSDS(fmt.Sprintf("item%d", i))
		if str == nil {
			t.Skip("SDS not implemented yet")
			return
		}
		list.RPush(str)
	}

	if list.Len() != 3 {
		t.Skip("RPush not implemented yet")
		return
	}

	// 正数索引
	result = list.LIndex(0)
	if result == nil {
		t.Skip("TODO: Alan 需要先实现LIndex函数")
		return
	}
	assert.Equal(t, "item0", result.String())

	result = list.LIndex(2)
	assert.Equal(t, "item2", result.String())

	// 负数索引
	result = list.LIndex(-1)
	assert.Equal(t, "item2", result.String())

	result = list.LIndex(-3)
	assert.Equal(t, "item0", result.String())

	// 越界索引
	result = list.LIndex(10)
	assert.Nil(t, result)

	result = list.LIndex(-10)
	assert.Nil(t, result)
}

func TestListLRange(t *testing.T) {
	list := NewRedisList()

	// 空列表
	result := list.LRange(0, 1)
	if result == nil {
		t.Skip("TODO: Alan 需要先实现LRange函数")
		return
	}
	assert.Empty(t, result, "空列表LRange应该返回空slice")

	// 添加测试数据
	for i := 0; i < 5; i++ {
		str := NewSDS(fmt.Sprintf("val%d", i))
		if str == nil {
			t.Skip("SDS not implemented yet")
			return
		}
		list.RPush(str)
	}

	if list.Len() != 5 {
		t.Skip("RPush not implemented yet")
		return
	}

	// 正常范围
	result = list.LRange(1, 3)
	assert.Len(t, result, 3)
	assert.Equal(t, "val1", result[0].String())
	assert.Equal(t, "val3", result[2].String())

	// 负数索引
	result = list.LRange(-2, -1)
	assert.Len(t, result, 2)
	assert.Equal(t, "val3", result[0].String())
	assert.Equal(t, "val4", result[1].String())

	// 超出范围
	result = list.LRange(0, 100)
	assert.Len(t, result, 5) // 应该返回全部元素

	// start > stop
	result = list.LRange(3, 1)
	assert.Empty(t, result)

	// 靠近尾部的范围也应该保持正向顺序
	result = list.LRange(2, 2)
	assert.Len(t, result, 1)
	assert.Equal(t, "val2", result[0].String())

	result = list.LRange(1, 2)
	assert.Len(t, result, 2)
	assert.Equal(t, "val1", result[0].String())
	assert.Equal(t, "val2", result[1].String())
}

func TestListOperationSequence(t *testing.T) {
	// 综合测试：模拟Redis LPUSH + RPOP的队列操作
	list := NewRedisList()

	// 添加元素
	for i := 0; i < 3; i++ {
		str := NewSDS(fmt.Sprintf("task%d", i))
		if str == nil {
			t.Skip("SDS not implemented yet")
			return
		}
		list.LPush(str) // 头部插入
	}

	if list.Len() != 3 {
		t.Skip("LPush not implemented yet")
		return
	}

	// 弹出元素
	result1 := list.RPop() // 尾部弹出
	if result1 == nil {
		t.Skip("RPop not implemented yet")
		return
	}
	assert.Equal(t, "task0", result1.String()) // 第一个插入的应该最先弹出

	result2 := list.RPop()
	assert.Equal(t, "task1", result2.String())

	assert.Equal(t, 1, list.Len())
}

func TestListClear(t *testing.T) {
	list := NewRedisList()

	// 添加元素
	str := NewSDS("test")
	if str == nil {
		t.Skip("SDS not implemented yet")
		return
	}

	list.LPush(str)
	if list.Len() == 0 {
		t.Skip("LPush not implemented yet")
		return
	}

	// 清空
	list.Clear()
	assert.Equal(t, 0, list.Len())
	assert.True(t, list.IsEmpty())

	// 清空后应该能正常使用
	list.LPush(str)
	assert.Equal(t, 1, list.Len())
}

func TestListLPopClearsNewHeadPrev(t *testing.T) {
	list := NewRedisList()

	list.RPush(NewSDS("a"))
	list.RPush(NewSDS("b"))
	list.RPush(NewSDS("c"))

	result := list.LPop()
	assert.NotNil(t, result)
	assert.Equal(t, "a", result.String())
	assert.Equal(t, 2, list.Len())
	assert.NotNil(t, list.Head)
	assert.Equal(t, "b", list.Head.Value.String())
	assert.Nil(t, list.Head.Prev)
}

func TestListLIndexEqualToLengthIsOutOfRange(t *testing.T) {
	list := NewRedisList()

	list.RPush(NewSDS("a"))
	list.RPush(NewSDS("b"))

	assert.Nil(t, list.LIndex(2))
}

// 基准测试：双端操作性能
func BenchmarkListLPush(b *testing.B) {
	list := NewRedisList()
	str := NewSDS("benchmark")
	if str == nil {
		b.Skip("SDS not implemented yet")
		return
	}

	for i := 0; i < b.N; i++ {
		list.LPush(str)
		if list.Len() == 0 {
			b.Skip("LPush not implemented yet")
			return
		}
	}
}

func BenchmarkListLPop(b *testing.B) {
	list := NewRedisList()
	str := NewSDS("benchmark")
	if str == nil {
		b.Skip("SDS not implemented yet")
		return
	}

	// 预填充数据
	for i := 0; i < b.N; i++ {
		list.LPush(str)
		if list.Len() == 0 {
			b.Skip("LPush not implemented yet")
			return
		}
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		result := list.LPop()
		if result == nil {
			b.Skip("LPop not implemented yet")
			return
		}
	}
}
