# KeyLock

A simple, high-performance key-based locking mechanism for Go.

## Features

- Lock by string key for fine-grained synchronization
- High performance with spinlocks and atomic operations

## Installation

```bash
go get github.com/werbenhu/keylock
```

## Usage

### Basic Example

```go
package main

import "github.com/werbenhu/keylock"

func main() {
    kl := keylock.New()
    
    // Lock a resource by key
    kl.Lock("user:123")
    // ... critical section ...
    kl.Unlock("user:123")
}
```

### Try Lock (Non-blocking)

```go
if kl.TryLock("resource") {
    defer kl.Unlock("resource")
    // ... do work ...
} else {
    // Lock is busy, handle accordingly
}
```

## API

| Method | Description |
|--------|-------------|
| `New()` | Create a new KeyLock |
| `Lock(key)` | Acquire lock for key (blocks) |
| `TryLock(key)` | Try to acquire lock (returns bool) |
| `Unlock(key)` | Release lock for key |
| `Size()` | Number of active locks |
| `Cleanup()` | Remove unused locks from memory |
| `WithMaxSpins(n)` | Set max spin attempts |

## Configuration

```go
// Default configuration
kl := keylock.New()

// Custom spin limit
kl := keylock.New().WithMaxSpins(32)
```

## Performance

KeyLock uses spinlocks with exponential backoff for optimal performance in high-concurrency scenarios while minimizing CPU usage during contention.