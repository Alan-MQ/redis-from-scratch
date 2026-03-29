package core

const (
	// DefaultDictBucketCount 是主哈希表的默认桶数量。
	DefaultDictBucketCount = 4

	// DictMaxLoadFactor 超过这个负载因子后，就应该开始扩容。
	DictMaxLoadFactor = 0.75
)

// DictEntry 是链表法冲突处理中的单个节点。
type DictEntry struct {
	Key   string
	Value RedisValue
	Next  *DictEntry
}

// HashTable 是 Dict 内部使用的单张哈希表。
// Redis 真正做渐进式 rehash 时，会同时维护两张表。
type HashTable struct {
	buckets []*DictEntry
	size    int
	used    int
}

func newHashTable(bucketCount int) HashTable {
	if bucketCount <= 0 {
		return HashTable{}
	}

	return HashTable{
		buckets: make([]*DictEntry, bucketCount),
		size:    bucketCount,
	}
}

// Dict 是 Redis dict 的简化版骨架：
// 1. ht[0] 是当前主表。
// 2. ht[1] 是 rehash 期间的新表。
// 3. rehashIdx == -1 表示当前不在 rehash。
type Dict struct {
	tables    [2]HashTable
	rehashIdx int
}

// NewDict 创建默认大小的字典。
func NewDict() *Dict {
	return NewDictWithCapacity(DefaultDictBucketCount)
}

// NewDictWithCapacity 按指定桶数量创建字典。
func NewDictWithCapacity(bucketCount int) *Dict {
	if bucketCount <= 0 {
		bucketCount = DefaultDictBucketCount
	}

	return &Dict{
		tables: [2]HashTable{
			newHashTable(bucketCount),
			{},
		},
		rehashIdx: -1,
	}
}

// Set 写入 key-value。
// 返回值约定：
// 1. true 表示插入了一个新 key。
// 2. false 表示更新了已有 key。
func (dict *Dict) Set(key string, value RedisValue) bool {
	if dict == nil {
		return false
	}

	if dict.IsRehashing() {
		dict.rehashStep()
	}

	// 先处理更新：rehash 期间需要检查两张表。
	if entry := dict.getEntryFromTable(0, key); entry != nil {
		entry.Value = value
		return false
	}
	if dict.IsRehashing() {
		if entry := dict.getEntryFromTable(1, key); entry != nil {
			entry.Value = value
			return false
		}
	}

	targetTable := 0
	if dict.IsRehashing() {
		targetTable = 1
	}

	inserted := dict.insertIntoTable(targetTable, key, value)

	// 基础 dict/set 你已经写过了，所以这里我直接补到
	// “能自动进入 rehash 状态”的程度。
	if !dict.IsRehashing() && dict.shouldStartRehash() {
		dict.startRehash(dict.nextExpandSize())
	}

	return inserted
}

// Get 读取 key 对应的值。
// 渐进式 rehash 期间，需要同时查旧表和新表。
func (dict *Dict) Get(key string) RedisValue {
	if dict == nil {
		return nil
	}

	if dict.IsRehashing() {
		dict.rehashStep()
	}

	if entry := dict.getEntryFromTable(0, key); entry != nil {
		return entry.Value
	}
	if dict.IsRehashing() {
		if entry := dict.getEntryFromTable(1, key); entry != nil {
			return entry.Value
		}
	}

	return nil
}

// Delete 删除指定 key。
// 渐进式 rehash 期间，可能需要从两张表里找。
func (dict *Dict) Delete(key string) bool {
	if dict == nil {
		return false
	}

	if dict.IsRehashing() {
		dict.rehashStep()
	}

	deleted := dict.deleteFromTable(0, key)
	if dict.IsRehashing() {
		deleted = dict.deleteFromTable(1, key) || deleted
	}

	return deleted
}

// Contains 判断 key 是否存在。
func (dict *Dict) Contains(key string) bool {
	return dict.Get(key) != nil
}

// Size 返回总元素个数。
func (dict *Dict) Size() int {
	if dict == nil {
		return 0
	}

	return dict.tables[0].used + dict.tables[1].used
}

// BucketCount 返回主表桶数量。
func (dict *Dict) BucketCount() int {
	return dict.TableSize(0)
}

// RehashBucketCount 返回新表桶数量。
func (dict *Dict) RehashBucketCount() int {
	return dict.TableSize(1)
}

// TableSize 返回指定哈希表的桶数量。
func (dict *Dict) TableSize(tableIndex int) int {
	table := dict.table(tableIndex)
	if table == nil {
		return 0
	}
	return table.size
}

// TableUsed 返回指定哈希表的已用节点数。
func (dict *Dict) TableUsed(tableIndex int) int {
	table := dict.table(tableIndex)
	if table == nil {
		return 0
	}
	return table.used
}

// LoadFactor 返回主表的负载因子。
func (dict *Dict) LoadFactor() float64 {
	if dict == nil || dict.tables[0].size == 0 {
		return 0
	}

	return float64(dict.tables[0].used) / float64(dict.tables[0].size)
}

// Keys 返回所有 key。
// rehash 期间需要遍历两张表。
func (dict *Dict) Keys() []string {
	if dict == nil || dict.Size() == 0 {
		return []string{}
	}

	keys := make([]string, 0, dict.Size())
	for tableIndex := 0; tableIndex < len(dict.tables); tableIndex++ {
		table := dict.table(tableIndex)
		if table == nil || table.used == 0 {
			continue
		}

		for _, head := range table.buckets {
			for entry := head; entry != nil; entry = entry.Next {
				keys = append(keys, entry.Key)
			}
		}
	}

	return keys
}

