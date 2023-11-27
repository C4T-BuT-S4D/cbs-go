package keylock

import (
	"sort"
	"sync"
)

type KeyLock struct {
	mu    sync.RWMutex
	locks map[string]chan struct{}
}

func New() *KeyLock {
	return &KeyLock{locks: make(map[string]chan struct{})}
}

func (l *KeyLock) LockKeys(keys []string, cancel <-chan struct{}) (canceled bool, unlock func()) {
	var keysCopy []string

	if len(keys) > 1 {
		keysCopy = make([]string, len(keys))
		copy(keysCopy, keys)
		sort.Strings(keysCopy)
	} else {
		keysCopy = keys
	}

	locked := make([]string, 0, len(keys))

	unlock = func() {
		l.mu.RLock()
		for _, k := range locked {
			select {
			case l.locks[k] <- struct{}{}:
			default:
			}
		}
		l.mu.RUnlock()
	}

	for _, key := range keysCopy {
		l.mu.Lock()
		kmu, ok := l.locks[key]
		if !ok {
			kmu = make(chan struct{}, 1)
			kmu <- struct{}{}
			l.locks[key] = kmu
		}
		l.mu.Unlock()

		select {
		case <-kmu:
			locked = append(locked, key)
		case <-cancel:
			unlock()
			return true, nil
		}
	}

	return false, unlock
}
