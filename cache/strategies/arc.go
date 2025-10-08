package strategies

import (
	"container/list"
	"github.com/kimvlry/caching/cache"
	"github.com/kimvlry/caching/cache/strategies/common"
)

// ARCCache implements Adaptive Replacement Cache algorithm
type ARCCache[K comparable, V any] struct {
	capacity int
	p        int // adaptive parameter (target size for T1)

	t1 *cacheList[K, V] // recently used once
	t2 *cacheList[K, V] // frequently used
	b1 *cacheList[K, V] // ghost list for T1
	b2 *cacheList[K, V] // ghost list for T2

	eventCallbacks []func(cache.Event[K, V])
}

type ghostEntry[K comparable] struct {
	key K
}

type cacheList[K comparable, V any] struct {
	l       *list.List
	m       map[K]*list.Element
	isGhost bool
}

func newArcCache[K comparable, V any](capacity int) cache.IterableCache[K, V] {
	if capacity <= 0 {
		capacity = 1
	}
	return &ARCCache[K, V]{
		capacity: capacity,
		t1:       newCacheList[K, V](false),
		t2:       newCacheList[K, V](false),
		b1:       newCacheList[K, V](true),
		b2:       newCacheList[K, V](true),
	}
}

func (a *ARCCache[K, V]) Get(key K) (V, error) {
	if elem, ok := a.t1.m[key]; ok {
		e := elem.Value.(*entry[K, V])
		a.t1.remove(key)
		a.t2.addFront(key, e.value)
		return e.value, nil
	}
	if elem, ok := a.t2.m[key]; ok {
		a.t2.l.MoveToFront(elem)
		e := elem.Value.(*entry[K, V])
		return e.value, nil
	}
	var zero V
	return zero, common.ErrKeyNotFound
}

func (a *ARCCache[K, V]) Set(key K, value V) error {
	switch {
	case a.t1.m[key] != nil:
		a.t1.remove(key)
		a.t2.addFront(key, value)
		return nil
	case a.t2.m[key] != nil:
		a.t2.moveToFront(key, value)
		return nil
	case a.b1.m[key] != nil:
		a.adapt(true)
		a.b1.remove(key)
		a.replace(key, true)
		a.t2.addFront(key, value)
		return nil
	case a.b2.m[key] != nil:
		a.adapt(false)
		a.b2.remove(key)
		a.replace(key, false)
		a.t2.addFront(key, value)
		return nil
	default:
		if a.t1.len()+a.t2.len() >= a.capacity {
			a.replace(key, false)
		}
		a.t1.addFront(key, value)
		return nil
	}
}

// replace decides which list to evict from
func (a *ARCCache[K, V]) replace(key K, inB1 bool) {
	if a.t1.len() > 0 && (a.t1.len() > a.p || (inB1 && a.t1.len() == a.p)) {
		a.evictFromList(a.t1, a.b1)
	} else {
		a.evictFromList(a.t2, a.b2)
	}
}

func (a *ARCCache[K, V]) evictFromList(lru, ghost *cacheList[K, V]) {
	elem := lru.l.Back()
	if elem == nil {
		return
	}
	e := elem.Value.(*entry[K, V])
	lru.remove(e.key)
	ghost.addFront(e.key, *new(V))
	a.emit(cache.Event[K, V]{Type: cache.EventTypeEviction, Key: e.key, Value: e.value})
}

func (a *ARCCache[K, V]) adapt(favorRecency bool) {
	var delta int
	if favorRecency {
		if a.b2.len() > 0 {
			delta = a.b2.len() / max(1, a.b1.len())
		} else {
			delta = 1
		}
		a.p = min(a.p+delta, a.capacity)
	} else {
		if a.b1.len() > 0 {
			delta = a.b1.len() / max(1, a.b2.len())
		} else {
			delta = 1
		}
		a.p = max(a.p-delta, 0)
	}
}

func (a *ARCCache[K, V]) Delete(key K) error {
	for _, l := range []*cacheList[K, V]{a.t1, a.t2, a.b1, a.b2} {
		if _, ok := l.m[key]; ok {
			l.remove(key)
			return nil
		}
	}
	return common.ErrKeyNotFound
}

func (a *ARCCache[K, V]) Clear() {
	a.t1 = newCacheList[K, V](false)
	a.t2 = newCacheList[K, V](false)
	a.b1 = newCacheList[K, V](true)
	a.b2 = newCacheList[K, V](true)
	a.p = 0
}

func (a *ARCCache[K, V]) Range(fn func(K, V) bool) {
	for elem := a.t1.l.Front(); elem != nil; elem = elem.Next() {
		e := elem.Value.(*entry[K, V])
		if !fn(e.key, e.value) {
			return
		}
	}
	for elem := a.t2.l.Front(); elem != nil; elem = elem.Next() {
		e := elem.Value.(*entry[K, V])
		if !fn(e.key, e.value) {
			return
		}
	}
}

func (a *ARCCache[K, V]) OnEvent(callback func(event cache.Event[K, V])) {
	a.eventCallbacks = append(a.eventCallbacks, callback)
}

func (a *ARCCache[K, V]) emit(event cache.Event[K, V]) {
	for _, cb := range a.eventCallbacks {
		cb(event)
	}
}

func newCacheList[K comparable, V any](isGhost bool) *cacheList[K, V] {
	return &cacheList[K, V]{l: list.New(), m: make(map[K]*list.Element), isGhost: isGhost}
}

func (cl *cacheList[K, V]) addFront(key K, value V) *list.Element {
	var val any
	if cl.isGhost {
		val = &ghostEntry[K]{key}
	} else {
		val = &entry[K, V]{key, value}
	}
	elem := cl.l.PushFront(val)
	cl.m[key] = elem
	return elem
}

func (cl *cacheList[K, V]) remove(key K) {
	if elem, ok := cl.m[key]; ok {
		cl.l.Remove(elem)
		delete(cl.m, key)
	}
}

func (cl *cacheList[K, V]) removeOldest() {
	elem := cl.l.Back()
	if elem == nil {
		return
	}
	cl.l.Remove(elem)
	switch v := elem.Value.(type) {
	case *entry[K, V]:
		delete(cl.m, v.key)
	case *ghostEntry[K]:
		delete(cl.m, v.key)
	}
}

func (cl *cacheList[K, V]) moveToFront(key K, value V) {
	cl.remove(key)
	cl.addFront(key, value)
}

func (cl *cacheList[K, V]) len() int {
	return cl.l.Len()
}
