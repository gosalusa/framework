package cache

import (
	"context"
	"time"

	"gosalusa.com/di"
	"gosalusa.com/extra/maps"
)

type MapCache struct {
	ttl time.Duration
	m   maps.Sync[string, MapCacheItem]
}

type MapCacheItem struct {
	Data       []byte
	Expiration time.Time
}

func (i *MapCacheItem) Expired() bool {
	return i.Expiration != time.Time{} && time.Now().After(i.Expiration)
}

func NewMapCache() *MapCache {
	return &MapCache{
		ttl: 5 * time.Minute,
		m:   maps.Sync[string, MapCacheItem]{},
	}
}

var _ MemoryCache = (*MapCache)(nil)

func Register(ctx context.Context) {
	di.RegisterLazySingleton(ctx, func() (MemoryCache, error) {
		return NewMapCache(), nil
	})
}

// Get implements [MemoryCache].
func (m *MapCache) Get(key string, defaultValue []byte) ([]byte, error) {
	i, ok := m.m.Get(key)
	if ok && i.Expired() {
		m.m.Delete(key)
		ok = false
	}
	if !ok {
		return defaultValue, nil
	}
	return i.Data, nil
}

// GetOrCreate implements [MemoryCache].
func (m *MapCache) GetOrCreate(key string, factory func() []byte) ([]byte, error) {
	i, ok := m.m.Get(key)
	if ok && i.Expired() {
		m.m.Delete(key)
		ok = false
	}
	if !ok {
		i = MapCacheItem{
			Data:       factory(),
			Expiration: time.Now().Add(m.ttl),
		}
		m.m.Set(key, i)
	}
	return i.Data, nil
}

// Remove implements [MemoryCache].
func (m *MapCache) Remove(key string) error {
	m.m.Delete(key)
	return nil
}

// Set implements [MemoryCache].
func (m *MapCache) Set(key string, value []byte, options *SetOptions) error {
	m.m.Set(key, MapCacheItem{
		Data:       value,
		Expiration: time.Now().Add(m.ttl),
	})
	return nil
}
