package v1

import "sync"

type RWLocker interface {
	Lock()
	Unlock()
	RLock()
	RUnlock()
}

type MutexStub struct{}

func (m *MutexStub) Lock()    {}
func (m *MutexStub) Unlock()  {}
func (m *MutexStub) RLock()   {}
func (m *MutexStub) RUnlock() {}

type Mutex struct {
	mu sync.Mutex
}

func (m *Mutex) Lock() {
	m.mu.Lock()
}

func (m *Mutex) Unlock() {
	m.mu.Unlock()
}

func (m *Mutex) RLock() {
	m.mu.Lock()
}

func (m *Mutex) RUnlock() {
	m.mu.Unlock()
}
