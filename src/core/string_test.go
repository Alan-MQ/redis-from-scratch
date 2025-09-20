package core

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewSDS(t *testing.T) {
	t.Run("创建普通字符串", func(t *testing.T) {
		sds := NewSDS("hello")
		if sds == nil {
			t.Skip("TODO: Alan 需要先实现NewSDS函数")
			return
		}
		
		assert.NotNil(t, sds)
		assert.Equal(t, "hello", sds.String())
		assert.Equal(t, 5, sds.Len())
		assert.True(t, sds.Alloc() >= sds.Len(), "分配容量应该大于等于字符串长度")
	})
	
	t.Run("空字符串处理", func(t *testing.T) {
		sds := NewSDS("")
		if sds == nil {
			t.Skip("TODO: Alan 需要先实现NewSDS函数")
			return
		}
		
		assert.Equal(t, "", sds.String())
		assert.Equal(t, 0, sds.Len())
		assert.True(t, sds.Alloc() >= SDS_MIN_SIZE, "空字符串至少分配1字节")
	})
	
	t.Run("预分配策略验证", func(t *testing.T) {
		// 小字符串：应该分配2倍容量
		smallSDS := NewSDS("test")
		if smallSDS == nil {
			t.Skip("TODO: Alan 需要先实现NewSDS函数")
			return
		}
		expectedSmallAlloc := len("test") * 2
		assert.Equal(t, expectedSmallAlloc, smallSDS.Alloc(), "小字符串应该分配2倍容量")
		
		// 大字符串：应该分配+1MB容量
		bigString := make([]byte, SDS_MAX_PREALLOC+100) // > 1MB
		for i := range bigString {
			bigString[i] = 'a'
		}
		bigSDS := NewSDS(string(bigString))
		if bigSDS == nil {
			t.Skip("TODO: Alan 需要先实现NewSDS函数")
			return
		}
		expectedBigAlloc := len(bigString) + SDS_MAX_PREALLOC
		assert.Equal(t, expectedBigAlloc, bigSDS.Alloc(), "大字符串应该分配+1MB容量")
	})
}

func TestSDSAppend(t *testing.T) {
	t.Run("基本追加功能", func(t *testing.T) {
		sds := NewSDS("hello")
		if sds == nil {
			t.Skip("NewSDS not implemented yet")
			return
		}
		
		sds.Append(" world")
		assert.Equal(t, "hello world", sds.String())
		assert.Equal(t, 11, sds.Len())
	})
	
	t.Run("追加空字符串", func(t *testing.T) {
		sds := NewSDS("test")
		if sds == nil {
			t.Skip("NewSDS not implemented yet")
			return
		}
		
		originalLen := sds.Len()
		sds.Append("")
		assert.Equal(t, originalLen, sds.Len(), "追加空字符串不应改变长度")
	})
	
	t.Run("扩容策略验证", func(t *testing.T) {
		sds := NewSDS("a")
		if sds == nil {
			t.Skip("NewSDS not implemented yet")
			return
		}
		
		originalAlloc := sds.Alloc()
		
		// 追加足够多的数据触发扩容
		longString := make([]byte, originalAlloc)
		for i := range longString {
			longString[i] = 'x'
		}
		
		sds.Append(string(longString))
		assert.True(t, sds.Alloc() > originalAlloc, "应该触发扩容")
		assert.True(t, sds.Alloc() >= sds.Len(), "扩容后容量应该足够")
	})
}

func TestSDSSubstr(t *testing.T) {
	t.Run("正常索引截取", func(t *testing.T) {
		sds := NewSDS("hello world")
		if sds == nil {
			t.Skip("NewSDS not implemented yet")
			return
		}
		
		result := sds.Substr(0, 4)
		assert.Equal(t, "hello", result)
		
		result = sds.Substr(6, 10)
		assert.Equal(t, "world", result)
	})
	
	t.Run("负数索引截取", func(t *testing.T) {
		sds := NewSDS("hello world")
		if sds == nil {
			t.Skip("NewSDS not implemented yet")
			return
		}
		
		// -1表示最后一个字符
		result := sds.Substr(-5, -1)
		assert.Equal(t, "world", result)
		
		// 混合正负索引
		result = sds.Substr(0, -7)
		assert.Equal(t, "hello", result)
	})
	
	t.Run("边界情况", func(t *testing.T) {
		sds := NewSDS("test")
		if sds == nil {
			t.Skip("NewSDS not implemented yet")
			return
		}
		
		// start > end
		result := sds.Substr(3, 1)
		assert.Equal(t, "", result, "start > end应该返回空字符串")
		
		// 超出范围的索引
		result = sds.Substr(-100, 100)
		assert.Equal(t, "test", result, "超出范围的索引应该调整到有效范围")
	})
}

func TestSDSUtilityMethods(t *testing.T) {
	sds := NewSDS("hello")
	if sds == nil {
		t.Skip("NewSDS not implemented yet")
		return
	}
	
	// 测试各种工具方法
	assert.Equal(t, StringType, sds.Type())
	assert.True(t, sds.Size() > 0, "Size应该大于0")
	assert.True(t, sds.Free() >= 0, "Free应该大于等于0")
	
	// 测试Clear
	sds.Clear()
	assert.Equal(t, 0, sds.Len())
	assert.Equal(t, "", sds.String())
	assert.True(t, sds.Alloc() > 0, "Clear后应该保留容量")
	
	// 测试Debug信息
	debugInfo := sds.Debug()
	assert.Contains(t, debugInfo, "SDS:")
	assert.Contains(t, debugInfo, "used=")
	assert.Contains(t, debugInfo, "alloc=")
}

func TestSDSNilSafety(t *testing.T) {
	// 测试nil指针安全性
	var sds *SDS
	
	assert.Equal(t, 0, sds.Len())
	assert.Equal(t, "", sds.String())
	assert.Equal(t, int64(0), sds.Size())
	assert.Equal(t, 0, sds.Free())
	assert.Equal(t, 0, sds.Alloc())
	assert.Equal(t, "SDS: <nil>", sds.Debug())
	
	// Clear nil指针应该安全
	sds.Clear() // 不应该panic
}

// 基准测试：比较SDS和Go原生string的性能
func BenchmarkSDSAppend(b *testing.B) {
	for i := 0; i < b.N; i++ {
		sds := NewSDS("initial")
		if sds == nil {
			b.Skip("NewSDS not implemented yet")
			return
		}
		
		for j := 0; j < 100; j++ {
			sds.Append("x")
		}
	}
}

func BenchmarkGoStringAppend(b *testing.B) {
	for i := 0; i < b.N; i++ {
		s := "initial"
		for j := 0; j < 100; j++ {
			s += "x"
		}
	}
}

func BenchmarkSDSCreate(b *testing.B) {
	testString := "hello world test string for benchmarking"
	
	for i := 0; i < b.N; i++ {
		sds := NewSDS(testString)
		if sds == nil {
			b.Skip("NewSDS not implemented yet")
			return
		}
		_ = sds
	}
}