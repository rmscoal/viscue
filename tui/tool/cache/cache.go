package cache

import "sync"

type Key string

const (
	SecretKey        Key = "secret_key"
	AccountUnlockKey Key = "account_unlock_key"
	PrivateKey       Key = "private_key"
	PublicKey        Key = "public_key"

	TerminalWidth  = "terminal_width"
	TerminalHeight = "terminal_height"
)

var (
	memStore map[Key]any
	once     sync.Once
	mutex    sync.RWMutex
)

func init() {
	once.Do(func() {
		memStore = make(map[Key]any)
	})
}

// Get retrieves a value from the cache. Returns nil for not existing key.
func Get[T any](key Key) T {
	mutex.RLock()
	defer mutex.RUnlock()
	if value, exists := memStore[key]; exists {
		return value.(T)
	}
	var zero T
	return zero
}

// Set inserts key value pair to the cache. It replaces existing pair.
func Set(key Key, value any) {
	mutex.Lock()
	defer mutex.Unlock()
	memStore[key] = value
}
