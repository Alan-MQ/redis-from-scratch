package command

import (
	"errors"
	"fmt"
	"strconv"
	"strings"
	"time"

	"redis-from-scratch/src/core"
	"redis-from-scratch/src/storage"
)

const unboundedArity = int(^uint(0) >> 1)

var (
	ErrSetCommandNotImplemented     = errors.New("SET command not implemented")
	ErrGetCommandNotImplemented     = errors.New("GET command not implemented")
	ErrDelCommandNotImplemented     = errors.New("DEL command not implemented")
	ErrLPushCommandNotImplemented   = errors.New("LPUSH command not implemented")
	ErrRPushCommandNotImplemented   = errors.New("RPUSH command not implemented")
	ErrLPopCommandNotImplemented    = errors.New("LPOP command not implemented")
	ErrRPopCommandNotImplemented    = errors.New("RPOP command not implemented")
	ErrLRangeCommandNotImplemented  = errors.New("LRANGE command not implemented")
	ErrHSetCommandNotImplemented    = errors.New("HSET command not implemented")
	ErrHGetCommandNotImplemented    = errors.New("HGET command not implemented")
	ErrHDelCommandNotImplemented    = errors.New("HDEL command not implemented")
	ErrHExistsCommandNotImplemented = errors.New("HEXISTS command not implemented")
	ErrHGetAllCommandNotImplemented = errors.New("HGETALL command not implemented")
	ErrHLenCommandNotImplemented    = errors.New("HLEN command not implemented")
	ErrExpireCommandNotImplemented  = errors.New("EXPIRE command not implemented")
	ErrTTLCommandNotImplemented     = errors.New("TTL command not implemented")
)

type commandFunc func(args []string) (Result, error)

type registration struct {
	minArgs int
	maxArgs int
	run     commandFunc
}

// Handler 负责命令路由和执行。
type Handler struct {
	engine   *storage.Engine
	commands map[string]registration
}

// NewHandler 创建命令执行器，并注册当前阶段要实现的命令。
func NewHandler(engine *storage.Engine) *Handler {
	if engine == nil {
		engine = storage.NewEngine()
	}

	handler := &Handler{
		engine:   engine,
		commands: make(map[string]registration),
	}

	handler.register("PING", 1, 2, handler.handlePing)
	handler.register("SET", 3, 3, handler.handleSet)
	handler.register("GET", 2, 2, handler.handleGet)
	handler.register("DEL", 2, unboundedArity, handler.handleDel)
	handler.register("LPUSH", 3, unboundedArity, handler.handleLPush)
	handler.register("RPUSH", 3, unboundedArity, handler.handleRPush)
	handler.register("LPOP", 2, 2, handler.handleLPop)
	handler.register("RPOP", 2, 2, handler.handleRPop)
	handler.register("LRANGE", 4, 4, handler.handleLRange)
	handler.register("HSET", 4, unboundedArity, handler.handleHSet)
	handler.register("HGET", 3, 3, handler.handleHGet)
	handler.register("HDEL", 3, unboundedArity, handler.handleHDel)
	handler.register("HEXISTS", 3, 3, handler.handleHExists)
	handler.register("HGETALL", 2, 2, handler.handleHGetAll)
	handler.register("HLEN", 2, 2, handler.handleHLen)
	handler.register("EXPIRE", 3, 3, handler.handleExpire)
	handler.register("TTL", 2, 2, handler.handleTTL)

	return handler
}

func (handler *Handler) register(name string, minArgs, maxArgs int, run commandFunc) {
	handler.commands[strings.ToUpper(name)] = registration{
		minArgs: minArgs,
		maxArgs: maxArgs,
		run:     run,
	}
}

// Execute 根据 argv 分发到对应命令。
func (handler *Handler) Execute(args []string) (Result, error) {
	if len(args) == 0 {
		return ErrorResult("ERR empty command"), nil
	}

	name := strings.ToUpper(args[0])
	command, ok := handler.commands[name]
	if !ok {
		return ErrorResult(fmt.Sprintf("ERR unknown command '%s'", strings.ToLower(args[0]))), nil
	}

	if err := validateArity(name, command.minArgs, command.maxArgs, len(args)); err != nil {
		return ErrorResult(err.Error()), nil
	}

	return command.run(args)
}

func (handler *Handler) handlePing(args []string) (Result, error) {
	if len(args) == 2 {
		return BulkStringResult(args[1]), nil
	}

	return SimpleStringResult("PONG"), nil
}

