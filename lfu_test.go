package lfu

import (
	"container/list"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestLFU(t *testing.T) {
	lfu := New(2)

	lfu.Set("a", 1)
	lfu.Set("b", 2)
	v, ok := lfu.Get("a")
	assert.True(t, ok)
	assert.Equal(t, 1, v)
	lfu.Evict(1)
	v, ok = lfu.Get("a")
	assert.True(t, ok)
	assert.Equal(t, 1, v)
	v, ok = lfu.Get("b")
	assert.False(t, ok)
	assert.Nil(t, v)
	lfu.Set("c", 3)
	v, ok = lfu.Get("c")
	assert.True(t, ok)
	assert.Equal(t, 3, v)
	lfu.Set("d", 4)
	v, ok = lfu.Get("c")
	assert.False(t, ok)
	assert.Nil(t, v)
	v, ok = lfu.Get("d")
	assert.True(t, ok)
	assert.Equal(t, 4, v)
}

func TestCache_Set(t *testing.T) {
	cache := &cache{
		cap:      2,
		kv:       make(map[string]*kvItem),
		freqList: list.New(),
	}

	// set "a"
	cache.Set("a", 1)
	assert.Equal(t, 1, cache.kv["a"].v)
	assert.Equal(t, 1, cache.freqList.Len())
	frontNode := cache.freqList.Front().Value.(*freqNode)
	assert.Equal(t, 1, frontNode.freq)
	assert.Equal(t, 1, len(frontNode.items))
	_, ok := frontNode.items[cache.kv["a"]]
	assert.True(t, ok)

	// set "a" again
	cache.Set("a", 2)
	assert.Equal(t, 2, cache.kv["a"].v)
	assert.Equal(t, 1, cache.freqList.Len())
	frontNode = cache.freqList.Front().Value.(*freqNode)
	assert.Equal(t, 2, frontNode.freq)
	assert.Equal(t, 1, len(frontNode.items))
	_, ok = frontNode.items[cache.kv["a"]]
	assert.True(t, ok)

	// set "b"
	cache.Set("b", 1)
	assert.Equal(t, 1, cache.kv["b"].v)
	assert.Equal(t, 2, cache.freqList.Len())
	frontNode = cache.freqList.Front().Value.(*freqNode)
	assert.Equal(t, 1, frontNode.freq)
	assert.Equal(t, 1, len(frontNode.items))
	_, ok = frontNode.items[cache.kv["b"]]
	assert.True(t, ok)
	nextNode := cache.freqList.Front().Next().Value.(*freqNode)
	assert.Equal(t, 2, nextNode.freq)
	assert.Equal(t, 1, len(nextNode.items))
	_, ok = nextNode.items[cache.kv["a"]]
	assert.True(t, ok)

	// set "c" should evict "b"
	cache.Set("c", 1)
	assert.Equal(t, 2, len(cache.kv))
	assert.Equal(t, 2, cache.freqList.Len())
	frontNode = cache.freqList.Front().Value.(*freqNode)
	assert.Equal(t, 1, frontNode.freq)
	assert.Equal(t, 1, len(frontNode.items))
	_, ok = frontNode.items[cache.kv["c"]]
	assert.True(t, ok)
}

func TestCache_Get(t *testing.T) {
	cache := &cache{
		kv:       make(map[string]*kvItem),
		freqList: list.New(),
	}

	v, ok := cache.Get("c")
	assert.False(t, ok)
	assert.Nil(t, v)
	assert.Equal(t, 0, cache.freqList.Len())
	assert.Equal(t, 0, len(cache.kv))

	cache.Set("a", 1)
	cache.Set("b", 2)

	v, ok = cache.Get("a")
	assert.True(t, ok)
	assert.Equal(t, 1, v)
	frontNode := cache.freqList.Front().Value.(*freqNode)
	nextNode := cache.freqList.Front().Next().Value.(*freqNode)
	assert.Equal(t, 1, frontNode.freq)
	assert.Equal(t, 1, len(frontNode.items))
	assert.Equal(t, frontNode, cache.kv["b"].parent.Value.(*freqNode))
	_, ok = frontNode.items[cache.kv["b"]]
	assert.True(t, ok)
	assert.Equal(t, 2, nextNode.freq)
	assert.Equal(t, 1, len(nextNode.items))
	_, ok = nextNode.items[cache.kv["a"]]
	assert.True(t, ok)
	assert.Equal(t, nextNode, cache.kv["a"].parent.Value.(*freqNode))

	v, ok = cache.Get("b")
	assert.True(t, ok)
	assert.Equal(t, 2, v)
	assert.Equal(t, 1, cache.freqList.Len())
	frontNode = cache.freqList.Front().Value.(*freqNode)
	assert.Equal(t, 2, frontNode.freq)
	assert.Equal(t, 2, len(frontNode.items))
	_, ok = frontNode.items[cache.kv["a"]]
	assert.True(t, ok)
	_, ok = frontNode.items[cache.kv["b"]]
	assert.True(t, ok)
	assert.Equal(t, frontNode, cache.kv["a"].parent.Value.(*freqNode))
	assert.Equal(t, nextNode, cache.kv["b"].parent.Value.(*freqNode))

	v, ok = cache.Get("c")
	assert.False(t, ok)
	assert.Nil(t, v)
}

func TestCache_Evict(t *testing.T) {
	cache := &cache{
		kv:       make(map[string]*kvItem),
		freqList: list.New(),
	}
	cache.Evict(10)
	assert.Equal(t, 0, len(cache.kv))
	assert.Equal(t, 0, cache.freqList.Len())

	cache.Set("a", 1)
	cache.Set("b", 2)
	cache.Get("a")

	for _, n := range []int{-10, -1, 0} {
		cache.Evict(n)
		assert.Equal(t, 2, len(cache.kv))
		assert.Equal(t, 2, cache.freqList.Len())
	}

	cache.Evict(1)
	assert.Equal(t, 1, len(cache.kv))
	assert.Equal(t, 1, cache.freqList.Len())
	assert.Equal(t, 2, cache.freqList.Front().Value.(*freqNode).freq)
	vv, ok := cache.Get("a")
	assert.True(t, ok)
	assert.Equal(t, 1, vv)
	frontNode := cache.freqList.Front().Value.(*freqNode)
	assert.Equal(t, 3, frontNode.freq)
	assert.Equal(t, 1, len(frontNode.items))

	cache.Evict(10)
	assert.Equal(t, 0, len(cache.kv))
	assert.Equal(t, 0, cache.freqList.Len())

	cache.Set("a", 1)
	cache.Set("b", 2)
	cache.Set("c", 3)
	cache.Get("a")
	cache.Get("a")
	cache.Get("b")

	cache.Evict(3)
	assert.Equal(t, 0, len(cache.kv))
	assert.Equal(t, 0, cache.freqList.Len())
}