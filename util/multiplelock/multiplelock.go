package multiplelock

import (
	"sync"
	"sync/atomic"
)

var (
	mlInstance   sync.Once
	multipleLock MultipleLock
)

type refCounter struct {
	counter int64
	lock    *sync.RWMutex
}

// MultipleLock - is the main interface for lock base on key
type MultipleLock interface {
	// Lock base on the key
	Lock(interface{})

	// RLock lock the rw for reading
	RLock(interface{})

	// Unlock the key
	Unlock(interface{})

	// RUnlock the the read lock
	RUnlock(interface{})
}

// lock - a multi lock type
type lock struct {
	inUse sync.Map
	pool  *sync.Pool // Leverage sync.pool to improve performance
}

func (l *lock) Lock(key interface{}) {
	m := l.getLocker(key)
	atomic.AddInt64(&m.counter, 1)
	m.lock.Lock()
}

func (l *lock) RLock(key interface{}) {
	m := l.getLocker(key)
	atomic.AddInt64(&m.counter, 1)
	m.lock.RLock()
}

func (l *lock) Unlock(key interface{}) {
	m := l.getLocker(key)
	m.lock.Unlock()
	l.putBackInPool(key, m)
}

func (l *lock) RUnlock(key interface{}) {
	m := l.getLocker(key)
	m.lock.RUnlock()
	// The lock corresponding to the key is deleted
	// and released by garbage collection after the last usage.
	l.putBackInPool(key, m)
}

func (l *lock) putBackInPool(key interface{}, m *refCounter) {
	atomic.AddInt64(&m.counter, -1)
	if m.counter <= 0 {
		l.pool.Put(m.lock)
		l.inUse.Delete(key)
	}
}

func (l *lock) getLocker(key interface{}) *refCounter {
	res, _ := l.inUse.LoadOrStore(key, &refCounter{
		counter: 0,
		lock:    l.pool.Get().(*sync.RWMutex),
	})

	return res.(*refCounter)
}

// NewMultipleLock - creates a new multiple lock leveraging the singleton pattern
func NewMultipleLock() MultipleLock {
	mlInstance.Do(func() { // <-- atomic, does not allow repeating
		multipleLock = &lock{
			pool: &sync.Pool{
				New: func() interface{} {
					return &sync.RWMutex{}
				},
			},
		}
	})
	return multipleLock
}
