package lfu

import (
	"container/list"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestLFU(t *testing.T) {
	cache := New(2)

	cache.Set("a", 1)
	assert.Equal(t, 1, cache.Size())
	cache.Set("b", 2)
	assert.Equal(t, 2, cache.Size())
	v, ok := cache.Get("a")
	assert.True(t, ok)
	assert.Equal(t, 1, v)
	cache.Evict(1)
	v, ok = cache.Get("a")
	assert.True(t, ok)
	assert.Equal(t, 1, v)
	v, ok = cache.Get("b")
	assert.False(t, ok)
	assert.Nil(t, v)
	cache.Set("c", 3)
	assert.Equal(t, 2, cache.Size())
	v, ok = cache.Get("c")
	assert.True(t, ok)
	assert.Equal(t, 3, v)
	cache.Set("d", 4)
	assert.Equal(t, 2, cache.Size())
	v, ok = cache.Get("c")
	assert.False(t, ok)
	assert.Nil(t, v)
	v, ok = cache.Get("d")
	assert.True(t, ok)
	assert.Equal(t, 4, v)
	cache.Evict(10)
	assert.Equal(t, 0, cache.Size())
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

func TestCache_Size(t *testing.T) {
	cache := &cache{
		kv:       make(map[string]*kvItem),
		freqList: list.New(),
	}

	assert.Equal(t, 0, cache.Size())

	cache.Set("a", 1)
	assert.Equal(t, 1, cache.Size())

	cache.Set("b", 1)
	assert.Equal(t, 2, cache.Size())

	cache.Evict(10)
	assert.Equal(t, 0, cache.Size())
}
