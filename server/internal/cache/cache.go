package LRUCache

import (
	"container/list"
	"sync"
)

type LRUCache struct {
	mutex sync.Mutex
	m     map[int64]*cacheMap
	cap   int
	l     list.List
}

type cacheMap struct {
	elem  *list.Element
	value string
}

func NewLRUCache(cap int) *LRUCache {
	return &LRUCache{
		m:   map[int64]*cacheMap{},
		cap: cap,
		l:   list.List{},
	}
}

func (ch *LRUCache) Get(key int64) (value string, ok bool) {
	ch.mutex.Lock()
	defer ch.mutex.Unlock()
	val, ok := ch.m[key]
	if !ok {
		return "", false
	}
	ch.l.MoveToFront(val.elem)
	return val.value, true
}

func (ch *LRUCache) Add(key int64, value string) {
	ch.mutex.Lock()
	defer ch.mutex.Unlock()
	if v, ok := ch.m[key]; !ok {
		el := ch.l.PushFront(key)
		ch.m[key] = &cacheMap{
			elem:  el,
			value: value,
		}
		if ch.l.Len() > ch.cap {
			backEl := ch.l.Back()
			backElKey := backEl.Value.(int64)
			ch.l.Remove(backEl)
			delete(ch.m, backElKey)
		}
	} else {
		v.value = value
		ch.l.MoveToFront(v.elem)
	}
}

func (ch *LRUCache) Remove(key int64) (ok bool) {
	if val, ok := ch.m[key]; ok {
		ch.l.Remove(val.elem)
		delete(ch.m, key)
		return true
	}
	return false
}
