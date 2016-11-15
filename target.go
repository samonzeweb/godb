package godb

import (
	"fmt"
	"reflect"

	"gitlab.com/samonzeweb/godb/dbreflect"
)

type targetDescription struct {
	// Target is always a pointer
	Target            interface{}
	InstanceType      reflect.Type
	StructMapping     *dbreflect.StructMapping
	IsSlice           bool
	IsSliceOfPointers bool
}

func extractType(target interface{}) (*targetDescription, error) {
	targetDesc := targetDescription{}
	targetDesc.Target = target

	targetType := reflect.TypeOf(target)
	if targetType.Kind() != reflect.Ptr {
		return nil, fmt.Errorf("Invalid argument, need a pointer, got a %s", targetType.Kind())
	}
	targetType = targetType.Elem()

	// A target could be a slice, or a single instance
	if targetType.Kind() == reflect.Slice {
		// Slice
		targetDesc.IsSlice = true
		targetDesc.IsSliceOfPointers = false
		targetType = targetType.Elem()
		if targetType.Kind() == reflect.Ptr {
			// Slice of pointers
			targetType = targetType.Elem()
			targetDesc.IsSliceOfPointers = true
		}
	} else {
		// Single instance
		targetDesc.IsSlice = false
		targetDesc.IsSliceOfPointers = false
	}

	if targetType.Kind() != reflect.Struct {
		return nil, fmt.Errorf("Invalid argument, need a struct or structs slice, got a (or slice of) %s", targetType.Kind())
	}

	var err error
	targetDesc.InstanceType = targetType
	targetDesc.StructMapping, err = getOrCreateStructMapping(targetType)
	if err != nil {
		return nil, err
	}

	return &targetDesc, nil
}

// fillTarget build if needed new target instance and call the given function
// with the current target.
// If the target is a singel instante it just use its pointer.
// If the target is a slice, it creates new instances and expand the slice.
func (t *targetDescription) fillTarget(f func(target interface{}) error) error {
	if t.IsSlice == false {
		return f(t.Target)
	}

	// It's a slice
	// Create a new instance (reflect.Value of a pointer of the type needed)
	newInstancePointerValue := reflect.New(t.InstanceType)
	newInstancePointer := newInstancePointerValue.Interface()
	// Call func with the struct pointer
	err := f(newInstancePointer)
	if err != nil {
		return err
	}
	// Add the new instance to the struct
	// Get the current slice (t.Target is a slice pointer)
	sliceValue := reflect.ValueOf(t.Target).Elem()
	// Add the new instance (or pointer to) into the slice
	instanceOrPointerValue := newInstancePointerValue
	if !t.IsSliceOfPointers {
		instanceOrPointerValue = newInstancePointerValue.Elem()
	}
	newSliceValue := reflect.Append(sliceValue, instanceOrPointerValue)
	// Update the content of t.Target with the new slice
	reflect.ValueOf(t.Target).Elem().Set(newSliceValue)

	return nil
}