func (handler *Handler) handleSet(args []string) (Result, error) {
	if len(args) != 3 {
		return ErrorResult("ERR wrong number of arguments for 'SET' command"), nil
	}
	key := args[1]
	value := args[2]
	sdsValue := core.NewSDS(value)
	handler.engine.Set(key, sdsValue)

	return SimpleStringResult("OK"), nil
}

func (handler *Handler) handleGet(args []string) (Result, error) {
	if len(args) != 2 {
		return ErrorResult("ERR wrong number of arguments for 'GET' command"), nil
	}
	key := args[1]
	res := handler.engine.Get(key)
	if res == nil {
		return NullBulkStringResult(), nil
	}
	if res.Type() != core.StringType {
		return ErrorResult("WRONGTYPE Operation against a key holding the wrong kind of value"), nil
	}
	return BulkStringResult(res.String()), nil
}

func (handler *Handler) handleDel(args []string) (Result, error) {
	keys := args[1:]
	deleted := handler.engine.Delete(keys...)
	return IntegerResult(int64(deleted)), nil
}

func (handler *Handler) handleLPush(args []string) (Result, error) {
	// TODO: Alan 在这里实现 LPUSH。
	// 建议步骤：
	// 1. 先用 getOrCreateList(args[1]) 拿到目标列表。
	// 2. 遍历 args[2:]，把每个 string 包装成 core.NewSDS 后调用 list.LPush。
	// 3. 返回插入后的列表长度，格式是 IntegerResult(int64(list.Len()))。
	key := args[1]
	if key == "" {
		return ErrorResult("ERR key must not be empty"), nil
	}
	target, result, err := handler.getOrCreateList(key)
	if err != nil {
		return ErrorResult("ERR TODO: implement LPUSH in src/command/handler.go"), ErrLPushCommandNotImplemented
	}
	if result.kind == errorResult {
		return result, nil
	}
	for _, arg := range args[2:] {
		target.LPush(core.NewSDS(arg))
	}
	return IntegerResult(int64(target.Len())), nil
}

func (handler *Handler) handleRPush(args []string) (Result, error) {
	// TODO: Alan 在这里实现 RPUSH。
	// 思路和 LPUSH 对称，只是调用 list.RPush。
	key := args[1]
	if key == "" {
		return ErrorResult("ERR key must not be empty"), nil
	}
	target, result, err := handler.getOrCreateList(key)
	if err != nil {
		return ErrorResult("ERR TODO: implement RPUSH in src/command/handler.go"), ErrRPushCommandNotImplemented
	}
	if result.kind == errorResult {
		return result, nil
	}
	for _, arg := range args[2:] {
		target.RPush(core.NewSDS(arg))
	}
	return IntegerResult(int64(target.Len())), nil
}

func (handler *Handler) handleLPop(args []string) (Result, error) {
	// 建议步骤：
	// 1. 用 getExistingList(args[1]) 读取列表。
	// 2. key 不存在时返回 NullBulkStringResult()。
	// 3. 调用 list.LPop()，结果为 nil 时也返回 NullBulkStringResult()。
	// 4. 否则把弹出的值转成 BulkStringResult。
	key := args[1]
	if key == "" {
		return ErrorResult("ERR key must not be empty"), nil
	}
	target, result, err := handler.getExistingList(key)
	if err != nil {
		return ErrorResult("ERR getting keys"), err
	}
	if result.kind == errorResult {
		return result, nil
	}
	if target == nil {
		return NullBulkStringResult(), nil
	}

	res := target.LPop()
	if res == nil {
		return NullBulkStringResult(), nil
	}
	return BulkStringResult(res.String()), nil
}

func (handler *Handler) handleRPop(args []string) (Result, error) {
	key := args[1]
	if key == "" {
		return ErrorResult("ERR key must not be empty"), nil
	}
	target, result, err := handler.getExistingList(key)
	if err != nil {
		return ErrorResult("ERR getting keys"), err
	}
	if result.kind == errorResult {
		return result, nil
	}
	if target == nil {
		return NullBulkStringResult(), nil
	}
	res := target.RPop()
	if res == nil {
		return NullBulkStringResult(), nil
	}
	return BulkStringResult(res.String()), nil
}

