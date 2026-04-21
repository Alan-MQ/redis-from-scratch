package storage

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"redis-from-scratch/src/core"
)

func TestEngineSetGetDelete(t *testing.T) {
	engine := NewEngine()

	engine.Set("name", core.NewSDS("alan"))

	obj := engine.Get("name")
	assert.NotNil(t, obj)
	assert.Equal(t, "alan", obj.String())
	assert.Equal(t, core.StringType, obj.Type())

	deleted := engine.Delete("name")
	assert.Equal(t, 1, deleted)
	assert.Nil(t, engine.Get("name"))
}

func TestEngineKeysAndSize(t *testing.T) {
	engine := NewEngine()

	engine.Set("alpha", core.NewSDS("A"))
	engine.Set("beta", core.NewSDS("B"))

	assert.Equal(t, 2, engine.Size())
	assert.ElementsMatch(t, []string{"alpha", "beta"}, engine.Keys())
}
