package godb

import (
	"reflect"
	"sync"

	"gitlab.com/samonzeweb/godb/dbreflect"
)

type structsMapping struct {
	lock           *sync.RWMutex
	structsMapping map[string]*dbreflect.StructMapping
}

// All goroutines share the same structs mapping cache.
var globalStructsMapping = structsMapping{}

// initGlobalStructsMapping is called by init()
func initGlobalStructsMapping() {
	globalStructsMapping.lock = &sync.RWMutex{}
	globalStructsMapping.structsMapping = make(map[string]*dbreflect.StructMapping)
}

//
// structMap return a StructMapping with a given type from the StructMapping
// cache. The StructMapping will be created if needed.
func getOrCreateStructMapping(structType reflect.Type) (*dbreflect.StructMapping, error) {
	globalStructsMapping.lock.RLock()
	structMapping := getStructMapping(structType)
	globalStructsMapping.lock.RUnlock()
	if structMapping != nil {
		return structMapping, nil
	}

	structMapping, err := createStructMapping(structType)
	return structMapping, err
}

// getStructMap return an existing StructMapping if it exists.
// It is not thread safe, the caller has to manage the lock !
// Dont't call it, use getOrCreateStructMap()
func getStructMapping(structType reflect.Type) *dbreflect.StructMapping {
	return globalStructsMapping.structsMapping[structType.Name()]
}

// createStrucuMap create a StructMapping an add it to the cache
// Dont't call it, use getOrCreateStructMap()
func createStructMapping(structType reflect.Type) (*dbreflect.StructMapping, error) {
	globalStructsMapping.lock.Lock()
	defer globalStructsMapping.lock.Unlock()

	// The lock was released, other goroutine could have done the job.
	structMapping := getStructMapping(structType)
	if structMapping != nil {
		return structMapping, nil
	}

	// Create the StructMapping and store it
	structMapping, err := dbreflect.NewStructMapping(structType)
	if err != nil {
		return nil, err
	}

	globalStructsMapping.structsMapping[structType.Name()] = structMapping
	return structMapping, nil
}