func (handler *Handler) handleLRange(args []string) (Result, error) {
	// TODO: Alan 在这里实现 LRANGE。
	// 建议步骤：
	// 1. 用 getExistingList(args[1]) 读取列表。
	// 2. key 不存在时返回空数组 ArrayResult([]Result{})。
	// 3. 解析 start/stop 为整数。
	// 4. 调用 list.LRange(start, stop)。
	// 5. 把每个 RedisValue 包装成 BulkStringResult，再返回 ArrayResult(items)。
	if len(args) != 4 {
		return ErrorResult("ERR wrong number of arguments for 'LRANGE' command"), nil
	}
	key, startStr, stopStr := args[1], args[2], args[3]
	if key == "" {
		return ErrorResult("ERR key must not be empty"), nil
	}
	target, result, err := handler.getExistingList(key)
	if err != nil {
		return ErrorResult("ERR getting keys"), err
	}
	if result.kind == errorResult {
		return result, nil
	}
	if target == nil {
		return ArrayResult([]Result{}), nil
	}
	start, err := strconv.Atoi(startStr)
	if err != nil {
		return ErrorResult("ERR start is not an integer"), nil
	}
	stop, err := strconv.Atoi(stopStr)
	if err != nil {
		return ErrorResult("ERR stop is not an integer"), nil
	}
	res := target.LRange(start, stop)
	items := make([]Result, 0, len(res))
	for _, v := range res {
		items = append(items, BulkStringResult(v.String()))
	}

	return ArrayResult(items), nil
}

func (handler *Handler) handleHSet(args []string) (Result, error) {
	// TODO: Alan 在这里实现 HSET。
	// 建议步骤：
	// 1. 先校验 args[2:] 的长度必须是偶数，因为是 field/value 成对出现。
	// 2. 用 getOrCreateHash(args[1]) 拿到目标 hash。
	// 3. 逐对写入 field/value，并累计“新增字段数”。
	// 4. 返回 IntegerResult(int64(inserted)).
	key := args[1]
	if key == "" {
		return ErrorResult("ERR key must not be empty"), nil
	}
	if (len(args)-2)%2 != 0 {
		return ErrorResult("ERR wrong number of arguments for 'HSET' command"), nil
	}
	target, result, err := handler.getOrCreateHash(key)
	if err != nil {
		return ErrorResult("ERR TODO: implement HSET in src/command/handler.go"), ErrHSetCommandNotImplemented
	}
	if result.kind == errorResult {
		return result, nil
	}
	inserted := 0
	for i := 2; i < len(args); i += 2 {
		field := args[i]
		value := args[i+1]
		if target.HSet(field, value) {
			inserted++
		}
	}
	return IntegerResult(int64(inserted)), nil
}

func (handler *Handler) handleHGet(args []string) (Result, error) {
	key := args[1]
	if key == "" {
		return ErrorResult("ERR key must not be empty"), nil
	}
	target, result, err := handler.getExistingHash(key)
	if err != nil {
		return ErrorResult("ERR getting keys"), err
	}
	if result.kind == errorResult {
		return result, nil
	}
	if target == nil {
		return NullBulkStringResult(), nil
	}
	field := args[2]
	value, exists := target.HGet(field)
	if !exists {
		return NullBulkStringResult(), nil
	}
	return BulkStringResult(value), nil
}

func (handler *Handler) handleHDel(args []string) (Result, error) {
	// TODO: Alan 在这里实现 HDEL。
	// 建议步骤：
	// 1. 用 getExistingHash(args[1]) 读取 hash。
	// 2. key 不存在时返回 IntegerResult(0)。
	// 3. 调用 hash.HDel(args[2:]...) 并返回删除数量。
	key := args[1]
	if key == "" {
		return ErrorResult("ERR key must not be empty"), nil
	}
	target, result, err := handler.getExistingHash(key)
	if err != nil {
		return ErrorResult("ERR getting keys"), err
	}
	if result.kind == errorResult {
		return result, nil
	}
	if target == nil {
		return IntegerResult(0), nil
	}
	deleted := target.HDel(args[2:]...)
	return IntegerResult(int64(deleted)), nil
}

func (handler *Handler) handleHExists(args []string) (Result, error) {
	// TODO: Alan 在这里实现 HEXISTS。
	// 返回 1 表示存在，0 表示不存在。
	key := args[1]
	if key == "" {
		return ErrorResult("ERR key must not be empty"), nil
	}
	target, result, err := handler.getExistingHash(key)
	if err != nil {
		return ErrorResult("ERR getting keys"), err
	}
	if result.kind == errorResult {
		return result, nil
	}
	if target == nil {
		return IntegerResult(0), nil
	}
	field := args[2]
	exists := target.HExists(field)
	if exists {
		return IntegerResult(1), nil
	}
	return IntegerResult(0), nil

}

