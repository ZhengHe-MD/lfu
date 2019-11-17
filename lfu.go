package lfu

import (
	"container/list"
	"sync"
)

type LFU interface {
	Set(k string, v interface{})
	Get(k string) (v interface{}, ok bool)
	Evict(n int)
	Size() int
}

func New(cap int) LFU {
	return &cache{
		cap:      cap,
		kv:       make(map[string]*kvItem),
		freqList: list.New(),
	}
}

var (
	placeholder = struct{}{}
)

type cache struct {
	sync.Mutex

	cap      int
	kv       map[string]*kvItem
	freqList *list.List
}

type kvItem struct {
	k      string
	v      interface{}
	parent *list.Element
}

type freqNode struct {
	freq  int
	items map[*kvItem]interface{}
}

func (c *cache) Set(k string, v interface{}) {
	if c.cap > 0 && len(c.kv) >= c.cap {
		c.Evict(1)
	}

	c.Lock()
	defer c.Unlock()

	var item *kvItem

	if item, ok := c.kv[k]; ok {
		item.v = v
		c.increment(item)
		return
	}

	front := c.freqList.Front()
	if c.freqList.Len() == 0 || front.Value.(*freqNode).freq != 1 {
		node := &freqNode{
			freq:  1,
			items: map[*kvItem]interface{}{},
		}

		c.freqList.PushFront(node)

		item = &kvItem{
			k:      k,
			v:      v,
			parent: c.freqList.Front(),
		}

		node.items[item] = placeholder
	} else {
		item = &kvItem{
			k:      k,
			v:      v,
			parent: front,
		}

		front.Value.(*freqNode).items[item] = placeholder
	}
	c.kv[k] = item
	return
}

func (c *cache) Get(k string) (vv interface{}, ok bool) {
	c.Lock()
	defer c.Unlock()

	v, ok := c.kv[k]
	if !ok {
		return
	}

	vv = v.v

	c.increment(v)
	return
}

func (c *cache) Evict(n int) {
	c.Lock()
	defer c.Unlock()

	if n <= 0 {
		return
	}

	i := 0

	for {
		if i == n || c.freqList.Len() == 0 {
			break
		}

		front := c.freqList.Front()
		frontNode := front.Value.(*freqNode)

		for item, _ := range frontNode.items {
			delete(c.kv, item.k)
			delete(frontNode.items, item)
			i += 1
			if i == n {
				break
			}
		}

		if len(frontNode.items) == 0 {
			c.freqList.Remove(front)
		}
	}
	return
}

func (c *cache) Size() int {
	c.Lock()
	defer c.Unlock()
	return len(c.kv)
}

func (c *cache) increment(item *kvItem) {
	curr := item.parent
	currNode := curr.Value.(*freqNode)

	next := curr.Next()
	var nextNode *freqNode
	if next != nil {
		nextNode = next.Value.(*freqNode)
	}

	if next == nil || (currNode.freq+1 != nextNode.freq) {
		node := &freqNode{
			freq: currNode.freq + 1,
			items: map[*kvItem]interface{}{
				item: placeholder,
			},
		}
		c.freqList.InsertAfter(node, curr)
	} else {
		nextNode.items[item] = placeholder
	}

	item.parent = curr.Next()

	// remove kvItem from current freq node
	delete(currNode.items, item)
	if len(currNode.items) == 0 {
		c.freqList.Remove(curr)
	}

	return
}