// IsRehashing 判断当前是否正在渐进式 rehash。
func (dict *Dict) IsRehashing() bool {
	return dict != nil && dict.rehashIdx >= 0
}

// RehashIndex 返回当前迁移进度。
// -1 表示当前不在 rehash。
func (dict *Dict) RehashIndex() int {
	if dict == nil {
		return -1
	}
	return dict.rehashIdx
}

func (dict *Dict) shouldStartRehash() bool {
	return !dict.IsRehashing() && dict.LoadFactor() > DictMaxLoadFactor
}

func (dict *Dict) nextExpandSize() int {
	current := dict.TableSize(0)
	if current < DefaultDictBucketCount {
		return DefaultDictBucketCount
	}
	return current * 2
}

// startRehash 初始化第二张表，并把 rehashIdx 设为 0。
// 这一步是“进入渐进式 rehash 状态”的起点。
func (dict *Dict) startRehash(newBucketCount int) bool {
	if dict == nil || dict.IsRehashing() {
		return false
	}

	if newBucketCount <= dict.TableSize(0) {
		newBucketCount = dict.nextExpandSize()
	}

	dict.tables[1] = newHashTable(newBucketCount)
	dict.rehashIdx = 0
	return true
}

// rehashStep 每次迁移一个 bucket。
// 这是你现在真正要练的核心逻辑。
func (dict *Dict) rehashStep() bool {
	if dict == nil || !dict.IsRehashing() {
		return false
	}

	// TODO: Alan 实现渐进式 rehash 的单步迁移
	// 建议步骤：
	// 1. 从 rehashIdx 开始，跳过空 bucket。
	// 2. 取出当前 bucket 的整条链表。
	// 3. 逐个节点迁移到 tables[1]。
	//    注意：迁移前先保存 next := entry.Next，再改 entry.Next。
	// 4. tables[0] 当前 bucket 清空，tables[0].used 递减，tables[1].used 递增。
	// 5. rehashIdx++。
	// 6. 如果 tables[0].used == 0，调用 finishRehash()。
	for dict.rehashIdx < dict.tables[0].size {
		old := dict.tables[0].buckets[dict.rehashIdx]
		if old == nil {
			dict.rehashIdx++
			continue
		}

		// 先把旧表当前 bucket 断开，避免迁移时同一条链同时挂在两张表上。
		dict.tables[0].buckets[dict.rehashIdx] = nil

		moved := 0
		for entry := old; entry != nil; {
			next := entry.Next
			dict.moveEntryToTable(1, entry)
			moved++
			entry = next
		}

		dict.tables[0].used -= moved
		dict.rehashIdx++

		if dict.tables[0].used == 0 {
			dict.finishRehash()
		}

		return true
	}

	if dict.tables[0].used == 0 {
		dict.finishRehash()
	}

	return false
}

// finishRehash 在迁移完成后，把新表切换为主表。
func (dict *Dict) finishRehash() {
	if dict == nil {
		return
	}

	dict.tables[0] = dict.tables[1]
	dict.tables[1] = HashTable{}
	dict.rehashIdx = -1
}

func (dict *Dict) table(tableIndex int) *HashTable {
	if dict == nil || tableIndex < 0 || tableIndex >= len(dict.tables) {
		return nil
	}
	return &dict.tables[tableIndex]
}

func (dict *Dict) insertIntoTable(tableIndex int, key string, value RedisValue) bool {
	table := dict.table(tableIndex)
	if table == nil || table.size == 0 {
		return false
	}

	index := dict.hash(key, table.size)
	for entry := table.buckets[index]; entry != nil; entry = entry.Next {
		if entry.Key == key {
			entry.Value = value
			return false
		}
	}

	table.buckets[index] = &DictEntry{
		Key:   key,
		Value: value,
		Next:  table.buckets[index],
	}
	table.used++
	return true
}

func (dict *Dict) getEntryFromTable(tableIndex int, key string) *DictEntry {
	table := dict.table(tableIndex)
	if table == nil || table.size == 0 {
		return nil
	}

	index := dict.hash(key, table.size)
	for entry := table.buckets[index]; entry != nil; entry = entry.Next {
		if entry.Key == key {
			return entry
		}
	}

	return nil
}

func (dict *Dict) deleteFromTable(tableIndex int, key string) bool {
	table := dict.table(tableIndex)
	if table == nil || table.size == 0 {
		return false
	}

	index := dict.hash(key, table.size)
	var prev *DictEntry
	for entry := table.buckets[index]; entry != nil; entry = entry.Next {
		if entry.Key == key {
			if prev == nil {
				table.buckets[index] = entry.Next
			} else {
				prev.Next = entry.Next
			}
			table.used--
			return true
		}
		prev = entry
	}

	return false
}

// moveEntryToTable 会把一个“已经存在的节点”挂到目标表上。
// 这个 helper 是为 rehashStep 准备的，方便你专心写迁移流程。
func (dict *Dict) moveEntryToTable(tableIndex int, entry *DictEntry) {
	table := dict.table(tableIndex)
	if table == nil || table.size == 0 || entry == nil {
		return
	}

	index := dict.hash(entry.Key, table.size)
	entry.Next = table.buckets[index]
	table.buckets[index] = entry
	table.used++
}

// hash 计算 key 在指定桶数量下的下标。
// 这里直接给你一个稳定的 FNV-1a 风格实现，避免把精力浪费在重复题上。
func (dict *Dict) hash(key string, bucketCount int) int {
	if bucketCount <= 0 {
		return 0
	}

	var hash uint32 = 2166136261
	for i := 0; i < len(key); i++ {
		hash ^= uint32(key[i])
		hash *= 16777619
	}

	return int(hash % uint32(bucketCount))
}
