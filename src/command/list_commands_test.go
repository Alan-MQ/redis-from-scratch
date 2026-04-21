package command

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"

	"redis-from-scratch/src/core"
	"redis-from-scratch/src/storage"
)

func TestHandlerListPushPopAndRange(t *testing.T) {
	handler := NewHandler(storage.NewEngine())

	result, err := handler.Execute([]string{"LPUSH", "tasks", "b", "a"})
	if errors.Is(err, ErrLPushCommandNotImplemented) {
		t.Skip("TODO: Alan 需要实现 LPUSH/RPUSH/LPOP/RPOP/LRANGE")
		return
	}
	assert.NoError(t, err)
	assert.Equal(t, ":2\r\n", string(result.Encode()))

	result, err = handler.Execute([]string{"RPUSH", "tasks", "c"})
	if errors.Is(err, ErrRPushCommandNotImplemented) {
		t.Skip("TODO: Alan 需要实现 RPUSH")
		return
	}
	assert.NoError(t, err)
	assert.Equal(t, ":3\r\n", string(result.Encode()))

	result, err = handler.Execute([]string{"LRANGE", "tasks", "0", "-1"})
	if errors.Is(err, ErrLRangeCommandNotImplemented) {
		t.Skip("TODO: Alan 需要实现 LRANGE")
		return
	}
	assert.NoError(t, err)
	assert.Equal(t, "*3\r\n$1\r\na\r\n$1\r\nb\r\n$1\r\nc\r\n", string(result.Encode()))

	result, err = handler.Execute([]string{"LPOP", "tasks"})
	if errors.Is(err, ErrLPopCommandNotImplemented) {
		t.Skip("TODO: Alan 需要实现 LPOP")
		return
	}
	assert.NoError(t, err)
	assert.Equal(t, "$1\r\na\r\n", string(result.Encode()))

	result, err = handler.Execute([]string{"RPOP", "tasks"})
	if errors.Is(err, ErrRPopCommandNotImplemented) {
		t.Skip("TODO: Alan 需要实现 RPOP")
		return
	}
	assert.NoError(t, err)
	assert.Equal(t, "$1\r\nc\r\n", string(result.Encode()))
}

func TestHandlerListCommandsWrongType(t *testing.T) {
	engine := storage.NewEngine()
	engine.Set("name", core.NewSDS("alan"))
	handler := NewHandler(engine)

	result, err := handler.Execute([]string{"LPUSH", "name", "x"})
	if errors.Is(err, ErrLPushCommandNotImplemented) {
		t.Skip("TODO: Alan 需要实现 LPUSH/RPUSH/LPOP/RPOP/LRANGE")
		return
	}
	assert.NoError(t, err)
	assert.Equal(t, "-WRONGTYPE Operation against a key holding the wrong kind of value\r\n", string(result.Encode()))
}
