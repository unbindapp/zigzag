package cache

import (
	"context"
	"errors"
	"sync"
	"time"
)

var ErrMiss = errors.New("cache miss")

type entry[T any] struct {
	value     T
	expiresAt time.Time
}

type MemoryCache[T any] struct {
	mu      sync.Mutex
	entries map[string]entry[T]
	now     func() time.Time
}

func NewMemoryCache[T any]() *MemoryCache[T] {
	return &MemoryCache[T]{entries: map[string]entry[T]{}, now: time.Now}
}

func (self *MemoryCache[T]) Get(_ context.Context, key string) (T, error) {
	self.mu.Lock()
	defer self.mu.Unlock()
	found, ok := self.entries[key]
	if !ok || !self.now().Before(found.expiresAt) {
		delete(self.entries, key)
		var zero T
		return zero, ErrMiss
	}
	return found.value, nil
}

func (self *MemoryCache[T]) SetWithExpiration(_ context.Context, key string, value T, ttl time.Duration) error {
	self.mu.Lock()
	defer self.mu.Unlock()
	now := self.now()
	for k, e := range self.entries {
		if !now.Before(e.expiresAt) {
			delete(self.entries, k)
		}
	}
	self.entries[key] = entry[T]{value: value, expiresAt: now.Add(ttl)}
	return nil
}
