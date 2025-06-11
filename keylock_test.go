package keylock

import (
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestKeyLockCreation(t *testing.T) {
	kl := New()
	assert.NotNil(t, kl, "KeyLock should be created")
	assert.Equal(t, 0, kl.Size(), "New KeyLock should have zero locks")
}

func TestKeyLockBasicLocking(t *testing.T) {
	kl := New()

	kl.Lock("test1")
	assert.Equal(t, 1, kl.Size(), "Count should be 1 after locking a key")

	kl.Unlock("test1")
	assert.Equal(t, 0, kl.Size(), "Count should be 0 after unlocking a key")
}

func TestKeyLockMultipleKeys(t *testing.T) {
	kl := New()

	kl.Lock("test1")
	kl.Lock("test2")
	kl.Lock("test3")
	assert.Equal(t, 3, kl.Size(), "Count should be 3 after locking 3 keys")

	kl.Unlock("test1")
	assert.Equal(t, 2, kl.Size(), "Count should be 2 after unlocking 1 key")
	kl.Unlock("test2")
	kl.Unlock("test3")
	assert.Equal(t, 0, kl.Size(), "Count should be 0 after unlocking all keys")
}

func TestKeyLockReferenceCount(t *testing.T) {
	kl := New()

	kl.Lock("test1")
	assert.Equal(t, 1, kl.Size(), "Count should be 1 after first lock")

	var secondLockAcquired bool
	go func() {
		kl.Lock("test1")
		secondLockAcquired = true
		kl.Unlock("test1")
	}()

	time.Sleep(100 * time.Millisecond)

	assert.False(t, secondLockAcquired, "Second lock should not be acquired while first lock is held")
	kl.Unlock("test1")

	time.Sleep(100 * time.Millisecond)
	assert.True(t, secondLockAcquired, "Second lock should be acquired after first lock is released")
	assert.Equal(t, 0, kl.Size(), "Count should be 0 after all locks are released")
}

func TestKeyLockUnlockNonExistentKey(t *testing.T) {
	kl := New()
	assert.NotPanics(t, func() {
		kl.Unlock("nonexistent")
	}, "Unlocking a non-existent key should not panic")

	assert.Equal(t, 0, kl.Size(), "Count should remain 0")
}

func TestKeyLockConcurrentAccess(t *testing.T) {
	kl := New()
	const numGoroutines = 100
	const numOperationsPerGoroutine = 10

	var wg sync.WaitGroup
	wg.Add(numGoroutines)

	for i := range numGoroutines {
		go func(id int) {
			defer wg.Done()
			key := "key" + string(rune('A'+id%26))

			for j := 0; j < numOperationsPerGoroutine; j++ {
				kl.Lock(key)
				time.Sleep(1 * time.Millisecond)
				kl.Unlock(key)
			}
		}(i)
	}

	wg.Wait()
	assert.Equal(t, 0, kl.Size(), "Count should be 0 after all goroutines are done")
}

func TestKeyLockDifferentKeysConcurrently(t *testing.T) {
	kl := New()
	const numKeys = 10

	var started sync.WaitGroup
	var complete sync.WaitGroup
	started.Add(numKeys)
	complete.Add(numKeys)

	completedCount := 0
	var countMutex sync.Mutex

	for i := 0; i < numKeys; i++ {
		go func(id int) {
			key := "concurrent-key-" + string(rune('A'+id))

			kl.Lock(key)
			started.Done()
			started.Wait()
			time.Sleep(10 * time.Millisecond)

			kl.Unlock(key)

			countMutex.Lock()
			completedCount++
			countMutex.Unlock()

			complete.Done()
		}(i)
	}

	complete.Wait()
	assert.Equal(t, numKeys, completedCount, "All goroutines should complete")
	assert.Equal(t, 0, kl.Size(), "All locks should be released")
}

func TestKeyLockNestedLocks(t *testing.T) {
	kl := New()

	kl.Lock("outer")
	kl.Lock("inner")

	assert.Equal(t, 2, kl.Size(), "Should have 2 locks")

	kl.Unlock("inner")
	assert.Equal(t, 1, kl.Size(), "Should have 1 lock after unlocking inner")

	kl.Unlock("outer")
	assert.Equal(t, 0, kl.Size(), "Should have 0 locks after unlocking outer")
}
