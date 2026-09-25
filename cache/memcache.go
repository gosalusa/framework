package cache

type MemoryCache interface {
	Get(key string, defaultValue []byte) ([]byte, error)
	GetOrCreate(key string, factory func() []byte) ([]byte, error)
	Set(key string, value []byte, options *SetOptions) error
	Remove(key string) error
}

type SetOptions struct{}
