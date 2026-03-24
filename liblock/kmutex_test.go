package liblock

import (
	"runtime"
	"sync"
	"sync/atomic"
	"testing"
	"time"
)

func mustPanic(t *testing.T, fn func()) {
	t.Helper()
	defer func() {
		if recover() == nil {
			t.Fatalf("expected panic, got none")
		}
	}()
	fn()
}

func refsFor[T comparable](km *KeyedMutex[T], key T) (int, bool) {
	km.mu.Lock()
	defer km.mu.Unlock()
	lk, ok := km.locks[key]
	if !ok {
		return 0, false
	}
	return lk.refs, true
}

func waitUntil(t *testing.T, timeout time.Duration, pred func() bool, msg string) {
	t.Helper()
	deadline := time.Now().Add(timeout)
	for time.Now().Before(deadline) {
		if pred() {
			return
		}
		time.Sleep(1 * time.Millisecond)
	}
	t.Fatalf("condition not met: %s", msg)
}

func TestKeyedMutex_BasicLockUnlock(t *testing.T) {
	km := NewKeyedMutex[string]()
	km.Lock("a")
	km.Unlock("a")
}

func TestKeyedMutex_UnlockUnknownKeyPanics(t *testing.T) {
	km := NewKeyedMutex[string]()
	mustPanic(t, func() {
		km.Unlock("missing")
	})
}

func TestKeyedMutex_DoubleUnlockPanics(t *testing.T) {
	km := NewKeyedMutex[string]()
	km.Lock("a")
	km.Unlock("a")
	mustPanic(t, func() {
		km.Unlock("a")
	})
}

func TestKeyedMutex_SameKeyMutualExclusion(t *testing.T) {
	km := NewKeyedMutex[string]()
	km.Lock("k")

	acquired := make(chan struct{})
	released := make(chan struct{})
	go func() {
		km.Lock("k")
		close(acquired)
		km.Unlock("k")
		close(released)
	}()

	select {
	case <-acquired:
		t.Fatalf("same key acquired while already locked")
	case <-time.After(25 * time.Millisecond):
		// expected: still blocked
	}

	km.Unlock("k")

	select {
	case <-released:
	case <-time.After(1 * time.Second):
		t.Fatalf("waiter did not acquire/release after unlock")
	}
}

func TestKeyedMutex_DifferentKeysDoNotBlock(t *testing.T) {
	km := NewKeyedMutex[string]()
	km.Lock("a")

	acquired := make(chan struct{})
	go func() {
		km.Lock("b")
		close(acquired)
		km.Unlock("b")
	}()

	select {
	case <-acquired:
		// expected
	case <-time.After(200 * time.Millisecond):
		t.Fatalf("different key unexpectedly blocked")
	}

	km.Unlock("a")
}

func TestKeyedMutex_RefcountTracksWaitersAndKeyIsCollected(t *testing.T) {
	km := NewKeyedMutex[string]()
	km.Lock("k")

	done := make(chan struct{})
	go func() {
		km.Lock("k")
		km.Unlock("k")
		close(done)
	}()

	waitUntil(t, 1*time.Second, func() bool {
		refs, ok := refsFor(km, "k")
		return ok && refs == 2
	}, "refs should include holder + waiter")

	km.Unlock("k")

	select {
	case <-done:
	case <-time.After(1 * time.Second):
		t.Fatalf("waiter did not finish")
	}

	waitUntil(t, 1*time.Second, func() bool {
		_, ok := refsFor(km, "k")
		return !ok
	}, "key should be removed after last unlock")
}

func TestKeyedMutex_ReuseKeyAfterCollection(t *testing.T) {
	km := NewKeyedMutex[int]()

	for i := 0; i < 3; i++ {
		km.Lock(42)
		km.Unlock(42)
		if _, ok := refsFor(km, 42); ok {
			t.Fatalf("key should be removed after cycle %d", i)
		}
	}
}

func TestKeyedMutex_ConcurrentStressSingleKey(t *testing.T) {
	km := NewKeyedMutex[int]()
	const workers = 32
	const iterations = 200

	var inCritical int32
	var violations int32
	var wg sync.WaitGroup

	wg.Add(workers)
	for i := 0; i < workers; i++ {
		go func() {
			defer wg.Done()
			for j := 0; j < iterations; j++ {
				km.Lock(7)
				if atomic.AddInt32(&inCritical, 1) != 1 {
					atomic.AddInt32(&violations, 1)
				}
				runtime.Gosched()
				atomic.AddInt32(&inCritical, -1)
				km.Unlock(7)
			}
		}()
	}

	wg.Wait()

	if got := atomic.LoadInt32(&violations); got != 0 {
		t.Fatalf("mutual exclusion violated %d times", got)
	}
	if _, ok := refsFor(km, 7); ok {
		t.Fatalf("key should be removed after stress test")
	}
}

func TestKeyedMutex_StructKey(t *testing.T) {
	type key struct {
		A int
		B string
	}

	km := NewKeyedMutex[key]()
	k := key{A: 1, B: "x"}
	km.Lock(k)
	km.Unlock(k)
	if _, ok := refsFor(km, k); ok {
		t.Fatalf("struct key should be removed after unlock")
	}
}
