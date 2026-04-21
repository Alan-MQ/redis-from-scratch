package command

import (
	"errors"
	"fmt"
	"strings"

	"redis-from-scratch/src/core"
	"redis-from-scratch/src/storage"
)

const unboundedArity = int(^uint(0) >> 1)

var (
	ErrSetCommandNotImplemented = errors.New("SET command not implemented")
	ErrGetCommandNotImplemented = errors.New("GET command not implemented")
	ErrDelCommandNotImplemented = errors.New("DEL command not implemented")
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

func validateArity(name string, minArgs, maxArgs, got int) error {
	if got < minArgs || got > maxArgs {
		return fmt.Errorf("ERR wrong number of arguments for '%s' command", strings.ToLower(name))
	}

	return nil
}
