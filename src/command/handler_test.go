package command

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"

	"redis-from-scratch/src/core"
	"redis-from-scratch/src/storage"
)

func TestHandlerExecutePing(t *testing.T) {
	handler := NewHandler(storage.NewEngine())

	result, err := handler.Execute([]string{"PING"})
	assert.NoError(t, err)
	assert.Equal(t, "+PONG\r\n", string(result.Encode()))
}

func TestHandlerExecutePingWithMessage(t *testing.T) {
	handler := NewHandler(storage.NewEngine())

	result, err := handler.Execute([]string{"PING", "hello"})
	assert.NoError(t, err)
	assert.Equal(t, "$5\r\nhello\r\n", string(result.Encode()))
}

func TestHandlerUnknownCommand(t *testing.T) {
	handler := NewHandler(storage.NewEngine())

	result, err := handler.Execute([]string{"NOPE"})
	assert.NoError(t, err)
	assert.Equal(t, "-ERR unknown command 'nope'\r\n", string(result.Encode()))
}

func TestHandlerSetGetDelete(t *testing.T) {
	handler := NewHandler(storage.NewEngine())

	result, err := handler.Execute([]string{"SET", "name", "alan"})
	if errors.Is(err, ErrSetCommandNotImplemented) {
		t.Skip("TODO: Alan 需要实现 SET/GET/DEL 命令处理")
		return
	}
	assert.NoError(t, err)
	assert.Equal(t, "+OK\r\n", string(result.Encode()))

	result, err = handler.Execute([]string{"GET", "name"})
	if errors.Is(err, ErrGetCommandNotImplemented) {
		t.Skip("TODO: Alan 需要实现 GET 命令处理")
		return
	}
	assert.NoError(t, err)
	assert.Equal(t, "$4\r\nalan\r\n", string(result.Encode()))

	result, err = handler.Execute([]string{"DEL", "name"})
	if errors.Is(err, ErrDelCommandNotImplemented) {
		t.Skip("TODO: Alan 需要实现 DEL 命令处理")
		return
	}
	assert.NoError(t, err)
	assert.Equal(t, ":1\r\n", string(result.Encode()))
}

func TestHandlerGetWrongType(t *testing.T) {
	engine := storage.NewEngine()
	engine.Set("tasks", core.NewRedisList())
	handler := NewHandler(engine)

	result, err := handler.Execute([]string{"GET", "tasks"})
	assert.NoError(t, err)
	assert.Equal(t, "-WRONGTYPE Operation against a key holding the wrong kind of value\r\n", string(result.Encode()))
}
