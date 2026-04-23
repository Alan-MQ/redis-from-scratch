package core

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewRedisHash(t *testing.T) {
	hash := NewRedisHash()

	assert.NotNil(t, hash)
	assert.Equal(t, HashType, hash.Type())
	assert.Equal(t, 0, hash.HLen())
	assert.Equal(t, "{}", hash.String())
	assert.Empty(t, hash.HGetAll())
}

func TestHashSetGetUpdateDelete(t *testing.T) {
	hash := NewRedisHash()

	inserted := hash.HSet("name", "alan")
	assert.True(t, inserted)
	assert.Equal(t, 1, hash.HLen())

	value, ok := hash.HGet("name")
	assert.True(t, ok)
	assert.Equal(t, "alan", value)

	inserted = hash.HSet("name", "redis")
	assert.False(t, inserted)
	assert.Equal(t, 1, hash.HLen())

	value, ok = hash.HGet("name")
	assert.True(t, ok)
	assert.Equal(t, "redis", value)

	assert.True(t, hash.HExists("name"))
	assert.False(t, hash.HExists("missing"))

	deleted := hash.HDel("name", "missing")
	assert.Equal(t, 1, deleted)
	assert.Equal(t, 0, hash.HLen())
}

func TestHashGetAllIsStable(t *testing.T) {
	hash := NewRedisHash()

	hash.HSet("name", "alan")
	hash.HSet("city", "hangzhou")
	hash.HSet("role", "engineer")

	assert.Equal(t, []HashEntry{
		{Field: "city", Value: "hangzhou"},
		{Field: "name", Value: "alan"},
		{Field: "role", Value: "engineer"},
	}, hash.HGetAll())
}
