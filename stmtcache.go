package godb

import (
	"database/sql"
	"fmt"
)

// StmtCache is a LRU cache for prepared statements.
type StmtCache struct {
	// isEnabled isn't useful for the cache itself, but it allow a complete
	// control of the cache with a StmtCache instance for godb users.
	isEnabled bool
	maxSize   int
	lastUse   uint64
	content   map[string]*stmtCacheItem
}

type stmtCacheItem struct {
	stmt    *sql.Stmt
	lastUse uint64
}

// DefaultStmtCacheSize is the default size of prepared statements LRU cache.
var DefaultStmtCacheSize = 64

// MaximumStmtCacheSize is the maximum allowed size for the prepared
// statements cache. The value is arbitrary, but big enough for most uses.
var MaximumStmtCacheSize = 4096

// newStmtCache builds and returns a new prepared statements cache.
// The returned cache is enabled by default.
func newStmtCache() *StmtCache {
	return &StmtCache{
		isEnabled: true,
		maxSize:   DefaultStmtCacheSize,
		content:   make(map[string]*stmtCacheItem),
	}
}

// Enable enables the cache.
func (cache *StmtCache) Enable() {
	cache.isEnabled = true
}

// Disable disables the cache.
// Disabling the cache does not clear it.
func (cache *StmtCache) Disable() {
	cache.isEnabled = false
}

// IsEnabled returns true if the cache is enabled, false otherwise.
func (cache *StmtCache) IsEnabled() bool {
	return cache.isEnabled
}

// SetSize sets the maximum cache size, and remove exceeding entries if needed.
func (cache *StmtCache) SetSize(size int) error {
	if size < 0 || size > MaximumStmtCacheSize {
		return fmt.Errorf("given cache size is out if allowed boundarie")
	}

	cache.maxSize = size

	for len(cache.content) > cache.maxSize {
		err := cache.removeLeastRecentlyUsed()
		if err != nil {
			return err
		}
	}

	return nil
}

// GetSize returns the maximum cache size
func (cache *StmtCache) GetSize() int {
	return cache.maxSize
}

// sdd adds a prepared statement into the cache.
func (cache *StmtCache) add(query string, stmt *sql.Stmt) error {
	// Cache full ?
	if len(cache.content) >= cache.maxSize {
		err := cache.removeLeastRecentlyUsed()
		if err != nil {
			return err
		}
	}

	// Add new entry
	cache.lastUse++
	cache.content[query] = &stmtCacheItem{
		stmt:    stmt,
		lastUse: cache.lastUse,
	}

	return nil
}

// removeLeastRecentlyUsed removes the least recently used entry from the cache.
func (cache *StmtCache) removeLeastRecentlyUsed() error {
	// Search the LRU item
	minQuery := ""
	var minUsageCount uint64 = 18446744073709551615

	for query, cacheItem := range cache.content {
		if cacheItem.lastUse < minUsageCount {
			minUsageCount = cacheItem.lastUse
			minQuery = query
		}
	}

	// Remove the item
	if minQuery > "" {
		cacheItem, ok := cache.content[minQuery]
		if !ok {
			return fmt.Errorf("removeLeastRecentlyUsed : query not found in cache")
		}
		delete(cache.content, minQuery)
		if err := cacheItem.stmt.Close(); err != nil {
			return err
		}
	}

	return nil
}

// get returns an existing prepared statement or nil.
func (cache *StmtCache) get(query string) *sql.Stmt {
	cacheItem, ok := cache.content[query]
	if !ok {
		return nil
	}

	cache.lastUse++
	cacheItem.lastUse = cache.lastUse
	return cacheItem.stmt
}

// Clear closes properly the the cached stmt, and clear the cache.
func (cache *StmtCache) Clear() error {
	defer func() {
		cache.clearWithoutClosingStmt()
	}()

	for _, cacheItem := range cache.content {
		if err := cacheItem.stmt.Close(); err != nil {
			return err
		}
	}

	return nil
}

// clearWithoutClosingStmt clears the cache but does not close the prepared
// statements.
//
// Use case if for Tx Commit or Rollback, when prepared statements could no
// longer be used, do not use it otherwise.
func (cache *StmtCache) clearWithoutClosingStmt() {
	cache.content = make(map[string]*stmtCacheItem)
}
