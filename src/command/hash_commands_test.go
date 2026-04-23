package command

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"

	"redis-from-scratch/src/core"
	"redis-from-scratch/src/storage"
)

func TestHandlerHashCommands(t *testing.T) {
	handler := NewHandler(storage.NewEngine())

	result, err := handler.Execute([]string{"HSET", "profile", "name", "alan", "city", "hangzhou"})
	if errors.Is(err, ErrHSetCommandNotImplemented) {
		t.Skip("TODO: Alan 需要实现 HSET/HGET/HDEL/HEXISTS/HGETALL/HLEN")
		return
	}
	assert.NoError(t, err)
	assert.Equal(t, ":2\r\n", string(result.Encode()))

	result, err = handler.Execute([]string{"HGET", "profile", "name"})
	if errors.Is(err, ErrHGetCommandNotImplemented) {
		t.Skip("TODO: Alan 需要实现 HGET")
		return
	}
	assert.NoError(t, err)
	assert.Equal(t, "$4\r\nalan\r\n", string(result.Encode()))

	result, err = handler.Execute([]string{"HEXISTS", "profile", "city"})
	if errors.Is(err, ErrHExistsCommandNotImplemented) {
		t.Skip("TODO: Alan 需要实现 HEXISTS")
		return
	}
	assert.NoError(t, err)
	assert.Equal(t, ":1\r\n", string(result.Encode()))

	result, err = handler.Execute([]string{"HLEN", "profile"})
	if errors.Is(err, ErrHLenCommandNotImplemented) {
		t.Skip("TODO: Alan 需要实现 HLEN")
		return
	}
	assert.NoError(t, err)
	assert.Equal(t, ":2\r\n", string(result.Encode()))

	result, err = handler.Execute([]string{"HGETALL", "profile"})
	if errors.Is(err, ErrHGetAllCommandNotImplemented) {
		t.Skip("TODO: Alan 需要实现 HGETALL")
		return
	}
	assert.NoError(t, err)
	assert.Equal(t, "*4\r\n$4\r\ncity\r\n$8\r\nhangzhou\r\n$4\r\nname\r\n$4\r\nalan\r\n", string(result.Encode()))

	result, err = handler.Execute([]string{"HDEL", "profile", "city"})
	if errors.Is(err, ErrHDelCommandNotImplemented) {
		t.Skip("TODO: Alan 需要实现 HDEL")
		return
	}
	assert.NoError(t, err)
	assert.Equal(t, ":1\r\n", string(result.Encode()))
}

func TestHandlerHashCommandsWrongType(t *testing.T) {
	engine := storage.NewEngine()
	engine.Set("tasks", core.NewRedisList())
	handler := NewHandler(engine)

	result, err := handler.Execute([]string{"HSET", "tasks", "name", "alan"})
	if errors.Is(err, ErrHSetCommandNotImplemented) {
		t.Skip("TODO: Alan 需要实现 HSET/HGET/HDEL/HEXISTS/HGETALL/HLEN")
		return
	}
	assert.NoError(t, err)
	assert.Equal(t, "-WRONGTYPE Operation against a key holding the wrong kind of value\r\n", string(result.Encode()))
}

func TestHandlerHashCommandsMissingKey(t *testing.T) {
	handler := NewHandler(storage.NewEngine())

	result, err := handler.Execute([]string{"HGET", "missing", "name"})
	if errors.Is(err, ErrHGetCommandNotImplemented) {
		t.Skip("TODO: Alan 需要实现 HGET/HGETALL/HLEN")
		return
	}
	assert.NoError(t, err)
	assert.Equal(t, "$-1\r\n", string(result.Encode()))

	result, err = handler.Execute([]string{"HGETALL", "missing"})
	if errors.Is(err, ErrHGetAllCommandNotImplemented) {
		t.Skip("TODO: Alan 需要实现 HGETALL")
		return
	}
	assert.NoError(t, err)
	assert.Equal(t, "*0\r\n", string(result.Encode()))

	result, err = handler.Execute([]string{"HLEN", "missing"})
	if errors.Is(err, ErrHLenCommandNotImplemented) {
		t.Skip("TODO: Alan 需要实现 HLEN")
		return
	}
	assert.NoError(t, err)
	assert.Equal(t, ":0\r\n", string(result.Encode()))
}
