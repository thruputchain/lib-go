package liblock

import "sync"

// KeyedMutex provides a lock per key value.
type KeyedMutex[T comparable] struct {
	mu    sync.Mutex
	locks map[T]*keyLock
}

type keyLock struct {
	mu   sync.Mutex
	refs int
}

// NewKeyedMutex creates a new key-scoped mutex.
func NewKeyedMutex[T comparable]() *KeyedMutex[T] {
	return &KeyedMutex[T]{
		locks: make(map[T]*keyLock),
	}
}

func (km *KeyedMutex[T]) Lock(key T) {
	km.mu.Lock()
	lk, ok := km.locks[key]
	if !ok {
		lk = &keyLock{}
		km.locks[key] = lk
	}
	// Count holders + waiters so we can safely gc empty keys.
	lk.refs++
	km.mu.Unlock()

	lk.mu.Lock()
}

func (km *KeyedMutex[T]) Unlock(key T) {
	km.mu.Lock()
	lk, ok := km.locks[key]
	km.mu.Unlock()
	if !ok {
		panic("liblock: unlock of unknown key")
	}

	lk.mu.Unlock()

	km.mu.Lock()
	lk.refs--
	if lk.refs == 0 {
		delete(km.locks, key)
	}
	km.mu.Unlock()
}
