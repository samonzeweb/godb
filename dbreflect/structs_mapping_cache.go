package dbreflect

import (
	"reflect"
	"sync"
)

type StructsMappingCache struct {
	lock           *sync.RWMutex
	structsMapping map[string]*StructMapping
}

// Cache is a global cache of StructMapping (StructsMappingCache).
// It's thread safe.
var Cache *StructsMappingCache

func init() {
	Cache = NewStructsMappingCache()
}

// NewStructsMappingCache builds a new StructsMappingCache.
func NewStructsMappingCache() *StructsMappingCache {
	smc := &StructsMappingCache{}
	smc.lock = &sync.RWMutex{}
	smc.structsMapping = make(map[string]*StructMapping)
	return smc
}

// GetOrCreateStructMapping returns a StructMapping with a given type from
// the StructMapping cache. The StructMapping will be created if needed.
func (smc *StructsMappingCache) GetOrCreateStructMapping(structType reflect.Type) (*StructMapping, error) {
	smc.lock.RLock()
	structMapping := smc.getStructMapping(structType)
	smc.lock.RUnlock()
	if structMapping != nil {
		return structMapping, nil
	}

	structMapping, err := smc.createStructMapping(structType)
	return structMapping, err
}

// getStructMapping return an existing StructMapping if it exists.
// It is not thread safe, the caller has to manage the lock !
// Dont't call it, use getOrCreateStructMap()
func (smc *StructsMappingCache) getStructMapping(structType reflect.Type) *StructMapping {
	return smc.structsMapping[structType.Name()]
}

// createStructMapping create a StructMapping an add it to the cache
// Dont't call it, use getOrCreateStructMap()
func (smc *StructsMappingCache) createStructMapping(structType reflect.Type) (*StructMapping, error) {
	smc.lock.Lock()
	defer smc.lock.Unlock()

	// The lock was released, other goroutine could have done the job.
	structMapping := smc.getStructMapping(structType)
	if structMapping != nil {
		return structMapping, nil
	}

	// Create the StructMapping and store it
	structMapping, err := NewStructMapping(structType)
	if err != nil {
		return nil, err
	}

	smc.structsMapping[structType.Name()] = structMapping
	return structMapping, nil
}
