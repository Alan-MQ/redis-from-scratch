package core

import (
	"fmt"
)

const (
	// Redis SDS分配策略常量
	SDS_MAX_PREALLOC = 1024 * 1024 // 1MB - 最大预分配大小
	SDS_MIN_SIZE     = 1           // 最小分配大小
)

// SDS (Simple Dynamic String) - Redis的动态字符串实现
// 设计对齐Redis 7.0的SDS结构
type SDS struct {
	buf   []byte // 数据缓冲区（raw buffer）
	used  int    // 已使用长度（实际字符串长度，对应Redis的len）
	alloc int    // 总分配长度（缓冲区容量，对应Redis的alloc）
}

// TODO: Alan 实现SDS创建函数
// 要求实现Redis的预分配策略：
// 1. 字符串长度 < 1MB: 分配 used * 2 的容量
// 2. 字符串长度 >= 1MB: 分配 used + 1MB 的容量
// 3. 空字符串至少分配1字节
// 4. 将输入字符串复制到缓冲区
func NewSDS(s string) *SDS {
	// Alan的实现已修复：去除逻辑重复，简化代码
	slen := len(s)
	var alloc int
	
	if slen == 0 {
		alloc = SDS_MIN_SIZE // 空字符串至少1字节
	} else if slen < SDS_MAX_PREALLOC {
		alloc = slen * 2 // 小于1MB：2倍
	} else {
		alloc = slen + SDS_MAX_PREALLOC // 大于1MB：+1MB
	}
	
	buf := make([]byte, alloc)
	copy(buf, s)
	
	return &SDS{
		buf:   buf,
		used:  slen,
		alloc: alloc,
	}
}

// TODO: Alan 实现字符串追加功能
// 要求实现Redis的扩容策略：
// 1. 计算新的总长度 newUsed = sds.used + len(s)
// 2. 如果 newUsed > sds.alloc，需要扩容：
//   - 如果 newUsed < 1MB: newAlloc = newUsed * 2
//   - 如果 newUsed >= 1MB: newAlloc = newUsed + 1MB
//
// 3. 重新分配缓冲区并复制原数据
// 4. 追加新字符串到缓冲区末尾
// 5. 更新 used 字段
func (sds *SDS) Append(s string) {
	// Alan的实现已修复：修复复制位置错误，移除调试信息
	if len(s) == 0 {
		return // 空字符串不做任何操作
	}
	
	newUsed := sds.used + len(s)
	
	// 检查容量是否足够
	if newUsed <= sds.alloc {
		// 容量足够，直接追加
		copy(sds.buf[sds.used:], s) // 修复：从 sds.used 开始复制
		sds.used = newUsed
		return
	}
	
	// 需要扩容，按Redis策略计算新容量
	var newAlloc int
	if newUsed < SDS_MAX_PREALLOC {
		newAlloc = newUsed * 2 // 小于1MB：2倍
	} else {
		newAlloc = newUsed + SDS_MAX_PREALLOC // 大于1MB：+1MB
	}
	
	// 重新分配缓冲区
	newBuf := make([]byte, newAlloc)
	copy(newBuf, sds.buf[:sds.used]) // 修复：只复制已使用部分
	copy(newBuf[sds.used:], s)       // 追加新数据
	
	// 更新SDS字段
	sds.buf = newBuf
	sds.used = newUsed
	sds.alloc = newAlloc
}

// Len 返回字符串长度 - 已实现
func (sds *SDS) Len() int {
	if sds == nil {
		return 0
	}
	return sds.used
}

// String 返回字符串内容 - 已实现
func (sds *SDS) String() string {
	if sds == nil || sds.used == 0 {
		return ""
	}
	return string(sds.buf[:sds.used])
}

// Size 计算内存占用 - 已实现
func (sds *SDS) Size() int64 {
	if sds == nil {
		return 0
	}
	// 结构体本身 + 缓冲区大小
	return int64(24) + int64(sds.alloc) // 24字节约等于结构体大小
}

// 实现RedisValue接口
func (sds *SDS) Type() ValueType {
	return StringType
}

// TODO: Alan 实现字符串截取功能
// 要求支持Redis GETRANGE命令的语义：
// 1. 支持负数索引：-1表示最后一个字符，-2表示倒数第二个
// 2. start > end 时返回空字符串
// 3. 索引超出范围时自动调整到有效范围
// 4. 左闭右闭区间：[start, end]
func (sds *SDS) Substr(start, end int) string {
	// 实现Redis GETRANGE命令的语义（支持负数索引）
	if sds == nil || sds.used == 0 {
		return ""
	}
	
	// 处理负数索引：-1表示最后一个字符
	if start < 0 {
		start = sds.used + start
	}
	if end < 0 {
		end = sds.used + end
	}
	
	// 边界检查：调整到有效范围
	if start < 0 {
		start = 0
	}
	if end >= sds.used {
		end = sds.used - 1
	}
	
	// start > end 时返回空字符串
	if start > end {
		return ""
	}
	
	// 返回子字符串（左闭右闭区间）
	return string(sds.buf[start : end+1])
}

// Clear 清空字符串内容但保留容量 - 已实现
func (sds *SDS) Clear() {
	if sds != nil {
		sds.used = 0
		// 缓冲区内容不需要清零，只需重置used即可
	}
}

// 新增方法：获取剩余容量
func (sds *SDS) Free() int {
	if sds == nil {
		return 0
	}
	return sds.alloc - sds.used
}

// 新增方法：获取总分配容量
func (sds *SDS) Alloc() int {
	if sds == nil {
		return 0
	}
	return sds.alloc
}

// Debug 用于调试的详细信息 - 已更新
func (sds *SDS) Debug() string {
	if sds == nil {
		return "SDS: <nil>"
	}
	return fmt.Sprintf("SDS: used=%d, alloc=%d, free=%d, data='%s'",
		sds.used, sds.alloc, sds.Free(), sds.String())
}
