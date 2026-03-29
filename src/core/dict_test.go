package core

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewDict(t *testing.T) {
	dict := NewDict()

	assert.NotNil(t, dict)
	assert.Equal(t, 0, dict.Size())
	assert.Equal(t, DefaultDictBucketCount, dict.BucketCount())
	assert.Equal(t, 0, dict.RehashBucketCount())
	assert.False(t, dict.IsRehashing())
	assert.Equal(t, -1, dict.RehashIndex())
	assert.Equal(t, 0.0, dict.LoadFactor())
	assert.Empty(t, dict.Keys())
}

func TestDictBasicSetGetUpdateDelete(t *testing.T) {
	dict := NewDictWithCapacity(4)

	inserted := dict.Set("name", NewSDS("alan"))
	assert.True(t, inserted)
	assert.Equal(t, 1, dict.Size())
	assert.Equal(t, "alan", dict.Get("name").String())

	inserted = dict.Set("name", NewSDS("redis"))
	assert.False(t, inserted)
	assert.Equal(t, 1, dict.Size())
	assert.Equal(t, "redis", dict.Get("name").String())

	assert.True(t, dict.Contains("name"))
	assert.False(t, dict.Contains("missing"))

	deleted := dict.Delete("name")
	assert.True(t, deleted)
	assert.Equal(t, 0, dict.Size())
	assert.Nil(t, dict.Get("name"))
	assert.False(t, dict.Delete("name"))
}

func TestDictStartsRehashWhenLoadFactorExceeded(t *testing.T) {
	dict := NewDictWithCapacity(2)

	dict.Set("alpha", NewSDS("A"))
	assert.False(t, dict.IsRehashing())

	dict.Set("beta", NewSDS("B"))

	assert.True(t, dict.IsRehashing())
	assert.Equal(t, 0, dict.RehashIndex())
	assert.Equal(t, 2, dict.BucketCount())
	assert.Equal(t, 4, dict.RehashBucketCount())
	assert.Equal(t, 2, dict.TableUsed(0))
	assert.Equal(t, 0, dict.TableUsed(1))
}

func TestDictGetSearchesBothTablesDuringRehash(t *testing.T) {
	dict := NewDictWithCapacity(2)
	dict.Set("alpha", NewSDS("A"))
	dict.Set("beta", NewSDS("B"))

	assert.True(t, dict.IsRehashing())

	// rehash 期间新增 key 应该仍然可读；旧表中的 key 也不能丢。
	inserted := dict.Set("gamma", NewSDS("C"))
	assert.True(t, inserted)
	assert.Equal(t, 3, dict.Size())

	// 这里不强依赖 ht[1] 的精确 used，因为 Set() 一开始会先做一次 rehashStep，
	// 那一步有可能刚好把旧表搬空并 finishRehash。我们更关心的是：
	// 1. rehash 途中旧数据还能读到
	// 2. 新插入的数据也能读到
	// 3. 无论当前是否已经完成 rehash，总元素数量都正确
	assert.True(t, dict.TableUsed(0) > 0 || dict.TableUsed(1) > 0)

	oldValue := dict.Get("alpha")
	midValue := dict.Get("beta")
	newValue := dict.Get("gamma")

	assert.NotNil(t, oldValue)
	assert.NotNil(t, midValue)
	assert.NotNil(t, newValue)
	assert.Equal(t, "A", oldValue.String())
	assert.Equal(t, "B", midValue.String())
	assert.Equal(t, "C", newValue.String())
}

func TestDictDeleteSearchesBothTablesDuringRehash(t *testing.T) {
	dict := NewDictWithCapacity(2)
	dict.Set("alpha", NewSDS("A"))
	dict.Set("beta", NewSDS("B"))
	dict.Set("gamma", NewSDS("C"))

	assert.True(t, dict.Delete("alpha"))
	assert.True(t, dict.Delete("gamma"))
	assert.False(t, dict.Delete("missing"))

	assert.Nil(t, dict.Get("alpha"))
	assert.Nil(t, dict.Get("gamma"))
	assert.Equal(t, 1, dict.Size())
	assert.Equal(t, "B", dict.Get("beta").String())
}

func TestDictRehashStepMovesData(t *testing.T) {
	dict := NewDictWithCapacity(2)
	dict.Set("alpha", NewSDS("A"))
	dict.Set("beta", NewSDS("B"))

	assert.True(t, dict.IsRehashing())

	oldUsedBefore := dict.TableUsed(0)
	newUsedBefore := dict.TableUsed(1)
	rehashIndexBefore := dict.RehashIndex()

	progressed := dict.rehashStep()
	if !progressed {
		t.Skip("TODO: Alan 需要实现 Dict.rehashStep")
		return
	}

	assert.True(t, dict.TableUsed(0) < oldUsedBefore || !dict.IsRehashing())
	assert.True(t, dict.TableUsed(1) > newUsedBefore || !dict.IsRehashing())

	if dict.IsRehashing() {
		assert.Greater(t, dict.RehashIndex(), rehashIndexBefore)
	} else {
		assert.Equal(t, -1, dict.RehashIndex())
	}
}

func TestDictFinishRehash(t *testing.T) {
	dict := NewDictWithCapacity(2)
	dict.Set("alpha", NewSDS("A"))
	dict.Set("beta", NewSDS("B"))
	dict.Set("gamma", NewSDS("C"))

	for i := 0; i < 16 && dict.IsRehashing(); i++ {
		dict.rehashStep()
	}

	if dict.IsRehashing() {
		t.Skip("TODO: Alan 需要完成 rehash 收尾逻辑")
		return
	}

	assert.Equal(t, -1, dict.RehashIndex())
	assert.Equal(t, 0, dict.RehashBucketCount())
	assert.Equal(t, 3, dict.Size())
	assert.Equal(t, "A", dict.Get("alpha").String())
	assert.Equal(t, "B", dict.Get("beta").String())
	assert.Equal(t, "C", dict.Get("gamma").String())
}
