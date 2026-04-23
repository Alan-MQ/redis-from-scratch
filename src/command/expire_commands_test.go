package command

import (
	"errors"
	"strconv"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"

	"redis-from-scratch/src/storage"
)

func TestHandlerExpireAndTTL(t *testing.T) {
	handler := NewHandler(storage.NewEngine())

	result, err := handler.Execute([]string{"SET", "session", "abc"})
	assert.NoError(t, err)
	assert.Equal(t, "+OK\r\n", string(result.Encode()))

	result, err = handler.Execute([]string{"TTL", "session"})
	if errors.Is(err, ErrTTLCommandNotImplemented) {
		t.Skip("TODO: Alan 需要实现 EXPIRE/TTL")
		return
	}
	assert.NoError(t, err)
	assert.Equal(t, ":-1\r\n", string(result.Encode()))

	result, err = handler.Execute([]string{"EXPIRE", "session", "10"})
	if errors.Is(err, ErrExpireCommandNotImplemented) {
		t.Skip("TODO: Alan 需要实现 EXPIRE")
		return
	}
	assert.NoError(t, err)
	assert.Equal(t, ":1\r\n", string(result.Encode()))

	result, err = handler.Execute([]string{"TTL", "session"})
	assert.NoError(t, err)
	ttl, parseErr := parseIntegerReply(string(result.Encode()))
	assert.NoError(t, parseErr)
	assert.True(t, ttl >= 0 && ttl <= 10, "TTL should be between 0 and 10 seconds")
}

func TestHandlerExpireAndTTLMissingKey(t *testing.T) {
	handler := NewHandler(storage.NewEngine())

	result, err := handler.Execute([]string{"EXPIRE", "missing", "10"})
	if errors.Is(err, ErrExpireCommandNotImplemented) {
		t.Skip("TODO: Alan 需要实现 EXPIRE/TTL")
		return
	}
	assert.NoError(t, err)
	assert.Equal(t, ":0\r\n", string(result.Encode()))

	result, err = handler.Execute([]string{"TTL", "missing"})
	if errors.Is(err, ErrTTLCommandNotImplemented) {
		t.Skip("TODO: Alan 需要实现 TTL")
		return
	}
	assert.NoError(t, err)
	assert.Equal(t, ":-2\r\n", string(result.Encode()))
}

func TestHandlerExpireRejectsInvalidSeconds(t *testing.T) {
	handler := NewHandler(storage.NewEngine())

	result, err := handler.Execute([]string{"SET", "session", "abc"})
	assert.NoError(t, err)
	assert.Equal(t, "+OK\r\n", string(result.Encode()))

	result, err = handler.Execute([]string{"EXPIRE", "session", "not-a-number"})
	if errors.Is(err, ErrExpireCommandNotImplemented) {
		t.Skip("TODO: Alan 需要实现 EXPIRE")
		return
	}
	assert.NoError(t, err)
	assert.Equal(t, "-ERR value is not an integer or out of range\r\n", string(result.Encode()))
}

func parseIntegerReply(reply string) (int64, error) {
	trimmed := strings.TrimSuffix(strings.TrimPrefix(reply, ":"), "\r\n")
	return strconv.ParseInt(trimmed, 10, 64)
}
