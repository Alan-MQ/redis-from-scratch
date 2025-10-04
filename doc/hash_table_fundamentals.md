# 哈希表基础知识 - Redis Set/Hash实现必备

## 🎯 为什么要学习哈希表？

**哈希表是Redis最核心的数据结构！**
- **Redis Set**: 直接基于哈希表实现
- **Redis Hash**: 渐进式rehash的哈希表
- **Redis内部**: 所有key-value存储都基于哈希表

---

## 📚 哈希表基础概念

### 1. 什么是哈希表？

**定义**: 通过哈希函数将key映射到数组索引的数据结构

```
Key → Hash Function → Index → Bucket → Value
```

**核心优势**:
- **查找**: O(1) 平均时间复杂度
- **插入**: O(1) 平均时间复杂度  
- **删除**: O(1) 平均时间复杂度

### 2. 哈希函数设计

**好的哈希函数特点**:
1. **确定性**: 相同输入总是产生相同输出
2. **均匀分布**: 尽可能均匀地分布到所有桶中
3. **快速计算**: 计算开销要小
4. **雪崩效应**: 输入微小变化导致输出大幅变化

**常见哈希算法**:
```go
// 简单乘法哈希
func simpleHash(key string, size int) int {
    hash := 0
    for _, c := range key {
        hash = hash*31 + int(c)
    }
    return hash % size
}

// FNV-1a哈希 (Redis使用的算法之一)
func fnvHash(data []byte) uint32 {
    hash := uint32(2166136261)
    for _, b := range data {
        hash ^= uint32(b)
        hash *= 16777619
    }
    return hash
}
```

---

## 🔧 哈希冲突处理

### 1. 链表法 (Separate Chaining)

**Redis采用的方法！**

```
Bucket 0: [key1, val1] → [key5, val5] → nil
Bucket 1: [key2, val2] → nil  
Bucket 2: [key3, val3] → [key7, val7] → [key9, val9] → nil
```

**优点**:
- 实现简单
- 删除操作容易
- 负载因子可以超过1
- 对哈希函数质量要求相对较低

**缺点**:
- 需要额外指针空间
- 缓存不友好
- 最坏情况O(n)

### 2. 开放寻址法 (Open Addressing)

**线性探测**:
```go
func linearProbe(hash, i, size int) int {
    return (hash + i) % size
}
```

**二次探测**:
```go  
func quadraticProbe(hash, i, size int) int {
    return (hash + i*i) % size
}
```

**双重哈希**:
```go
func doubleHash(hash1, hash2, i, size int) int {
    return (hash1 + i*hash2) % size
}
```

---

## 📈 动态扩容策略

### 1. 负载因子 (Load Factor)

```
负载因子 = 已存储元素数量 / 桶总数
```

**关键阈值**:
- **扩容阈值**: 通常是 0.75
- **缩容阈值**: 通常是 0.25

### 2. 扩容过程

**传统方式**:
```go
func resize(oldTable []Bucket, newSize int) []Bucket {
    newTable := make([]Bucket, newSize)
    
    // 重新哈希所有元素
    for _, bucket := range oldTable {
        for node := bucket.head; node != nil; node = node.next {
            newIndex := hash(node.key) % newSize
            // 插入到新表
            insertToBucket(&newTable[newIndex], node.key, node.value)
        }
    }
    
    return newTable
}
```

**问题**: 一次性rehash会导致长时间阻塞！

### 3. 渐进式rehash (Redis的解决方案)

**Redis的精妙设计**:
```go
type HashTable struct {
    table     [2][]Bucket  // 两个哈希表
    size      [2]int       // 两个表的大小
    used      [2]int       // 两个表的使用量
    rehashing bool         // 是否在rehash中
    rehashIdx int          // rehash进度索引
}
```

**渐进式过程**:
1. **触发**: 负载因子达到阈值
2. **分步**: 每次操作时迁移一小部分数据
3. **完成**: 所有数据迁移完毕，切换表

---

## 🏗️ Redis Set的设计选择

### 1. 小集合优化

**IntSet编码** (元素都是整数且数量少时):
```go
type IntSet struct {
    encoding int32    // 编码类型：int16/int32/int64
    length   int32    // 元素数量
    contents []byte   // 实际存储数组
}
```

**优势**:
- 内存紧凑
- 缓存友好
- 节省指针开销

### 2. 编码转换

```
小整数集合 → IntSet编码
↓ (元素增多或包含非整数)
哈希表编码 → HashTable编码
```

### 3. 集合运算优化

**并集 (UNION)**:
```go
func union(set1, set2 *RedisSet) *RedisSet {
    // 选择较小集合作为迭代对象
    if set1.size > set2.size {
        set1, set2 = set2, set1
    }
    
    result := copySet(set2)  // 复制大集合
    // 遍历小集合，添加到结果中
    for element := range set1 {
        result.Add(element)
    }
    return result
}
```

---

## 🧠 关键设计决策

### 1. 为什么Redis选择链表法？

**优势**:
- **删除简单**: 不需要特殊标记
- **扩容灵活**: 负载因子可以超过1
- **实现清晰**: 代码逻辑简单明了

**权衡**:
- 牺牲一些缓存性能
- 换取实现简单性和灵活性

### 2. 为什么需要渐进式rehash？

**问题**: 传统rehash的阻塞问题
```
数据量: 100万条记录
rehash时间: 100ms+
影响: 所有操作被阻塞
```

**解决**: 分摊到多次操作
```
每次操作: 迁移1-100个桶
总时间: 分散到数千次操作
影响: 每次操作延迟+1ms
```

---

## 💡 学习重点

### 1. 核心概念
- [x] 哈希函数的作用和特点
- [x] 冲突处理的不同方法
- [x] 负载因子和动态扩容
- [x] 渐进式rehash的必要性

### 2. 实现技巧
- [ ] 设计好的哈希函数
- [ ] 实现链表法冲突处理
- [ ] 控制负载因子
- [ ] 实现基本的集合运算

### 3. 性能优化
- [ ] 小集合的特殊编码
- [ ] 批量操作的优化
- [ ] 内存布局的考虑

---

## 🎯 下一步行动

1. **理解概念**: 确保掌握所有基础概念
2. **设计接口**: 定义RedisSet的API
3. **选择实现**: 从简单的链表法哈希表开始
4. **逐步优化**: 添加小集合优化和扩容机制

**准备好开始实现了吗？** 我们可以从最基础的链表法哈希表开始！