func (handler *Handler) handleHGetAll(args []string) (Result, error) {
	hashKey := args[1]
	if hashKey == "" {
		return ErrorResult("ERR key must not be empty"), nil
	}
	target, result, err := handler.getExistingHash(hashKey)
	if err != nil {
		return ErrorResult("ERR getting keys"), err
	}
	if result.kind == errorResult {
		return result, nil
	}
	if target == nil {
		return ArrayResult([]Result{}), nil
	}
	pairs := target.HGetAll()
	items := make([]Result, 0, len(pairs)*2)
	for _, pair := range pairs {
		items = append(items, BulkStringResult(pair.Field))
		items = append(items, BulkStringResult(pair.Value))
	}

	return ArrayResult(items), nil
}

func (handler *Handler) handleHLen(args []string) (Result, error) {
	// TODO: Alan 在这里实现 HLEN。
	// key 不存在时应该返回 0。
	key := args[1]
	if key == "" {
		return ErrorResult("ERR key must not be empty"), nil
	}
	target, result, err := handler.getExistingHash(key)
	if err != nil {
		return ErrorResult("ERR getting keys"), err
	}
	if result.kind == errorResult {
		return result, nil
	}
	if target == nil {
		return IntegerResult(0), nil
	}
	return IntegerResult(int64(target.HLen())), nil
}

func (handler *Handler) handleExpire(args []string) (Result, error) {
	key, secondsStr := args[1], args[2]
	if key == "" {
		return ErrorResult("ERR key must not be empty"), nil
	}
	seconds, err := strconv.Atoi(secondsStr)
	if err != nil || seconds < 0 {
		return ErrorResult("ERR value is not an integer or out of range"), nil
	}
	obj := handler.engine.Get(key)
	if obj == nil {
		return IntegerResult(0), nil
	}
	obj.SetExpire(time.Duration(seconds) * time.Second)
	return IntegerResult(1), nil
}

func (handler *Handler) handleTTL(args []string) (Result, error) {
	key := args[1]
	if key == "" {
		return ErrorResult("ERR key must not be empty"), nil
	}
	obj := handler.engine.Get(key)
	if obj == nil {
		return IntegerResult(-2), nil
	}
	if obj.ExpireTime == nil {
		return IntegerResult(-1), nil
	}
	ttl := int64(obj.ExpireTime.Sub(time.Now()).Seconds())
	if ttl < 0 {
		return IntegerResult(-2), nil
	}
	return IntegerResult(ttl), nil
}

func (handler *Handler) getOrCreateList(key string) (*core.RedisList, Result, error) {
	obj := handler.engine.Get(key)
	if obj == nil {
		list := core.NewRedisList()
		handler.engine.Set(key, list)
		return list, Result{}, nil
	}

	list, ok := obj.Value.(*core.RedisList)
	if !ok {
		return nil, ErrorResult("WRONGTYPE Operation against a key holding the wrong kind of value"), nil
	}

	return list, Result{}, nil
}

func (handler *Handler) getExistingList(key string) (*core.RedisList, Result, error) {
	obj := handler.engine.Get(key)
	if obj == nil {
		return nil, Result{}, nil
	}

	list, ok := obj.Value.(*core.RedisList)
	if !ok {
		return nil, ErrorResult("WRONGTYPE Operation against a key holding the wrong kind of value"), nil
	}

	return list, Result{}, nil
}

func (handler *Handler) getOrCreateHash(key string) (*core.RedisHash, Result, error) {
	obj := handler.engine.Get(key)
	if obj == nil {
		hash := core.NewRedisHash()
		handler.engine.Set(key, hash)
		return hash, Result{}, nil
	}

	hash, ok := obj.Value.(*core.RedisHash)
	if !ok {
		return nil, ErrorResult("WRONGTYPE Operation against a key holding the wrong kind of value"), nil
	}

	return hash, Result{}, nil
}

func (handler *Handler) getExistingHash(key string) (*core.RedisHash, Result, error) {
	obj := handler.engine.Get(key)
	if obj == nil {
		return nil, Result{}, nil
	}

	hash, ok := obj.Value.(*core.RedisHash)
	if !ok {
		return nil, ErrorResult("WRONGTYPE Operation against a key holding the wrong kind of value"), nil
	}

	return hash, Result{}, nil
}

func validateArity(name string, minArgs, maxArgs, got int) error {
	if got < minArgs || got > maxArgs {
		return fmt.Errorf("ERR wrong number of arguments for '%s' command", strings.ToLower(name))
	}

	return nil
}
