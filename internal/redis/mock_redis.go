package redis

import (
	"errors"
	"log"
	"sync"
	"time"
)

type CacheMock struct {
	store map[string]cacheItem
	mu    sync.RWMutex
}

type cacheItem struct {
	value      string
	expiration time.Time
}

// NewCacheMock создает новый экземпляр мока Redis.
func NewCacheMock() *CacheMock {
	return &CacheMock{
		store: make(map[string]cacheItem),
	}
}

// Get возвращает значение из мока Redis.
func (c *CacheMock) Get(key string) (string, error) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	item, exists := c.store[key]
	if !exists || time.Now().After(item.expiration) {
		return "", nil // Вернуть пустую строку, если ключ отсутствует или истек срок действия
	}

	return item.value, nil
}

// Set сохраняет значение в мок Redis.
func (c *CacheMock) Set(key string, value string, ttl time.Duration) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.store[key] = cacheItem{
		value:      value,
		expiration: time.Now().Add(ttl),
	}

	return nil
}

// Close завершает работу мока Redis и очищает его состояние.
func (c *CacheMock) Close() error {
	c.mu.Lock()
	defer c.mu.Unlock()

	if c.store == nil {
		log.Println("CacheMock already closed or uninitialized")
		return errors.New("cache already closed")
	}

	// Очищаем данные
	c.store = nil
	log.Println("CacheMock successfully closed")
	return nil
}
