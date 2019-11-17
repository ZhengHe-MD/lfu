# lfu
> Thead-safe Least Frequently Used (LFU) cache replacement policy implementation with O(1) complexity

[![Build Status](https://travis-ci.org/ZhengHe-MD/lfu.svg?branch=master)](https://travis-ci.org/ZhengHe-MD/lfu)
[![Go Report Card](https://goreportcard.com/badge/github.com/ZhengHe-MD/lfu)](https://goreportcard.com/report/github.com/ZhengHe-MD/lfu)
[![Coverage Status](https://coveralls.io/repos/github/ZhengHe-MD/lfu/badge.svg?branch=master)](https://coveralls.io/github/ZhengHe-MD/lfu?branch=master)
[![godoc](https://godoc.org/github.com/ZhengHe-MD/lfu?status.svg)](https://godoc.org/github.com/ZhengHe-MD/lfu)
![GitHub release](https://img.shields.io/github/release-pre/ZhengHe-MD/lfu.svg)

## Usages

```go
capacity = 10
// create a new LFU cache with capacity
// if capacity is set to non-positive integer
// the cache won't do any eviction
cache := lfu.New(capacity)

// set k, v
cache.Set("k1", "v1")

// get v for k
v, ok := cache.Get("k1")

// evict certain number of items
cache.Evict(1)

// get current size of lfu
size := cache.Size()
```

## References

* [An O(1) algorithm for implementing the LFU cache eviction scheme](https://github.com/papers-we-love/papers-we-love/blob/master/caching/a-constant-algorithm-for-implementing-the-lfu-cache-eviction-scheme.pdf), [chinese version](https://zhenghe.gitbook.io/open-courses/papers-we-love/lfu-implementation-with-o-1-complexity-2010)

