package keylock

import (
	"runtime"
	"sync"
	"sync/atomic"
)

// KeyLock provides a lock mechanism associated with a string key.
// Internally, it uses spinlocks and sync.Map to manage locks per key.
type KeyLock struct {
	locks    sync.Map // map[string]*spinLock
	count    int32    // number of currently held locks
	maxSpins int32    // max backoff spins while waiting for lock
}

// spinLock represents a lightweight spin-based lock using atomic operations.
type spinLock int32

const (
	defaultMaxSpins = 16 // Default maximum number of spin attempts
)

// New creates a new KeyLock instance with default configuration.
func New() *KeyLock {
	return &KeyLock{
		maxSpins: defaultMaxSpins,
	}
}

// WithMaxSpins sets the maximum backoff spins before yielding the CPU.
func (kl *KeyLock) WithMaxSpins(max int32) *KeyLock {
	kl.maxSpins = max
	return kl
}

// TryLock attempts to acquire the lock for the given key immediately.
// Returns true if the lock was successfully acquired, false otherwise.
func (kl *KeyLock) TryLock(key string) bool {
	value, _ := kl.locks.LoadOrStore(key, new(spinLock))
	lock := value.(*spinLock)
	if atomic.CompareAndSwapInt32((*int32)(lock), 0, 1) {
		atomic.AddInt32(&kl.count, 1)
		return true
	}
	return false
}

// Lock acquires the lock for the given key, using exponential backoff if necessary.
func (kl *KeyLock) Lock(key string) {
	value, _ := kl.locks.LoadOrStore(key, new(spinLock))
	lock := value.(*spinLock)

	backoff := 1
	for !atomic.CompareAndSwapInt32((*int32)(lock), 0, 1) {
		for range backoff {
			runtime.Gosched() // Yield processor
		}
		if backoff < int(kl.maxSpins) {
			backoff <<= 1 // Exponential backoff
		}
	}
	atomic.AddInt32(&kl.count, 1)
}

// Unlock releases the lock associated with the given key.
func (kl *KeyLock) Unlock(key string) {
	if value, ok := kl.locks.Load(key); ok {
		lock := value.(*spinLock)
		atomic.StoreInt32((*int32)(lock), 0)
		atomic.AddInt32(&kl.count, -1)
	}
}

// Size returns the number of currently active (held) locks.
func (kl *KeyLock) Size() int {
	return int(atomic.LoadInt32(&kl.count))
}

// Cleanup removes all unused (unlocked) locks from the map.
// This is optional and can be used to free memory.
func (kl *KeyLock) Cleanup() {
	kl.locks.Range(func(key, value any) bool {
		if atomic.LoadInt32((*int32)(value.(*spinLock))) == 0 {
			kl.locks.Delete(key)
		}
		return true
	})
}
