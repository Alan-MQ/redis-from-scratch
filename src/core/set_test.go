package core

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewRedisSet(t *testing.T) {
	set := NewRedisSet()

	assert.NotNil(t, set)
	assert.Equal(t, SetType, set.Type())
	assert.Equal(t, 0, set.SCard())
	assert.Equal(t, "{}", set.String())
}

func TestSetBasicOperations(t *testing.T) {
	set := NewRedisSet()

	added := set.SAdd("apple", "banana", "apple")
	assert.Equal(t, 2, added)
	assert.Equal(t, 2, set.SCard())

	assert.True(t, set.SIsMember("apple"))
	assert.True(t, set.SIsMember("banana"))
	assert.False(t, set.SIsMember("orange"))

	removed := set.SRem("banana", "missing")
	assert.Equal(t, 1, removed)
	assert.False(t, set.SIsMember("banana"))

	assert.ElementsMatch(t, []string{"apple"}, set.SMembers())
}

func TestSetWorksWhileUnderlyingDictIsRehashing(t *testing.T) {
	set := &RedisSet{
		dict: NewDictWithCapacity(2),
	}

	assert.Equal(t, 1, set.SAdd("alpha"))
	assert.Equal(t, 1, set.SAdd("beta"))
	assert.True(t, set.dict.IsRehashing())

	assert.Equal(t, 1, set.SAdd("gamma"))
	assert.True(t, set.SIsMember("alpha"))
	assert.True(t, set.SIsMember("gamma"))
	assert.ElementsMatch(t, []string{"alpha", "beta", "gamma"}, set.SMembers())
}
