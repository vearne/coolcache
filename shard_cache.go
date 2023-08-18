package coolcache

import (
	"container/list"
	"sync"
	"time"
)

type KVItem struct {
	key       any
	value     any
	expiredAt time.Time
}

/*
For expired keys, lazy deletion and periodic deletion are adopted to clarify
*/
type ShardCache struct {
	sync.RWMutex
	// Element.Value is a instance of KVItem
	dataMap      map[any]*list.Element
	lruList      *list.List
	cap          uint64
	CallBackFunc CallBackFunc
}

func NewShardCache(cap uint64, f CallBackFunc) *ShardCache {
	sc := ShardCache{
		dataMap:      make(map[any]*list.Element, 100),
		cap:          cap,
		CallBackFunc: f,
	}
	sc.lruList = list.New()
	go func(sc *ShardCache) {

	}(&sc)
	return &sc
}

func (sc *ShardCache) Set(key any, value any, duration time.Duration) {
	sc.Lock()
	defer sc.Unlock()
	ele, ok := sc.dataMap[key]
	if duration <= 0 {
		duration = 1000000 * time.Hour
	}
	// 1.exist
	if ok {
		// move to the tail of the bidirectional queue
		kv := ele.Value.(KVItem)
		kv.value = value
		kv.expiredAt = time.Now().Add(duration)
		sc.lruList.MoveToBack(ele)
		return
	}

	// 2. don't exist
	if len(sc.dataMap) >= int(sc.cap) {
		// remove oldest item
		front := sc.lruList.Front()
		kv := front.Value.(KVItem)
		delete(sc.dataMap, kv.key)
		sc.lruList.Remove(front)

		if sc.CallBackFunc != nil {
			go sc.CallBackFunc(key, kv.value)
		}
	}
	ele = sc.lruList.PushBack(KVItem{key: key, value: value, expiredAt: time.Now().Add(duration)})
	sc.dataMap[key] = ele
}

func (sc *ShardCache) Remove(key any) any {
	sc.RLock()
	ele, ok := sc.dataMap[key]
	sc.RUnlock()
	if ok {
		kv := ele.Value.(KVItem)
		sc.Lock()
		delete(sc.dataMap, key)
		sc.lruList.Remove(ele)
		sc.Unlock()

		if sc.CallBackFunc != nil {
			go sc.CallBackFunc(key, kv.value)
		}
		return kv.value
	}
	return nil
}

func (sc *ShardCache) Get(key any) any {
	sc.RLock()
	ele, ok := sc.dataMap[key]
	sc.RUnlock()
	if ok {
		kv := ele.Value.(KVItem)
		now := time.Now()
		// key has expired
		if now.After(kv.expiredAt) {
			sc.Lock()
			delete(sc.dataMap, key)
			sc.lruList.Remove(ele)
			sc.Unlock()
			
			if sc.CallBackFunc != nil {
				go sc.CallBackFunc(key, kv.value)
			}
			return nil
		} else {
			kv := ele.Value.(KVItem)
			return kv.value
		}
	}
	return nil
}

func (sc *ShardCache) Close() {
	sc.dataMap = nil
	sc.lruList = nil
}

func (sc *ShardCache) Size() int {
	sc.RLock()
	defer sc.RUnlock()
	return len(sc.dataMap)
}

func (sc *ShardCache) Clean() int {
	sc.RLock()
	expiredKeys := make([]any, 0)
	for key, ele := range sc.dataMap {
		kv := ele.Value.(KVItem)
		if time.Now().After(kv.expiredAt) {
			expiredKeys = append(expiredKeys, key)
		}
	}
	sc.RUnlock()
	for _, key := range expiredKeys {
		sc.Remove(key)
	}
	return len(expiredKeys)
}
